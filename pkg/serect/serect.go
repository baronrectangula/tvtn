// Package serect wraps values that must be hidden from logs and formatting.
package serect

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strconv"
)

type Raw any

// Value hides its value during formatting and marshaling.
// Typed accessors and Expose explicitly expose the value to callers.
//
// Marshaling a Value struct field with a tag such as `json:"password"`
// writes "[REDACTED]", not the secret. The tag names the field; it does not
// bypass redaction. Unmarshaling
// accepts secret values, but a marshal/unmarshal round trip loses the original
// value: decoding "[REDACTED]" stores that literal string. When an API, database,
// or other consumer needs the actual secret, expose it and pass the exposed value
// explicitly. For example:
//
//	password := serect.New("my-password")
//	payload := struct {
//		Password any `json:"password"`
//	}{Password: password.Expose()}
//	data, err := json.Marshal(payload)
//	if err != nil {
//		return err
//	}
//	// Pass data to the intended consumer; it contains the actual password.
//	// Do not log data or the exposed value.
type Value struct{ inner Raw }

const mask = "[REDACTED]"

func New(value Raw) Value                   { return Value{inner: value} }
func (s Value) Expose() Raw                 { return s.inner }
func (Value) String() string                { return mask }
func (Value) GoString() string              { return mask }
func (Value) Format(w fmt.State, verb rune) { _, _ = io.WriteString(w, mask) }
func (Value) LogValue() slog.Value          { return slog.StringValue(mask) }
func (Value) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(mask)
	if err != nil {
		return nil, fmt.Errorf("serect.MarshalJSON: %w", err)
	}
	return data, nil
}
func (Value) MarshalText() ([]byte, error) { return []byte(mask), nil }

func (s *Value) UnmarshalJSON(data []byte) error {
	var value any
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		return fmt.Errorf("serect.UnmarshalJSON: %w", err)
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		if err == nil {
			err = errors.New("unexpected trailing JSON value")
		}
		return fmt.Errorf("serect.UnmarshalJSON: %w", err)
	}
	s.inner = value
	return nil
}

func (s *Value) UnmarshalText(data []byte) error {
	s.inner = string(data)
	return nil
}

// Scan implements sql.Scanner. Driver-owned byte slices are copied.
func (s *Value) Scan(value any) error {
	if data, ok := value.([]byte); ok {
		value = bytes.Clone(data)
	}
	s.inner = value
	return nil
}

// ExposeAsString accepts primitive values, JSON numbers, and SQL byte strings.
func (s Value) ExposeAsString() (string, error) {
	value, err := s.asString()
	if err != nil {
		return "", fmt.Errorf("serect.ExposeAsString: %w", err)
	}
	return value, nil
}

func (s Value) asString() (string, error) {
	switch v := s.inner.(type) {
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	case json.Number:
		return string(v), nil
	case bool, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr, float32, float64, complex64, complex128:
		return fmt.Sprint(v), nil
	default:
		return "", errors.New("unsupported value type")
	}
}

func (s Value) ExposeAsBool() (bool, error) {
	text, err := s.asString()
	if err != nil {
		return false, fmt.Errorf("serect.ExposeAsBool: %w", err)
	}
	value, err := strconv.ParseBool(text)
	if err != nil {
		return false, fmt.Errorf("serect.ExposeAsBool: %w", err)
	}
	return value, nil
}

func (s Value) ExposeAsInt() (int, error) {
	text, err := s.asString()
	if err != nil {
		return 0, fmt.Errorf("serect.ExposeAsInt: %w", err)
	}
	value, err := strconv.ParseInt(text, 10, 0)
	if err != nil {
		return 0, fmt.Errorf("serect.ExposeAsInt: %w", err)
	}
	return int(value), nil
}

func (s Value) ExposeAsInt8() (int8, error) {
	text, err := s.asString()
	if err != nil {
		return 0, fmt.Errorf("serect.ExposeAsInt8: %w", err)
	}
	value, err := strconv.ParseInt(text, 10, 8)
	if err != nil {
		return 0, fmt.Errorf("serect.ExposeAsInt8: %w", err)
	}
	return int8(value), nil
}

func (s Value) ExposeAsInt16() (int16, error) {
	text, err := s.asString()
	if err != nil {
		return 0, fmt.Errorf("serect.ExposeAsInt16: %w", err)
	}
	value, err := strconv.ParseInt(text, 10, 16)
	if err != nil {
		return 0, fmt.Errorf("serect.ExposeAsInt16: %w", err)
	}
	return int16(value), nil
}

