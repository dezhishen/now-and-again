package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

var todoCmd = &cobra.Command{
	Use:   "todo",
	Short: "管理待办",
	Long:  `列出和完成活跃家庭的待办事项。支持短ID（3位以上即可）。`,
}

var todoListCmd = &cobra.Command{
	Use:   "list",
	Short: "列出待办",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := autoEnsureFamily(); err != nil {
			return err
		}
		todos, err := na.GetPendingTodos(context.Background())
		if err != nil {
			return err
		}
		if len(todos) == 0 {
			fmt.Println("🎉 暂无待办")
			return nil
		}
		fmt.Printf("📋 待办 (%d项):\n\n", len(todos))
		for i, t := range todos {
			name := t.TaskName
			if t.Task != nil {
				name = t.Task.Name
			}
			due := ""
			if !t.DueDate.IsZero() {
				due = "  ⏰ " + t.DueDate.Format("01-02 15:04")
			}
			fmt.Printf("  %2d. [%s] %s%s\n", i+1, t.ID[:6], name, due)
		}
		fmt.Println("\n💡 使用 na todo done --id abc123 完成（支持短ID，至少3位）")
		fmt.Println("💡 或使用 na daily 交互式处理")
		return nil
	},
}

var todoDoneCmd = &cobra.Command{
	Use:     "done",
	Short:   "完成待办（支持短ID）",
	Example: "  na todo done --id abc      # 使用6位前缀即□",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := autoEnsureFamily(); err != nil {
			return err
		}
		todoID, _ := cmd.Flags().GetString("id")
		remark, _ := cmd.Flags().GetString("remark")
		if todoID == "" {
			return fmt.Errorf("--id 不能为空（支持完整UUID或至少3位前缀）")
		}
		t, err := na.DoneTodo(context.Background(), todoID, remark)
		if err != nil {
			return fmt.Errorf("完成失败: %w", err)
		}
		name := t.TaskName
		if t.Task != nil {
			name = t.Task.Name
		}
		fmt.Printf("✅ 已完成: %s", name)
		if remark != "" {
			fmt.Printf(" (%s)", remark)
		}
		fmt.Println()
		return nil
	},
}

var todoSkipCmd = &cobra.Command{
	Use:   "skip",
	Short: "跳过待办（支持短ID）",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := autoEnsureFamily(); err != nil {
			return err
		}
		todoID, _ := cmd.Flags().GetString("id")
		if todoID == "" {
			return fmt.Errorf("--id 不能为空")
		}
		if _, err := na.SkipTodo(context.Background(), todoID); err != nil {
			return err
		}
		fmt.Println("⏭️  已跳过")
		return nil
	},
}

func init() {
	todoDoneCmd.Flags().String("id", "", "待办ID（完整UUID或至少3位前缀）")
	todoDoneCmd.Flags().String("remark", "", "完成备注")
	todoSkipCmd.Flags().String("id", "", "待办ID（完整UUID或至少3位前缀）")

	todoCmd.AddCommand(todoListCmd)
	todoCmd.AddCommand(todoDoneCmd)
	todoCmd.AddCommand(todoSkipCmd)
}
