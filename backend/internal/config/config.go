package config

import (
	"os"
	"path/filepath"
)

type Config struct {
	Port      string
	JWTSecret string
	Database  DatabaseConfig
	UploadDir string
}

type DatabaseConfig struct {
	Driver string // "sqlite" | "postgres"
	DSN    string
}

func Load() (*Config, error) {
	dataDir := os.Getenv("DATA_DIR")
	dbDSN := envOrDefault("DB_DSN", "now-and-again.db")
	uploadDir := envOrDefault("UPLOAD_DIR", "./uploads")

	if dataDir != "" {
		abs, err := filepath.Abs(dataDir)
		if err == nil {
			os.MkdirAll(abs, 0755)
			dbDSN = filepath.Join(abs, "now-and-again.db")
			uploadDir = filepath.Join(abs, "uploads")
		}
	}

	cfg := &Config{
		Port:      envOrDefault("PORT", "8080"),
		JWTSecret: envOrDefault("JWT_SECRET", "now-and-again-dev-secret-change-me"),
		Database: DatabaseConfig{
			Driver: envOrDefault("DB_DRIVER", "sqlite"),
			DSN:    dbDSN,
		},
		UploadDir: uploadDir,
	}
	return cfg, nil
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
