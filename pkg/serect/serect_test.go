package serect

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"
)

func TestRedaction(t *testing.T) {
	s := New("private-value")
	for _, format := range []string{"%v", "%+v", "%#v", "%s", "%q", "%x"} {
		if got := fmt.Sprintf(format, s); got != mask {
			t.Fatalf("format %s: %s", format, got)
		}
	}
	data, err := json.Marshal(s)
	if err != nil || string(data) != `"[REDACTED]"` {
		t.Fatal("JSON did not redact")
	}
	if s.Expose() != "private-value" {
		t.Fatal("expose mismatch")
	}
	_, err = s.ExposeAsInt64()
	if err == nil {
		t.Fatal("expected conversion error")
	}
}

func TestJSONTaggedFieldRoundTrip(t *testing.T) {
	type credentials struct {
		Password Value `json:"password"`
	}
	var original credentials
	if err := json.Unmarshal([]byte(`{"password":"my-password"}`), &original); err != nil {
		t.Fatal(err)
	}
	if original.Password.Expose() != "my-password" {
		t.Fatal("unmarshaling should accept the secret")
	}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"password":"[REDACTED]"}` {
		t.Fatal("tagged field should marshal as redacted")
	}
	var restored credentials
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatal(err)
	}
	if restored.Password.Expose() != "[REDACTED]" {
		t.Fatal("round trip should store the literal redaction marker")
	}

	payload := struct {
		Password any `json:"password"`
	}{Password: original.Password.Expose()}
	data, err = json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"password":"my-password"}` {
		t.Fatal("explicitly exposed value should marshal as the actual secret")
	}
}

func TestDecodeAndConvert(t *testing.T) {
	var s Value
	if err := json.Unmarshal([]byte(`9223372036854775807`), &s); err != nil {
		t.Fatal(err)
	}
	if got, err := s.ExposeAsInt64(); err != nil || got != 9223372036854775807 {
		t.Fatal("lost precision")
	}
	if _, err := s.ExposeAsInt8(); err == nil {
		t.Fatal("expected overflow")
	}
	if _, err := New("-1").ExposeAsUint64(); err == nil {
		t.Fatal("expected unsigned failure")
	}
	if _, err := New("1.5").ExposeAsInt64(); err == nil {
		t.Fatal("expected fractional failure")
	}
	if got, err := New("true").ExposeAsBool(); err != nil || !got {
		t.Fatal("bool conversion")
	}
	if got, err := New("1+2i").ExposeAsComplex128(); err != nil || got != 1+2i {
		t.Fatal("complex conversion")
	}
	data := []byte("42")
	if err := s.Scan(data); err != nil {
		t.Fatal(err)
	}
	data[0] = '0'
	if got, err := s.ExposeAsInt64(); err != nil || got != 42 {
		t.Fatal("SQL bytes were not copied")
	}
	if err := s.UnmarshalJSON([]byte(`"private-value" trailing`)); err == nil {
		t.Fatal("expected JSON error")
	}
}

func TestBoundaryErrorsPreserveCause(t *testing.T) {
	_, err := New("private-value").ExposeAsInt64()
	if !errors.Is(err, strconv.ErrSyntax) {
		t.Fatal("lost strconv cause")
	}
	var numberError *strconv.NumError
	if !errors.As(err, &numberError) {
		t.Fatal("lost strconv error type")
	}
	if !strings.Contains(err.Error(), "serect.ExposeAsInt64: ") {
		t.Fatal("missing boundary context")
	}
	var s Value
	err = s.UnmarshalJSON([]byte(`{"private-value":}`))
	var syntaxError *json.SyntaxError
	if !errors.As(err, &syntaxError) {
		t.Fatal("lost JSON cause")
	}
	if err := s.UnmarshalJSON([]byte(`1 2`)); err == nil {
		t.Fatal("accepted trailing value")
	}
	_, err = New(struct{}{}).ExposeAsInt64()
	if !strings.Contains(err.Error(), "ExposeAsInt64") || strings.Contains(err.Error(), "ExposeAsString") {
		t.Fatal("incorrect internal boundary")
	}
}
