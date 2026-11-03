package config

import (
	"encoding/json"
	"fmt"
	"net/url"
	"testing"
	"tvtn/pkg/serect"
)

func TestPostgresDSN(t *testing.T) {
	for _, host := range []string{"127.0.0.1", "::1"} {
		t.Run(host, func(t *testing.T) {
			p := Postgres{Host: host, Port: 5432, Database: "db/name ?#%", SSLMode: "disable", User: serect.New("user@name"), Password: serect.New("p@ss:/?#% word")}
			dsn, err := p.DSN()
			if err != nil {
				t.Fatal(err)
			}
			raw, err := dsn.ExposeAsString()
			if err != nil {
				t.Fatal(err)
			}
			u, err := url.Parse(raw)
			if err != nil {
				t.Fatal("invalid DSN URL")
			}
			password, _ := u.User.Password()
			if u.Scheme != "postgresql" || u.Hostname() != host || u.Port() != "5432" || u.Path != "/"+p.Database || u.User.Username() != "user@name" || password != "p@ss:/?#% word" || u.Query().Get("sslmode") != "disable" {
				t.Fatal("DSN did not preserve connection parameters")
			}
			if fmt.Sprintf("%+v", dsn) != "[REDACTED]" {
				t.Fatal("DSN not redacted")
			}
			data, err := json.Marshal(dsn)
			if err != nil || string(data) != `"[REDACTED]"` {
				t.Fatal("DSN JSON not redacted")
			}
		})
	}
}

func TestPostgresDSNInvalidCredentials(t *testing.T) {
	for _, p := range []Postgres{
		{User: serect.New(struct{}{}), Password: serect.New("password")},
		{User: serect.New("user"), Password: serect.New(struct{}{})},
	} {
		if _, err := p.DSN(); err == nil {
			t.Fatal("expected credential conversion error")
		}
	}
}
