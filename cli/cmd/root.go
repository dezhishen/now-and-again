package cmd

import (
	"os"

	"github.com/dezhishen/now-and-again/cli/internal/client"
	"github.com/spf13/cobra"
)

var (
	cfgFile   string
	serverURL string
	apiToken  string
	outputFmt string

	// allClients is initialized in initConfig and used by subcommands.
	allClients *client.AllClients
)

var rootCmd = &cobra.Command{
	Use:   "na",
	Short: "Now & Again — manage family chores from the terminal",
	Long:  `Now & Again CLI (na) provides script-friendly access to your family and account.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default $HOME/.na.yaml)")
	rootCmd.PersistentFlags().StringVar(&serverURL, "server", "http://localhost:8080", "API server URL")
	rootCmd.PersistentFlags().StringVar(&apiToken, "token", "", "API auth token (or set NA_TOKEN env)")
	rootCmd.PersistentFlags().StringVarP(&outputFmt, "output", "o", "table", "output format: table, json, yaml")

	rootCmd.AddCommand(familyCmd)
	rootCmd.AddCommand(loginCmd)
}

func initConfig() {
	if apiToken == "" {
		apiToken = os.Getenv("NA_TOKEN")
	}

	httpClient := client.NewHTTPClient(serverURL, apiToken)
	allClients = client.NewAllClients(httpClient)
}
