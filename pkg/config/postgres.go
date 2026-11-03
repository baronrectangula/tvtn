package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"

	"tvtn/pkg/serect"

	"github.com/pelletier/go-toml/v2"
)

type Postgres struct {
	Host     string `toml:"host"`
	Port     uint16 `toml:"port"`
	Database string `toml:"database"`
	SSLMode  string `toml:"ssl_mode"`

	User     serect.Value `toml:"-"`
	Password serect.Value `toml:"-"`
}

// DSN returns a PostgreSQL connection URI wrapped to prevent accidental logging.
// Expose the returned string only when passing it to the database driver.
func (p *Postgres) DSN() (serect.Value, error) {
	user, err := p.User.ExposeAsString()
	if err != nil {
		return serect.Value{}, fmt.Errorf("config.Postgres.DSN: %w", err)
	}

	password, err := p.Password.ExposeAsString()
	if err != nil {
		return serect.Value{}, fmt.Errorf("config.Postgres.DSN: %w", err)
	}

	uri := url.URL{
		Scheme:   "postgresql",
		User:     url.UserPassword(user, password),
		Host:     net.JoinHostPort(p.Host, strconv.Itoa(int(p.Port))),
		Path:     "/" + p.Database,
		RawPath:  "/" + url.PathEscape(p.Database),
		RawQuery: url.Values{"sslmode": {p.SSLMode}}.Encode(),
	}

	return serect.New(uri.String()), nil
}

// Load reads non-secret settings from the [postgres] section of path.
// Credentials come exclusively from POSTGRES_USER and POSTGRES_PASSWORD.
// The receiver is unchanged on failure.
func (p *Postgres) Load(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("config.Postgres.Load: %w", err)
	}

	var file struct {
		Postgres Postgres `toml:"postgres"`
	}

	if err := toml.Unmarshal(data, &file); err != nil {
		return fmt.Errorf("config.Postgres.Load: %w", err)
	}

	settings := file.Postgres
	if settings.Host == "" {
		return fmt.Errorf("config.Postgres.Load: postgres.host is required")
	}

	if settings.Port == 0 {
		return fmt.Errorf("config.Postgres.Load: postgres.port must be between 1 and 65535")
	}

	if settings.Database == "" {
		return fmt.Errorf("config.Postgres.Load: postgres.database is required")
	}
	switch settings.SSLMode {
	case "disable", "allow", "prefer", "require", "verify-ca", "verify-full":
	default:
		return fmt.Errorf("config.Postgres.Load: postgres.ssl_mode is missing or invalid")
	}

	user := os.Getenv("POSTGRES_USER")
	if user == "" {
		return fmt.Errorf("config.Postgres.Load: POSTGRES_USER is required")
	}

	password := os.Getenv("POSTGRES_PASSWORD")
	if password == "" {
		return fmt.Errorf("config.Postgres.Load: POSTGRES_PASSWORD is required")
	}

	settings.User = serect.New(user)
	settings.Password = serect.New(password)

	*p = settings

	return nil
}
