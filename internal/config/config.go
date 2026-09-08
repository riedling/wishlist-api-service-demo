// Package config centralizes application configuration loaded from
// environment variables, with sensible defaults for local development.
package config

import (
	"os"
)

// Config holds all runtime configuration for the service.
type Config struct {
	// Port is the TCP port the HTTP server listens on.
	Port string
	// Env is the running environment (e.g. "development", "production").
	Env string
	// LogLevel controls verbosity of application logging.
	LogLevel string
}

// Load reads configuration from environment variables, falling back to
// defaults when a variable is not set.
func Load() *Config {
	return &Config{
		Port:     getEnv("PORT", "8080"),
		Env:      getEnv("APP_ENV", "development"),
		LogLevel: getEnv("LOG_LEVEL", "info"),
	}
}

// IsProduction reports whether the service is running in production mode.
func (c *Config) IsProduction() bool {
	return c.Env == "production"
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}
