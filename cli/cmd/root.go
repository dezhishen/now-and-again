package cmd

import (
	"fmt"
	"os"

	"github.com/dezhishen/now-and-again/sdk"
	"github.com/spf13/cobra"
)

var (
	serverURL  string
	apiToken   string
	outputFmt  string
	configFile string

	na *sdk.NA
)

var rootCmd = &cobra.Command{
	Use:   "na",
	Short: "Now & Again — family chore management CLI",
	Long: `Now & Again CLI (na) — 家庭事务管理，一行命令搞定。

快速上手:
  na init -u <用户名> -p <密码>   一次性初始化
  na daily                         查看并处理今天的待办
  na template list                  查看可用模板
  na template use --code weekly_cleaning --params '{"area_name":"客厅"}'
  na task list                      查看任务`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initNA)

	rootCmd.PersistentFlags().StringVar(&serverURL, "server", "", "API server URL (overrides saved config)")
	rootCmd.PersistentFlags().StringVar(&apiToken, "token", "", "API auth token (overrides saved config)")
	rootCmd.PersistentFlags().StringVarP(&outputFmt, "output", "o", "table", "output format: table, json, yaml")
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "config file path (default: ~/.na.yaml)")

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(dailyCmd)
	rootCmd.AddCommand(familyCmd)
	rootCmd.AddCommand(taskCmd)
	rootCmd.AddCommand(todoCmd)
	rootCmd.AddCommand(templateCmd)
}

func initNA() {
	var err error
	if configFile != "" {
		na, err = sdk.NewWithConfigPath(configFile)
	} else {
		na, err = sdk.New()
	}
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
