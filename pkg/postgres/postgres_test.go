package postgres

import (
	"context"
	"errors"
	"strings"
	"testing"

	"tvtn/pkg/config"
	"tvtn/pkg/serect"
)

func TestNewInvalidCredentials(t *testing.T) {
	pool, err := New(context.Background(), &config.Postgres{})
	if pool != nil {
		pool.Close()
		t.Fatal("expected no pool")
	}
	if err == nil || !strings.HasPrefix(err.Error(), "postgres.New: ") {
		t.Fatal("expected wrapped configuration error")
	}
}

func TestNewCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	pool, err := New(ctx, &config.Postgres{
		Host: "127.0.0.1", Port: 5432, Database: "tvtn", SSLMode: "disable",
		User: serect.New("tvtn"), Password: serect.New("tvtn_local_password"),
	})
	if pool != nil {
		pool.Close()
		t.Fatal("expected no pool")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected cancellation, got %v", err)
	}
}
