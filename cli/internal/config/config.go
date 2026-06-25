// Package config manages CLI configuration stored in ~/.na.yaml.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config represents the CLI configuration.
type Config struct {
	ServerURL string `mapstructure:"server_url"`
	Token     string `mapstructure:"token"`
}

// Load reads configuration from file and environment.
func Load(cfgFile string) (*Config, error) {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("home dir: %w", err)
		}
		viper.AddConfigPath(home)
		viper.SetConfigName(".na")
		viper.SetConfigType("yaml")
	}

	viper.SetEnvPrefix("NA")
	viper.AutomaticEnv()

	// Defaults
	viper.SetDefault("server_url", "http://localhost:8080")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("read config: %w", err)
		}
		// Config file not found is OK; use defaults + env.
	}

	cfg := &Config{
		ServerURL: viper.GetString("server_url"),
		Token:     viper.GetString("token"),
	}
	return cfg, nil
}

// Save writes the configuration to disk.
func Save(cfgFile string, cfg *Config) error {
	if cfgFile == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("home dir: %w", err)
		}
		cfgFile = filepath.Join(home, ".na.yaml")
	}

	viper.Set("server_url", cfg.ServerURL)
	viper.Set("token", cfg.Token)

	if err := os.MkdirAll(filepath.Dir(cfgFile), 0755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

	return viper.WriteConfigAs(cfgFile)
}
