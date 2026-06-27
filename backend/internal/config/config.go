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
	dataDir := envOrDefault("DATA_DIR", "data")
	abs, err := filepath.Abs(dataDir)
	if err != nil {
		abs = dataDir
	}
	os.MkdirAll(abs, 0755)

	dbDSN := filepath.Join(abs, "now-and-again.db")
	uploadDir := filepath.Join(abs, "uploads")

	// For non-sqlite drivers (e.g. postgres), allow explicit DSN
	if envOrDefault("DB_DRIVER", "sqlite") != "sqlite" {
		if dsn := os.Getenv("DB_DSN"); dsn != "" {
			dbDSN = dsn
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

// BaseDir returns the data directory or current directory.
func (c *Config) BaseDir() string {
	dir := filepath.Dir(c.Database.DSN)
	if dir == "." {
		return "."
	}
	return dir
}
