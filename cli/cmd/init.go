package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "One-time setup: login, create API key, and save config",
	Long: `Authenticates with username/password, creates an API key (or reuses existing),
auto-selects the first available family, and saves everything to ~/.na.yaml.

After init, all other commands use the saved configuration automatically.`,
	Example: `  na init -u admin -p 12345678
  na init --key na_key_xxxxxxxxxx
  na init -u admin -p 12345678 --server https://my-server.com`,
	RunE: func(cmd *cobra.Command, args []string) error {
		key, _ := cmd.Flags().GetString("key")
		username, _ := cmd.Flags().GetString("username")
		password, _ := cmd.Flags().GetString("password")

		if key != "" {
			if err := na.InitWithKey(key); err != nil {
				return fmt.Errorf("init failed: %w", err)
			}
			fmt.Println("✓ Initialized with API key")
			return nil
		}

		if username == "" || password == "" {
			return fmt.Errorf("--username and --password are required (or use --key)")
		}

		if err := na.Init(username, password); err != nil {
			return fmt.Errorf("init failed: %w", err)
		}
		fmt.Println("✓ Initialized successfully")
		cfg := na.Config()
		fmt.Printf("  Server: %s\n", cfg.ServerURL)
		if cfg.ActiveFamilyName != "" {
			fmt.Printf("  Family: %s\n", cfg.ActiveFamilyName)
		}
		return nil
	},
}

func init() {
	initCmd.Flags().StringP("username", "u", "", "Username")
	initCmd.Flags().StringP("password", "p", "", "Password")
	initCmd.Flags().String("key", "", "API key (skip login)")
}
