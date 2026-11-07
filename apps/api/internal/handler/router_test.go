package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type fakePinger func(context.Context) error

func (f fakePinger) PingContext(ctx context.Context) error { return f(ctx) }

func TestEndpoints(t *testing.T) {
	for _, method := range []string{http.MethodGet, http.MethodHead} {
		for _, path := range []string{"/health", "/ready"} {
			t.Run(method+path, func(t *testing.T) {
				var logs bytes.Buffer
				calls := 0
				router := NewRouter(fakePinger(func(ctx context.Context) error {
					calls++
					deadline, ok := ctx.Deadline()
					if !ok || time.Until(deadline) > ReadinessTimeout {
						t.Error("readiness ping missing bounded deadline")
					}
					return nil
				}), slog.New(slog.NewJSONHandler(&logs, nil)))
				rec := httptest.NewRecorder()
				router.ServeHTTP(rec, httptest.NewRequest(method, path, nil))
				if rec.Code != http.StatusOK || rec.Header().Get("X-Request-ID") == "" {
					t.Fatalf("unexpected response: %d %v", rec.Code, rec.Header())
				}
				if method == http.MethodHead && rec.Body.Len() != 0 {
					t.Fatal("HEAD must not have body")
				}
				if path == "/health" {
					if calls != 0 || rec.Header().Get("Content-Type") != "text/plain; charset=utf-8" || (method == http.MethodGet && rec.Body.String() != "OK\n") {
						t.Fatal("liveness unexpectedly depends on database or has incorrect response")
					}
				} else {
					if calls != 1 || rec.Header().Get("Content-Type") != "application/json; charset=utf-8" {
						t.Fatal("readiness must ping database and return JSON")
					}
					if method == http.MethodGet {
						var body map[string]string
						if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil || body["status"] != "ok" || body["database"] != "ok" || len(body) != 2 {
							t.Fatalf("unexpected JSON: %s", rec.Body.String())
						}
					}
				}
				if !strings.Contains(logs.String(), rec.Header().Get("X-Request-ID")) || !strings.Contains(logs.String(), `"status":200`) {
					t.Fatal("request log lacks request ID or response status")
				}
			})
		}
	}
}

func TestReadinessFailureDoesNotLeak(t *testing.T) {
	for _, db := range []Pinger{nil, fakePinger(func(context.Context) error { return errors.New("secret password and internal host") })} {
		for _, method := range []string{http.MethodGet, http.MethodHead} {
			rec := httptest.NewRecorder()
			NewRouter(db, nil).ServeHTTP(rec, httptest.NewRequest(method, "/ready", nil))
			if rec.Code != http.StatusServiceUnavailable || strings.Contains(rec.Body.String(), "secret") || strings.Contains(rec.Body.String(), "host") {
				t.Fatalf("unexpected failure response: %d %s", rec.Code, rec.Body.String())
			}
			if method == http.MethodHead && rec.Body.Len() != 0 {
				t.Fatal("failed HEAD must not have body")
			}
			if method == http.MethodGet {
				var body map[string]string
				if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil || body["status"] != "error" || body["database"] != "unavailable" || len(body) != 2 {
					t.Fatalf("unexpected failure JSON: %s", rec.Body.String())
				}
			}
		}
	}
}

func TestReadinessPropagatesCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	db := fakePinger(func(ctx context.Context) error {
		if !errors.Is(ctx.Err(), context.Canceled) {
			t.Error("request cancellation not propagated")
		}
		return ctx.Err()
	})
	rec := httptest.NewRecorder()
	NewRouter(db, nil).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/ready", nil).WithContext(ctx))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("got %d", rec.Code)
	}
}

func TestUnknownRoutes(t *testing.T) {
	for _, path := range []string{"/", "/unknown", "/v1", "/v1/", "/v1/anything"} {
		rec := httptest.NewRecorder()
		NewRouter(nil, nil).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != http.StatusNotFound {
			t.Fatalf("%s returned %d, expected 404", path, rec.Code)
		}
	}
	rec := httptest.NewRecorder()
	NewRouter(nil, nil).ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/health", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST /health returned %d", rec.Code)
	}
}

func TestRecoverer(t *testing.T) {
	db := fakePinger(func(context.Context) error { panic("private panic details") })
	rec := httptest.NewRecorder()
	NewRouter(db, nil).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/ready", nil))
	if rec.Code != http.StatusInternalServerError || strings.Contains(rec.Body.String(), "private") || rec.Header().Get("X-Request-ID") == "" {
		t.Fatalf("unsafe panic response: %d %s", rec.Code, rec.Body.String())
	}
}
