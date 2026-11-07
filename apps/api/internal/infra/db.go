// Package infra owns infrastructure connections; no schema or migrations run here.
package infra

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"time"

	"auto-comic/internal/config"

	_ "github.com/lib/pq"
)

const StartupPingTimeout = 5 * time.Second

// DSN uses URL escaping rather than interpolating credentials into a keyword DSN.
// Never log the result: it contains the database password.
func DSN(cfg config.Database) string {
	u := url.URL{
		Scheme:  "postgres",
		User:    url.UserPassword(cfg.User, cfg.Password),
		Host:    net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port)),
		Path:    "/" + cfg.Name,
		RawPath: "/" + url.PathEscape(cfg.Name),
	}
	query := url.Values{}
	query.Set("sslmode", "disable")
	query.Set("connect_timeout", "5")
	u.RawQuery = query.Encode()
	return u.String()
}

// OpenDB configures the pool and verifies connectivity before returning ownership.
func OpenDB(ctx context.Context, cfg config.Database) (*sql.DB, error) {
	db, err := sql.Open("postgres", DSN(cfg))
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)

	pingCtx, cancel := context.WithTimeout(ctx, StartupPingTimeout)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return db, nil
}
