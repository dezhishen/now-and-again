package config

import (
	"os"
)

type Config struct {
	Port      string
	JWTSecret string
	Database  DatabaseConfig
}

type DatabaseConfig struct {
	Driver string // "sqlite" | "postgres"
	DSN    string
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:      envOrDefault("PORT", "8080"),
		JWTSecret: envOrDefault("JWT_SECRET", "now-and-again-dev-secret-change-me"),
		Database: DatabaseConfig{
			Driver: envOrDefault("DB_DRIVER", "sqlite"),
			DSN:    envOrDefault("DB_DSN", "now-and-again.db"),
		},
	}
	return cfg, nil
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
