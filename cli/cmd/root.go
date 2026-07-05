package cmd

import (
	"fmt"
	"os"

	"github.com/dezhishen/now-and-again/sdk"
	"github.com/spf13/cobra"
)

var (
	serverURL string
	apiToken  string
	outputFmt string

	na *sdk.NA
)

var rootCmd = &cobra.Command{
	Use:   "na",
	Short: "Now & Again — family chore management CLI",
	Long: `Now & Again CLI (na) provides streamlined access to your family, tasks, and todos.

Quick start:
  na init -u <username> -p <password>       One-time setup (login → API key → save config)
  na template list                            List available templates
  na template use --code weekly_cleaning --params '{"area_name":"客厅"}'
  na task list                                List tasks in active family
  na todo list                                List pending todos
  na todo done --id <todo-id> --remark "已完成"

Output formats: table (default), json, yaml (--output / -o flag)`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initNA)

	rootCmd.PersistentFlags().StringVar(&serverURL, "server", "", "API server URL (overrides saved config)")
	rootCmd.PersistentFlags().StringVar(&apiToken, "token", "", "API auth token (overrides saved config)")
	rootCmd.PersistentFlags().StringVarP(&outputFmt, "output", "o", "table", "output format: table, json, yaml")

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(familyCmd)
	rootCmd.AddCommand(taskCmd)
	rootCmd.AddCommand(todoCmd)
	rootCmd.AddCommand(templateCmd)
}

func initNA() {
	var err error
	na, err = sdk.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\nrun 'na init' first\n", err)
		os.Exit(1)
	}

	if serverURL != "" {
		na.SetServerURL(serverURL)
	}
	if apiToken != "" {
		na.SetToken(apiToken)
	}
}
