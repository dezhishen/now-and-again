// Package sdk provides a high-level client SDK for the Now & Again platform.
// It wraps the low-level HTTP clients and adds convenience methods for common workflows
// such as login→save config, template→task creation, and todo processing.
package sdk

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/dezhishen/now-and-again/sdk/internal/client"
	"gopkg.in/yaml.v3"
)

// Config holds persistent SDK configuration stored in ~/.na.yaml.
type Config struct {
	ServerURL string `yaml:"server_url"`
	Token     string `yaml:"token"` // API key (preferred) or JWT
	// Cached user preferences
	ActiveFamilyID   string `yaml:"active_family_id,omitempty"`
	ActiveFamilyName string `yaml:"active_family_name,omitempty"`
}

// Save writes configuration to disk at ~/.na.yaml.
func (c *Config) Save() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("home dir: %w", err)
	}
	cfgFile := filepath.Join(home, ".na.yaml")
	if err := os.MkdirAll(filepath.Dir(cfgFile), 0755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	return os.WriteFile(cfgFile, data, 0600)
}

// LoadConfig reads configuration from ~/.na.yaml.
// Returns zero-value Config if the file does not exist.
func LoadConfig() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("home dir: %w", err)
	}
	cfgFile := filepath.Join(home, ".na.yaml")
	data, err := os.ReadFile(cfgFile)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{ServerURL: "http://localhost:8080"}, nil
		}
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	if cfg.ServerURL == "" {
		cfg.ServerURL = "http://localhost:8080"
	}
	return &cfg, nil
}

// NA is the main SDK entry point. It holds configuration, an authenticated HTTP client,
// and domain-specific clients for all Now & Again APIs.
//
// Usage:
//
//	na, err := sdk.New()
//	if err != nil { ... }
//	// If not yet initialized:
//	if err := na.Init("admin", "12345678"); err != nil { ... }
//	// Use domain clients:
//	families, _ := na.Family.ListMyFamilies(ctx)
type NA struct {
	mu  sync.RWMutex
	cfg *Config

	http *client.HTTPClient

	// Domain clients — direct access for advanced usage.
	User   *client.UserClient
	Family *client.FamilyClient
	ApiKey *client.ApiKeyClient
	Task   *client.TaskClient

	// Cached value: the active family ID (set via Init or SetActiveFamily).
	activeFamilyID string
}

// New creates an NA instance from saved configuration (~/.na.yaml).
// If no saved config exists, a zero-value config is used (server defaults to localhost:8080).
func New() (*NA, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("sdk: load config: %w", err)
	}
	return NewWithConfig(cfg), nil
}

// NewWithConfig creates an NA instance from an explicit Config.
func NewWithConfig(cfg *Config) *NA {
	httpClient := client.NewHTTPClient(cfg.ServerURL, cfg.Token)
	return &NA{
		cfg:            cfg,
		http:           httpClient,
		User:           client.NewUserClient(httpClient),
		Family:         client.NewFamilyClient(httpClient),
		ApiKey:         client.NewApiKeyClient(httpClient),
		Task:           client.NewTaskClient(httpClient),
		activeFamilyID: cfg.ActiveFamilyID,
	}
}

// Config returns the current configuration (read-only).
func (na *NA) Config() *Config {
	na.mu.RLock()
	defer na.mu.RUnlock()
	cp := *na.cfg
	return &cp
}

// SetServerURL changes the server URL and updates the underlying client.
func (na *NA) SetServerURL(url string) {
	na.mu.Lock()
	defer na.mu.Unlock()
	na.cfg.ServerURL = url
	na.http.SetBaseURL(url)
}

// GetToken returns the current auth token (API key or JWT).
func (na *NA) GetToken() string {
	na.mu.RLock()
	defer na.mu.RUnlock()
	return na.cfg.Token
}

// SetToken updates the auth token on the HTTP client and in config.
func (na *NA) SetToken(token string) {
	na.mu.Lock()
	defer na.mu.Unlock()
	na.cfg.Token = token
	na.http.SetToken(token)
}

// SetActiveFamily caches the primary family for subsequent operations.
func (na *NA) SetActiveFamily(familyID, familyName string) {
	na.mu.Lock()
	defer na.mu.Unlock()
	na.activeFamilyID = familyID
	na.cfg.ActiveFamilyID = familyID
	na.cfg.ActiveFamilyName = familyName
}

// ActiveFamilyID returns the cached family ID (set via Init or SetActiveFamily).
func (na *NA) ActiveFamilyID() string {
	na.mu.RLock()
	defer na.mu.RUnlock()
	return na.activeFamilyID
}
