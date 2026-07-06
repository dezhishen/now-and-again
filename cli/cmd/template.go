package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dezhishen/now-and-again/cli/internal/action"
	"github.com/spf13/cobra"
)

var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "任务模板",
	Long:  `查看可用模板，或从模板快速创建任务。`,
}

var templateListCmd = &cobra.Command{
	Use:   "list",
	Short: "查看可用模板",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := autoEnsureFamily(); err != nil {
			return err
		}
		kind, _ := cmd.Flags().GetString("kind")
		templates, err := na.ListTemplates(context.Background(), kind)
		if err != nil {
			return err
		}
		if len(templates) == 0 {
			fmt.Println("📭 暂无可用模板")
			return nil
		}
		fmt.Printf("📋 可用模板 (%d个):\n\n", len(templates))
		for _, t := range templates {
			fmt.Printf("  %s %-25s  %-8s  %s\n", t.Icon, t.TemplateCode, t.Kind, t.Name)
		}
		return nil
	},
}

var templateUseCmd = &cobra.Command{
	Use:   "use",
	Short: "从模板创建任务",
	Long: `渲染一个模板并用结果创建任务。

示例:
  na template use --code weekly_cleaning --params '{"area_name":"客厅"}'`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := autoEnsureFamily(); err != nil {
			return err
		}
		code, _ := cmd.Flags().GetString("code")
		paramsStr, _ := cmd.Flags().GetString("params")
		if code == "" {
			return fmt.Errorf("--code 不能为空")
		}
		var params map[string]interface{}
		if paramsStr != "" {
			if err := json.Unmarshal([]byte(paramsStr), &params); err != nil {
				return fmt.Errorf("--params JSON格式错误: %w", err)
			}
		}
		if params == nil {
			params = make(map[string]interface{})
		}
		t, err := na.CreateTaskFromTemplate(context.Background(), code, params)
		if err != nil {
			return err
		}
		action.Printf("✅ 已从模板创建: %s (%s)", t.Name, t.ID[:6])
		return nil
	},
}

func init() {
	templateListCmd.Flags().String("kind", "", "按类型筛选: simple|inspection|chain")
	templateUseCmd.Flags().String("code", "", "模板代码")
	templateUseCmd.Flags().String("params", "", "模板参数 JSON")

	templateCmd.AddCommand(templateListCmd)
	templateCmd.AddCommand(templateUseCmd)
}
