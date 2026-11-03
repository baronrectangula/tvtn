package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const localTOML = `[postgres]
host = "127.0.0.1"
port = 5432
database = "tvtn"
ssl_mode = "disable"
user = "ignored-file-user"
password = "ignored-file-password"
`

func writeConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "custom.toml")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoad(t *testing.T) {
	t.Setenv("POSTGRES_USER", "env-user")
	t.Setenv("POSTGRES_PASSWORD", "env-password")
	var cfg Config
	if err := cfg.Load(writeConfig(t, localTOML)); err != nil {
		t.Fatal(err)
	}
	p := cfg.Postgres
	if p.Host != "127.0.0.1" || p.Port != 5432 || p.Database != "tvtn" || p.SSLMode != "disable" {
		t.Fatal("incorrect TOML settings")
	}
	if p.User.Expose() != "env-user" || p.Password.Expose() != "env-password" {
		t.Fatal("credentials must come from environment")
	}
	if strings.Contains(fmt.Sprintf("%+v", cfg), "env-password") {
		t.Fatal("logged secret")
	}
	if err := cfg.Load(filepath.Join(t.TempDir(), "missing.toml")); err == nil {
		t.Fatal("expected missing file error")
	}
	if cfg.Postgres != p {
		t.Fatal("failed reload changed config")
	}
}

func TestLoadInvalidConfig(t *testing.T) {
	for _, tc := range []struct{ name, content, user, password string }{
		{"invalid TOML", "[", "u", "p"},
		{"missing settings", "", "u", "p"},
		{"missing user", localTOML, "", "p"},
		{"missing password", localTOML, "u", ""},
		{"invalid port", strings.Replace(localTOML, "5432", "65536", 1), "u", "p"},
		{"invalid SSL mode", strings.Replace(localTOML, "disable", "invalid", 1), "u", "p"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("POSTGRES_USER", tc.user)
			t.Setenv("POSTGRES_PASSWORD", tc.password)
			var cfg Config
			if err := cfg.Load(writeConfig(t, tc.content)); err == nil {
				t.Fatal("expected load error")
			}
		})
	}
}
