package config

import (
	"strings"
	"testing"
)

func loadValues(values map[string]string) (Config, error) {
	return load(func(key string) (string, bool) { v, ok := values[key]; return v, ok })
}

func TestDefaults(t *testing.T) {
	cfg, err := loadValues(map[string]string{"DB_PASSWORD": "secret"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Database != (Database{Host: "localhost", Port: 7002, User: "auto-comic", Password: "secret", Name: "auto-comic"}) || cfg.HTTPAddr != ":7001" {
		t.Fatalf("unexpected defaults: %#v", cfg)
	}
}

func TestOverridesAndPasswordPreserved(t *testing.T) {
	values := map[string]string{
		"DB_HOST": "::1", "DB_PORT": "5432", "DB_USER": "service",
		"DB_PASSWORD": " p@ss:/?#&'\\word ", "DB_NAME": "other", "HTTP_ADDR": "127.0.0.1:9000",
	}
	cfg, err := loadValues(values)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Database.Password != values["DB_PASSWORD"] || cfg.Database.Host != "::1" || cfg.Database.Port != 5432 || cfg.Database.User != "service" || cfg.Database.Name != "other" || cfg.HTTPAddr != "127.0.0.1:9000" {
		t.Fatalf("overrides not preserved: %#v", cfg)
	}
}

func TestInvalidConfiguration(t *testing.T) {
	cases := []struct{ key, value string }{
		{"DB_PASSWORD", ""}, {"DB_PASSWORD", "  "}, {"DB_PASSWORD", "a\x00b"},
		{"DB_HOST", ""}, {"DB_HOST", " localhost"}, {"DB_HOST", "localhost:5432"}, {"DB_HOST", "postgres://localhost"},
		{"DB_PORT", ""}, {"DB_PORT", "abc"}, {"DB_PORT", "0"}, {"DB_PORT", "65536"}, {"DB_PORT", "-1"}, {"DB_PORT", "+5432"}, {"DB_PORT", " 5432"},
		{"DB_USER", ""}, {"DB_NAME", " "},
		{"HTTP_ADDR", ""}, {"HTTP_ADDR", "8080"}, {"HTTP_ADDR", ":0"}, {"HTTP_ADDR", ":65536"}, {"HTTP_ADDR", ":abc"}, {"HTTP_ADDR", "bad host:8080"},
	}
	for _, tc := range cases {
		t.Run(tc.key+"/"+tc.value, func(t *testing.T) {
			values := map[string]string{"DB_PASSWORD": "private-password"}
			values[tc.key] = tc.value
			_, err := loadValues(values)
			if err == nil || !strings.Contains(err.Error(), tc.key) {
				t.Fatalf("expected %s validation error, got %v", tc.key, err)
			}
			if strings.Contains(err.Error(), "private-password") {
				t.Fatal("error exposed password")
			}
		})
	}
	if _, err := loadValues(nil); err == nil {
		t.Fatal("missing password accepted")
	}
}

func TestLoadFromEnvironment(t *testing.T) {
	for key, value := range map[string]string{"DB_HOST": "localhost", "DB_PORT": "7002", "DB_USER": "auto-comic", "DB_PASSWORD": "test", "DB_NAME": "auto-comic", "HTTP_ADDR": ":7001"} {
		t.Setenv(key, value)
	}
	if _, err := Load(); err != nil {
		t.Fatal(err)
	}
}
