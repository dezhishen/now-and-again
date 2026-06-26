package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	cfgFile   string
	serverURL string
	apiToken  string
	outputFmt string
)

// rootCmd is the entry point for the CLI.
var rootCmd = &cobra.Command{
	Use:   "na",
	Short: "Now & Again — manage family chores from the terminal",
	Long:  `Now & Again CLI (na) provides script-friendly access to your family and account.`,
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

	// Subcommands
	rootCmd.AddCommand(familyCmd)
	rootCmd.AddCommand(loginCmd)
}

func initConfig() {
	// If token not set via flag, try env var.
	if apiToken == "" {
		apiToken = os.Getenv("NA_TOKEN")
	}
}
