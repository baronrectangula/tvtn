package health

import (
	"context"
	"io"
	"net/http"
	"time"
)

const readinessTimeout = 2 * time.Second

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Register mounts health endpoints directly on the root mux, without an API prefix.
// GET patterns also accept HEAD requests.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /livez", h.Livez)
	mux.HandleFunc("GET /readyz", h.Readyz)
}

// Livez always returns OK without checking external dependencies.
func (h *Handler) Livez(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)

	if r.Method != http.MethodHead {
		_, _ = io.WriteString(w, "ok\n")
	}
}

// Readyz returns 503 when any dependency fails or the readiness deadline expires.
// Dependency errors are deliberately excluded from the HTTP response.
func (h *Handler) Readyz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), readinessTimeout)
	defer cancel()

	status, body := http.StatusOK, "ok\n"
	if err := h.service.Ready(ctx); err != nil {
		status, body = http.StatusServiceUnavailable, "not ready\n"
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)

	if r.Method != http.MethodHead {
		_, _ = io.WriteString(w, body)
	}
}
