package cmd

import (
	"github.com/dezhishen/now-and-again/cli/internal/action"
	"github.com/spf13/cobra"
)

// Version is set at build time via ldflags: -X github.com/dezhishen/now-and-again/cli/cmd.Version=v1.2.3
var Version = "dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "打印 CLI 版本",
	Long:  `显示当前 na 命令行工具的版本号。`,
	Run: func(cmd *cobra.Command, args []string) {
		action.Printf("na version %s", Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
