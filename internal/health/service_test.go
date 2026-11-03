package health

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

type mockPinger struct {
	err      error
	calls    int
	contexts []context.Context
}

func (m *mockPinger) Ping(ctx context.Context) error {
	m.calls++
	m.contexts = append(m.contexts, ctx)
	return m.err
}

func TestServiceReadyPingsAll(t *testing.T) {
	first, second := errors.New("first failure"), errors.New("second failure")
	for _, tc := range []struct {
		name     string
		failures []error
	}{
		{"all healthy", []error{nil, nil, nil}},
		{"first fails", []error{first, nil, nil}},
		{"middle fails", []error{nil, first, nil}},
		{"multiple fail", []error{first, nil, second}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mocks := make([]*mockPinger, len(tc.failures))
			pingers := make([]Pinger, len(tc.failures))
			for i, failure := range tc.failures {
				mocks[i] = &mockPinger{err: failure}
				pingers[i] = mocks[i]
			}
			ctx := context.Background()
			err := NewService(pingers...).Ready(ctx)
			wantFailure := false

			for i, mock := range mocks {
				if mock.calls != 1 || mock.contexts[0] != ctx {
					t.Errorf("pinger %d: expected one call with the caller context", i)
				}
				if mock.err != nil {
					wantFailure = true
					if !errors.Is(err, mock.err) {
						t.Errorf("lost cause from pinger %d", i)
					}
				}
			}
			if (err != nil) != wantFailure {
				t.Fatalf("unexpected readiness result: %v", err)
			}
		})
	}
}

func TestServiceReadyEmpty(t *testing.T) {
	for _, service := range []*Service{NewService(), NewService([]Pinger{}...), NewService([]Pinger(nil)...), {}} {
		if err := service.Ready(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
}

func TestServiceReadyNilPinger(t *testing.T) {
	var typedNil *mockPinger
	var nilFunc pingFunc
	for _, tc := range []struct {
		name   string
		pinger Pinger
	}{
		{"nil interface", nil},
		{"nil pointer", typedNil},
		{"nil function", nilFunc},
	} {
		t.Run(tc.name, func(t *testing.T) {
			before, after := &mockPinger{}, &mockPinger{}
			service := NewService(before, tc.pinger, after)

			err := service.Ready(context.Background())

			if err == nil || !strings.Contains(err.Error(), "nil pinger") {
				t.Fatalf("expected nil pinger error, got %v", err)
			}
			if before.calls != 1 || after.calls != 1 {
				t.Fatal("nil pinger prevented other checks")
			}
		})
	}
}

func TestServiceReadyCanceledOrExpired(t *testing.T) {
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	expired, stop := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer stop()

	for _, ctx := range []context.Context{canceled, expired} {
		for _, withDependency := range []bool{false, true} {
			mock := &mockPinger{}
			service := NewService()
			if withDependency {
				service = NewService(mock)
			}

			if err := service.Ready(ctx); !errors.Is(err, ctx.Err()) {
				t.Fatalf("expected context error, got %v", err)
			}
			if withDependency && mock.calls != 1 {
				t.Fatal("dependency was not checked")
			}
		}
	}
}

func TestNewServiceCopiesPingers(t *testing.T) {
	original, replacement := &mockPinger{}, &mockPinger{}
	pingers := []Pinger{original}
	service := NewService(pingers...)
	pingers[0] = replacement

	if err := service.Ready(context.Background()); err != nil {
		t.Fatal(err)
	}
	if original.calls != 1 || replacement.calls != 0 {
		t.Fatal("caller mutation changed registered dependencies")
	}
}