func (s Value) ExposeAsInt32() (int32, error) {
	text, err := s.asString()
	if err != nil {
		return 0, fmt.Errorf("serect.ExposeAsInt32: %w", err)
	}
	value, err := strconv.ParseInt(text, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("serect.ExposeAsInt32: %w", err)
	}
	return int32(value), nil
}

func (s Value) ExposeAsInt64() (int64, error) {
	text, err := s.asString()
	if err != nil {
		return 0, fmt.Errorf("serect.ExposeAsInt64: %w", err)
	}
	value, err := strconv.ParseInt(text, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("serect.ExposeAsInt64: %w", err)
	}
	return int64(value), nil
}

func (s Value) ExposeAsUint() (uint, error) {
	text, err := s.asString()
	if err != nil {
		return 0, fmt.Errorf("serect.ExposeAsUint: %w", err)
	}
	value, err := strconv.ParseUint(text, 10, 0)
	if err != nil {
		return 0, fmt.Errorf("serect.ExposeAsUint: %w", err)
	}
	return uint(value), nil
}

func (s Value) ExposeAsUint8() (uint8, error) {
	text, err := s.asString()
	if err != nil {
		return 0, fmt.Errorf("serect.ExposeAsUint8: %w", err)
	}
	value, err := strconv.ParseUint(text, 10, 8)
	if err != nil {
		return 0, fmt.Errorf("serect.ExposeAsUint8: %w", err)
	}
	return uint8(value), nil
}

func (s Value) ExposeAsUint16() (uint16, error) {
	text, err := s.asString()
	if err != nil {
		return 0, fmt.Errorf("serect.ExposeAsUint16: %w", err)
	}
	value, err := strconv.ParseUint(text, 10, 16)
	if err != nil {
		return 0, fmt.Errorf("serect.ExposeAsUint16: %w", err)
	}
	return uint16(value), nil
}

func (s Value) ExposeAsUint32() (uint32, error) {
	text, err := s.asString()
	if err != nil {
		return 0, fmt.Errorf("serect.ExposeAsUint32: %w", err)
	}
	value, err := strconv.ParseUint(text, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("serect.ExposeAsUint32: %w", err)
	}
	return uint32(value), nil
}

func (s Value) ExposeAsUint64() (uint64, error) {
	text, err := s.asString()
	if err != nil {
		return 0, fmt.Errorf("serect.ExposeAsUint64: %w", err)
	}
	value, err := strconv.ParseUint(text, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("serect.ExposeAsUint64: %w", err)
	}
	return uint64(value), nil
}

func (s Value) ExposeAsUintptr() (uintptr, error) {
	text, err := s.asString()
	if err != nil {
		return 0, fmt.Errorf("serect.ExposeAsUintptr: %w", err)
	}
	value, err := strconv.ParseUint(text, 10, strconv.IntSize)
	if err != nil {
		return 0, fmt.Errorf("serect.ExposeAsUintptr: %w", err)
	}
	return uintptr(value), nil
}

func (s Value) ExposeAsFloat32() (float32, error) {
	text, err := s.asString()
	if err != nil {
		return 0, fmt.Errorf("serect.ExposeAsFloat32: %w", err)
	}
	value, err := strconv.ParseFloat(text, 32)
	if err != nil {
		return 0, fmt.Errorf("serect.ExposeAsFloat32: %w", err)
	}
	return float32(value), nil
}

func (s Value) ExposeAsFloat64() (float64, error) {
	text, err := s.asString()
	if err != nil {
		return 0, fmt.Errorf("serect.ExposeAsFloat64: %w", err)
	}
	value, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return 0, fmt.Errorf("serect.ExposeAsFloat64: %w", err)
	}
	return float64(value), nil
}

func (s Value) ExposeAsComplex64() (complex64, error) {
	text, err := s.asString()
	if err != nil {
		return 0, fmt.Errorf("serect.ExposeAsComplex64: %w", err)
	}
	value, err := strconv.ParseComplex(text, 64)
	if err != nil {
		return 0, fmt.Errorf("serect.ExposeAsComplex64: %w", err)
	}
	return complex64(value), nil
}

func (s Value) ExposeAsComplex128() (complex128, error) {
	text, err := s.asString()
	if err != nil {
		return 0, fmt.Errorf("serect.ExposeAsComplex128: %w", err)
	}
	value, err := strconv.ParseComplex(text, 128)
	if err != nil {
		return 0, fmt.Errorf("serect.ExposeAsComplex128: %w", err)
	}
	return complex128(value), nil
}

func (s Value) ExposeAsByte() (byte, error) { return s.ExposeAsUint8() }
func (s Value) ExposeAsRune() (rune, error) { return s.ExposeAsInt32() }
