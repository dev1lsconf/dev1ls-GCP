package config

import (
	"os"
)

// Config holds runtime configuration loaded from environment variables
type Config struct {
	Port        string
	Environment string
	Version     string
}

// Load reads configuration from environment variables with safe defaults
func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port for Cloud Run, ECS, and local dev
	}

	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	version := os.Getenv("APP_VERSION")
	if version == "" {
		version = "1.0.0"
	}

	return &Config{
		Port:        port,
		Environment: env,
		Version:     version,
	}
}
