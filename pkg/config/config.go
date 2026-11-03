// Package config loads application settings explicitly from TOML and environment variables.
package config

import "fmt"

type configLoader interface {
	Load(path string) error
}

type Config struct {
	Postgres Postgres
}

// Load loads each configuration section from path and its environment variables.
// The receiver is updated only after all loaders succeed.
func (c *Config) Load(path string) error {
	var next Config
	loaders := []configLoader{&next.Postgres}

	for _, loader := range loaders {
		if err := loader.Load(path); err != nil {
			return fmt.Errorf("config.Config.Load: %w", err)
		}
	}

	*c = next

	return nil
}
