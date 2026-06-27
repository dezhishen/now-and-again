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
	Short: "Now & Again — family chore management CLI",
	Long: `Now & Again CLI (na) provides script-friendly access to your family, tasks, and todos.

Authentication:
  Set NA_TOKEN environment variable or use --token flag.
  Get a token: na login -u <user> -p <pass>

Quick start:
  na family list                     List your families
  na family create --name "我的家"    Create a family
  na task list --family-id <id>      List tasks
  na task todo --family-id <id>      List pending todos
  na task create --family-id <id> --name "倒垃圾" --schedule daily --data '{"time":"09:00"}'

Output formats: table (default), json, yaml (--output / -o flag)

Environment:
  NA_TOKEN     API authentication token
  NA_SERVER    API server URL (default: http://localhost:8080)`,
	Example: `  na login -u admin -p secret
  na family list
  na task list --family-id abc123
  na task todo --family-id abc123
  na task done --id def456 --status done`,
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
