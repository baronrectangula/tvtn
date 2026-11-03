// Package health provides liveness and dependency readiness checks.
package health

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"slices"
)

// Pinger checks a dependency's availability. Implementations must honor context
// cancellation and support concurrent calls from independent readiness requests.
// A *pgxpool.Pool implements Pinger directly.
type Pinger interface {
	Ping(context.Context) error
}

type Service struct {
	pingers []Pinger
}

func NewService(pingers ...Pinger) *Service {
	return &Service{pingers: slices.Clone(pingers)}
}

// Ready pings every dependency, including when an earlier check fails.
// With no dependencies, readiness succeeds unless the context is canceled.
// Nil dependencies, including typed nil values, fail readiness without preventing
// the remaining dependencies from being checked.
func (s *Service) Ready(ctx context.Context) error {
	var failures []error

	for _, pinger := range s.pingers {
		if isNilPinger(pinger) {
			failures = append(failures, errors.New("nil pinger"))
			continue
		}

		if err := pinger.Ping(ctx); err != nil {
			failures = append(failures, err)
		}
	}

	if err := ctx.Err(); err != nil {
		failures = append(failures, err)
	}

	if err := errors.Join(failures...); err != nil {
		return fmt.Errorf("health.Service.Ready: %w", err)
	}

	return nil
}

func isNilPinger(pinger Pinger) bool {
	if pinger == nil {
		return true
	}

	value := reflect.ValueOf(pinger)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}
