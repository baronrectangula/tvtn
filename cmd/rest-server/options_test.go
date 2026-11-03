package main

import (
	"context"
	"errors"
	"flag"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestServerFlags(t *testing.T) {
	for _, tc := range []struct {
		name     string
		args     []string
		address  string
		timeouts [4]time.Duration
	}{
		{"defaults", nil, "127.0.0.1:8080", [4]time.Duration{5 * time.Second, 10 * time.Second, 10 * time.Second, 60 * time.Second}},
		{"overrides", []string{"-addr=:9090", "-read-header-timeout=500ms", "-read-timeout=20s", "-write-timeout=30s", "-idle-timeout=2m"}, ":9090", [4]time.Duration{500 * time.Millisecond, 20 * time.Second, 30 * time.Second, 2 * time.Minute}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var options serverOptions
			flags := flag.NewFlagSet("test", flag.ContinueOnError)
			options.registerFlags(flags)
			if err := flags.Parse(tc.args); err != nil {
				t.Fatal(err)
			}

			router := newRouter(nil)
			server := options.newServer(router)
			actual := [4]time.Duration{server.ReadHeaderTimeout, server.ReadTimeout, server.WriteTimeout, server.IdleTimeout}
			if server.Addr != tc.address || actual != tc.timeouts || server.Handler != router {
				t.Fatal("server did not use command-line settings")
			}
		})
	}
}

type mockDatabase struct {
	err   error
	calls int
}

func (m *mockDatabase) Ping(context.Context) error { m.calls++; return m.err }

func TestRouterWiresDatabaseReadiness(t *testing.T) {
	database := &mockDatabase{err: errors.New("unavailable")}
	router := newRouter(database)

	for _, tc := range []struct {
		path          string
		status, calls int
	}{
		{"/livez", http.StatusOK, 0},
		{"/readyz", http.StatusServiceUnavailable, 1},
	} {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest("GET", tc.path, nil))
		if response.Code != tc.status || database.calls != tc.calls {
			t.Fatalf("incorrect wiring for %s", tc.path)
		}
	}
}
