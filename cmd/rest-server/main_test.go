package main

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRunMissingConfig(t *testing.T) {
	err := run(context.Background(), serverOptions{configPath: filepath.Join(t.TempDir(), "missing.toml"), address: "127.0.0.1:0"})
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected missing configuration error, got %v", err)
	}
}

func TestServeGracefulShutdown(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	started := make(chan struct{})
	shutdownStarted := make(chan struct{})
	server := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		<-shutdownStarted
		if r.Context().Err() != nil {
			t.Error("shutdown canceled an active request")
		}
		_, _ = io.WriteString(w, "finished")
	})}
	server.RegisterOnShutdown(func() { close(shutdownStarted) })
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- serve(ctx, server, listener) }()

	requestDone := make(chan error, 1)
	go func() {
		client := &http.Client{Timeout: 3 * time.Second}
		response, err := client.Get("http://" + listener.Addr().String())
		if err != nil {
			requestDone <- err
			return
		}
		defer response.Body.Close()
		body, err := io.ReadAll(response.Body)
		if err == nil && (response.StatusCode != 200 || string(body) != "finished") {
			err = errors.New("active request did not complete")
		}
		requestDone <- err
	}()

	select {
	case <-started:
	case <-time.After(3 * time.Second):
		t.Fatal("request did not start")
	}
	cancel()

	if err := <-requestDone; err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("server did not stop")
	}
}

func TestServeListenerFailure(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	listener.Close()

	if err := serve(context.Background(), &http.Server{}, listener); !errors.Is(err, net.ErrClosed) {
		t.Fatalf("expected listener error, got %v", err)
	}
}
