package infra

import (
	"net"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"auto-comic/internal/config"

	"github.com/lib/pq"
)

func TestDSNRoundTrip(t *testing.T) {
	for _, host := range []string{"localhost", "127.0.0.1", "::1"} {
		t.Run(host, func(t *testing.T) {
			cfg := config.Database{
				Host: host, Port: 7001, User: "user:@/' 中文", Password: " p@ss:/'\\?#&=%+中文 ",
				Name: "auto/comic?#% 中文",
			}
			dsn := DSN(cfg)
			u, err := url.Parse(dsn)
			if err != nil {
				t.Fatal(err)
			}
			password, ok := u.User.Password()
			if !ok || password != cfg.Password || u.User.Username() != cfg.User || strings.TrimPrefix(u.Path, "/") != cfg.Name {
				t.Fatal("DSN did not preserve escaped credentials/database")
			}
			if u.Host != net.JoinHostPort(host, strconv.Itoa(cfg.Port)) || u.Query().Get("sslmode") != "disable" || u.Query().Get("connect_timeout") != "5" {
				t.Fatal("DSN lost host, SSL mode, or connect timeout")
			}
			if _, err := pq.ParseURL(dsn); err != nil {
				t.Fatalf("lib/pq rejected DSN: %v", err)
			}
		})
	}
}
