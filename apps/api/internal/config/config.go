// Package config loads validated process configuration without reading .env files.
package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"unicode"
)

type Database struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
}

type Config struct {
	Database Database
	HTTPAddr string
}

// Load distinguishes unset values (defaults) from explicitly empty values (errors).
// Passwords are required and preserved verbatim, including spaces and punctuation.
func Load() (Config, error) {
	return load(os.LookupEnv)
}

func load(lookup func(string) (string, bool)) (Config, error) {
	value := func(key, fallback string) string {
		if v, ok := lookup(key); ok {
			return v
		}
		return fallback
	}
	cfg := Config{
		Database: Database{
			Host:     value("DB_HOST", "localhost"),
			User:     value("DB_USER", "auto-comic"),
			Password: value("DB_PASSWORD", ""),
			Name:     value("DB_NAME", "auto-comic"),
		},
		HTTPAddr: value("HTTP_ADDR", ":7001"),
	}
	for _, field := range []struct{ key, value string }{
		{"DB_HOST", cfg.Database.Host}, {"DB_USER", cfg.Database.User},
		{"DB_NAME", cfg.Database.Name},
		{"HTTP_ADDR", cfg.HTTPAddr},
	} {
		if strings.TrimSpace(field.value) == "" || strings.TrimSpace(field.value) != field.value || strings.ContainsRune(field.value, '\x00') {
			return Config{}, fmt.Errorf("%s must be non-empty without surrounding whitespace or NUL", field.key)
		}
	}
	if strings.TrimSpace(cfg.Database.Password) == "" || strings.ContainsRune(cfg.Database.Password, '\x00') {
		return Config{}, fmt.Errorf("DB_PASSWORD must be set and non-blank without NUL")
	}
	if strings.ContainsAny(cfg.Database.Host, "/?#@[]") || strings.ContainsFunc(cfg.Database.Host, unicode.IsSpace) || (strings.Contains(cfg.Database.Host, ":") && net.ParseIP(cfg.Database.Host) == nil) {
		return Config{}, fmt.Errorf("DB_HOST must be a hostname or an unbracketed IP address")
	}
	port, err := parsePort(value("DB_PORT", "7002"))
	if err != nil {
		return Config{}, fmt.Errorf("DB_PORT must be an integer between 1 and 65535")
	}
	cfg.Database.Port = port
	host, httpPort, err := net.SplitHostPort(cfg.HTTPAddr)
	if err != nil {
		return Config{}, fmt.Errorf("HTTP_ADDR must have host:port format")
	}
	if _, err := parsePort(httpPort); err != nil || strings.ContainsFunc(host, unicode.IsSpace) {
		return Config{}, fmt.Errorf("HTTP_ADDR must have a valid host and port between 1 and 65535")
	}
	return cfg, nil
}

func parsePort(value string) (int, error) {
	for _, c := range value {
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("invalid port")
		}
	}
	port, err := strconv.Atoi(value)
	if err != nil || port < 1 || port > 65535 {
		return 0, fmt.Errorf("invalid port")
	}
	return port, nil
}
