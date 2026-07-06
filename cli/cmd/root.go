package cmd

import (
	"fmt"
	"os"

	"github.com/dezhishen/now-and-again/cli/internal/action"
	"github.com/dezhishen/now-and-again/sdk"
	"github.com/spf13/cobra"
)

var (
	serverURL  string
	apiToken   string
	outputFmt  string
	configFile string
	actionID   string

	na *sdk.NA
)

var rootCmd = &cobra.Command{
	Use:   "na",
	Short: "Now & Again — family chore management CLI",
	Long: `Now & Again CLI (na) — 家庭事务管理，一行命令搞定。

每条命令自动生成唯一 actionId，输出带 [action:xxx] 前缀。
使用 --output json 获取机器可解析的结构化输出。

快速上手:
  na init -u <用户名> -p <密码>   一次性初始化
  na daily --done 1                直接完成第1项待办（非交互）
  na task list                     查看任务
  na task create --name "洗碗" --schedule daily --data '{"time":"19:00"}'
  na template list                 查看可用模板`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		action.Init(actionID, outputFmt)
		return nil
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initNA)

	rootCmd.PersistentFlags().StringVar(&serverURL, "server", "", "API server URL (overrides saved config)")
	rootCmd.PersistentFlags().StringVar(&apiToken, "token", "", "API auth token (overrides saved config)")
	rootCmd.PersistentFlags().StringVarP(&outputFmt, "output", "o", "text", "output format: text, json, yaml")
	rootCmd.PersistentFlags().StringVar(&actionID, "action-id", "", "action ID (auto-generated UUID if empty)")
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
