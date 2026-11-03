package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"tvtn/pkg/config"
	"tvtn/pkg/postgres"
)

const (
	startupTimeout  = 10 * time.Second
	shutdownTimeout = 10 * time.Second
)

func main() {
	var options serverOptions
	options.registerFlags(flag.CommandLine)
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, options); err != nil {
		slog.Error("REST server stopped", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, options serverOptions) error {
	var cfg config.Config
	if err := cfg.Load(options.configPath); err != nil {
		return fmt.Errorf("rest-server.run: %w", err)
	}

	startupCtx, cancel := context.WithTimeout(ctx, startupTimeout)
	pool, err := postgres.New(startupCtx, &cfg.Postgres)
	cancel()
	if err != nil {
		return fmt.Errorf("rest-server.run: %w", err)
	}
	defer pool.Close()

	server := options.newServer(newRouter(pool))

	listener, err := net.Listen("tcp", server.Addr)
	if err != nil {
		return fmt.Errorf("rest-server.run: %w", err)
	}
	defer listener.Close()

	slog.Info("REST server listening", "address", listener.Addr().String())

	return serve(ctx, server, listener)
}

func serve(ctx context.Context, server *http.Server, listener net.Listener) error {
	serveErrors := make(chan error, 1)
	go func() {
		serveErrors <- server.Serve(listener)
	}()

	select {
	case err := <-serveErrors:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("rest-server.serve: %w", err)
		}
		return nil

	case <-ctx.Done():
		slog.Info("REST server shutting down")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		_ = server.Close()
		<-serveErrors
		return fmt.Errorf("rest-server.serve: %w", err)
	}

	if err := <-serveErrors; err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("rest-server.serve: %w", err)
	}

	return nil
}
