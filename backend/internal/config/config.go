package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	Port            string
	JWTSecret       string
	Database        DatabaseConfig
	UploadDir       string
	DefaultTimezone string // IANA timezone, e.g. "Asia/Shanghai". Used as fallback.
}

type DatabaseConfig struct {
	DSN string
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

	jwtSecret, err := resolveJWTSecret(abs)
	if err != nil {
		return nil, fmt.Errorf("resolve jwt secret: %w", err)
	}

	cfg := &Config{
		Port:      envOrDefault("PORT", "8080"),
		JWTSecret: jwtSecret,
		Database: DatabaseConfig{
			DSN: dbDSN,
		},
		UploadDir:       uploadDir,
		DefaultTimezone: envOrDefault("DEFAULT_TIMEZONE", "Asia/Shanghai"),
	}
	return cfg, nil
}

// resolveJWTSecret returns the JWT signing key.
// Priority: 1) JWT_SECRET env var  2) .jwt_secret file in dataDir  3) auto-generate and persist.
func resolveJWTSecret(dataDir string) (string, error) {
	if s := os.Getenv("JWT_SECRET"); s != "" {
		return s, nil
	}

	secretFile := filepath.Join(dataDir, ".jwt_secret")
	if data, err := os.ReadFile(secretFile); err == nil && len(data) > 0 {
		return string(data), nil
	}

	// Auto-generate a random 64-byte secret
	bytes := make([]byte, 64)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate random secret: %w", err)
	}
	secret := hex.EncodeToString(bytes)

	if err := os.WriteFile(secretFile, []byte(secret), 0600); err != nil {
		return "", fmt.Errorf("write jwt secret file: %w", err)
	}

	fmt.Printf("🔐 JWT secret auto-generated and saved to %s\n", secretFile)
	return secret, nil
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
