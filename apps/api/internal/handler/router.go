// Package handler defines transport endpoints, independently of database drivers.
package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const ReadinessTimeout = 2 * time.Second

type Pinger interface {
	PingContext(context.Context) error
}

func NewRouter(db Pinger, logger *slog.Logger) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(requestLog(logger))
	router.Use(middleware.Recoverer)

	live := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		if r.Method != http.MethodHead {
			_, _ = w.Write([]byte("OK\n"))
		}
	}
	ready := func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), ReadinessTimeout)
		defer cancel()
		status, payload := http.StatusOK, map[string]string{"status": "ok", "database": "ok"}
		if db == nil || db.PingContext(ctx) != nil {
			status = http.StatusServiceUnavailable
			payload = map[string]string{"status": "error", "database": "unavailable"}
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(status)
		if r.Method != http.MethodHead {
			_ = json.NewEncoder(w).Encode(payload)
		}
	}
	router.Get("/health", live)
	router.Head("/health", live)
	router.Get("/ready", ready)
	router.Head("/ready", ready)
	// Future business routes belong under /v1. No placeholder catch-all is registered;
	// /v1 and all unknown paths retain chi's normal 404 behavior.
	return router
}

func requestLog(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			started := time.Now()
			requestID := middleware.GetReqID(r.Context())
			w.Header().Set("X-Request-ID", requestID)
			wrapped := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			defer func() {
				status := wrapped.Status()
				if status == 0 {
					status = http.StatusOK
				}
				logger.InfoContext(r.Context(), "HTTP request", "request_id", requestID,
					"method", r.Method, "path", r.URL.Path, "status", status,
					"duration", time.Since(started))
			}()
			next.ServeHTTP(wrapped, r)
		})
	}
}
