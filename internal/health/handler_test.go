package health

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var _ Pinger = (*pgxpool.Pool)(nil)

type pingFunc func(context.Context) error

func (f pingFunc) Ping(ctx context.Context) error { return f(ctx) }

func TestRoutes(t *testing.T) {
	for _, tc := range []struct {
		name, method, path string
		failure            bool
		status, calls      int
		body               string
	}{
		{"live despite failure", "GET", "/livez", true, 200, 0, "ok\n"},
		{"ready", "GET", "/readyz", false, 200, 1, "ok\n"},
		{"not ready", "GET", "/readyz", true, 503, 1, "not ready\n"},
		{"head live", "HEAD", "/livez", true, 200, 0, ""},
		{"head ready", "HEAD", "/readyz", false, 200, 1, ""},
		{"prefixed route absent", "GET", "/api/v1/readyz", false, 404, 0, "404 page not found\n"},
		{"subpath absent", "GET", "/livez/extra", false, 404, 0, "404 page not found\n"},
		{"unsupported method", "POST", "/readyz", false, 405, 0, "Method Not Allowed\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			service := NewService(pingFunc(func(ctx context.Context) error {
				calls++
				deadline, ok := ctx.Deadline()
				if !ok || time.Until(deadline) > readinessTimeout {
					t.Error("missing readiness deadline")
				}
				if tc.failure {
					return errors.New("private connection details")
				}
				return nil
			}))
			mux := http.NewServeMux()
			NewHandler(service).Register(mux)
			response := httptest.NewRecorder()

			mux.ServeHTTP(response, httptest.NewRequest(tc.method, tc.path, nil))

			if response.Code != tc.status || response.Body.String() != tc.body || calls != tc.calls {
				t.Fatalf("status=%d body=%q calls=%d", response.Code, response.Body.String(), calls)
			}
			if tc.status == 200 || tc.status == 503 {
				if response.Header().Get("Cache-Control") != "no-store" {
					t.Fatal("health response must not be cached")
				}
			}
		})
	}
}

func TestReadinessCancellation(t *testing.T) {
	service := NewService(pingFunc(func(ctx context.Context) error {
		<-ctx.Done()
		return ctx.Err()
	}))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := service.Ready(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected cancellation, got %v", err)
	}
	response := httptest.NewRecorder()
	NewHandler(service).Readyz(response, httptest.NewRequest("GET", "/readyz", nil).WithContext(ctx))
	if response.Code != http.StatusServiceUnavailable {
		t.Fatal("canceled probe should fail")
	}
}
