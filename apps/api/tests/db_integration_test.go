//go:build integration

package tests

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"auto-comic/internal/config"
	"auto-comic/internal/handler"
	"auto-comic/internal/infra"
)

// TestDatabaseReadiness exercises a real PostgreSQL connection without creating
// tables or data. SELECT 1 stays valid when business schemas are introduced.
func TestDatabaseReadiness(t *testing.T) {
	required := func(key string) string {
		t.Helper()
		value := os.Getenv(key)
		if value == "" {
			t.Fatalf("%s is required; run tests/run.sh to start a temporary database", key)
		}
		return value
	}
	port, err := strconv.Atoi(required("TEST_DB_PORT"))
	if err != nil || port < 1 || port > 65535 {
		t.Fatal("TEST_DB_PORT must be between 1 and 65535")
	}
	cfg := config.Database{
		Host: required("TEST_DB_HOST"), Port: port,
		User: required("TEST_DB_USER"), Password: required("TEST_DB_PASSWORD"),
		Name: required("TEST_DB_NAME"),
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	db, err := infra.OpenDB(ctx, cfg)
	if err != nil {
		t.Fatalf("open real PostgreSQL connection: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("close database: %v", err)
		}
	})
	var value int
	if err := db.QueryRowContext(ctx, "SELECT 1").Scan(&value); err != nil || value != 1 {
		t.Fatalf("SELECT 1: value=%d err=%v", value, err)
	}

	router := handler.NewRouter(db, slog.New(slog.NewTextHandler(io.Discard, nil)))
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/ready", nil).WithContext(ctx))
	if response.Code != http.StatusOK {
		t.Fatalf("GET /ready: status=%d body=%s", response.Code, response.Body.String())
	}
	var body map[string]string
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode readiness JSON: %v", err)
	}
	if body["status"] != "ok" || body["database"] != "ok" {
		t.Fatalf("unexpected readiness response: %v", body)
	}
}
