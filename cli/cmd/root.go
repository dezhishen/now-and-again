package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	cfgFile   string
	serverURL string
	apiToken  string
	apiKey    string
	outputFmt string // "table" | "json" | "yaml"
)

// rootCmd is the entry point for the CLI.
var rootCmd = &cobra.Command{
	Use:   "na",
	Short: "Now & Again — manage family chores from the terminal",
	Long: `Now & Again CLI (na) provides script-friendly access to your family tasks.

Examples:
  na task list --family-id <id>
  na task create --family-id <id> --title "Buy milk" --type chore_shopping
  na chain start --chain-id <id>
  na inspection start --family-id <id> --title "Monthly check"`,
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default $HOME/.na.yaml)")
	rootCmd.PersistentFlags().StringVar(&serverURL, "server", "http://localhost:8080", "API server URL")
	rootCmd.PersistentFlags().StringVar(&apiToken, "token", "", "API auth token (or set NA_TOKEN env)")
	rootCmd.PersistentFlags().StringVarP(&outputFmt, "output", "o", "table", "output format: table, json, yaml")

	// Subcommands are registered via init() in their respective files.
	rootCmd.AddCommand(taskCmd)
	rootCmd.AddCommand(chainCmd)
	rootCmd.AddCommand(inspectionCmd)
	rootCmd.AddCommand(familyCmd)
	rootCmd.AddCommand(loginCmd)
}

func initConfig() {
	// If token not set via flag, try env var.
	if apiToken == "" {
		apiToken = os.Getenv("NA_TOKEN")
	}
}
