package main

import (
	"net/http"

	"tvtn/internal/health"
)

// newRouter wires domain services and registers their HTTP routes.
func newRouter(database health.Pinger) *http.ServeMux {
	mux := http.NewServeMux()

	healthService := health.NewService(database)
	healthHandler := health.NewHandler(healthService)
	healthHandler.Register(mux)

	return mux
}
