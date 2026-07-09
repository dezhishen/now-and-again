package cmd

import (
	"context"
	"fmt"

	"github.com/dezhishen/now-and-again/cli/internal/action"
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
		ctx := context.Background()
		todos, err := na.GetPendingTodos(ctx)
		if err != nil {
			return err
		}
		if len(todos) == 0 {
			action.Println("🎉 暂无待办")
			return nil
		}

		// Pre-load groups and locations for display
		groups, _ := na.ListGroups(ctx, "")
		locations, _ := na.ListLocations(ctx, "")
		groupMap := make(map[string]string)
		for _, g := range groups {
			groupMap[g.ID] = g.Name
		}
		locMap := make(map[string]string)
		for _, l := range locations {
			locMap[l.ID] = l.Name
		}

		fmt.Printf("📋 待办 (%d项):\n\n", len(todos))
		for i, t := range todos {
			name := t.TaskName
			if t.Task != nil {
				name = t.Task.Name
			}
			due := ""
			if !t.DueDate.IsZero() {
				due = "  ⏰ " + na.FormatTime(t.DueDate, "01-02 15:04")
			}
			extra := ""
			if t.Task != nil {
				if gn, ok := groupMap[t.Task.GroupID]; ok {
					extra += fmt.Sprintf(" [%s]", gn)
				}
			}
			if ln, ok := locMap[t.LocationID]; ok {
				extra += fmt.Sprintf(" @%s", ln)
			}
			fmt.Printf("  %2d. %s%s%s\n", i+1, name, due, extra)
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
		action.Printf("✅ 已完成: %s", name)
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
		action.Println("⏭️  已跳过")
		return nil
	},
}

var todoInterruptCmd = &cobra.Command{
	Use:     "interrupt",
	Short:   "中断待办（标记为已中断，支持短ID）",
	Example: "  na todo interrupt --id abc",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := autoEnsureFamily(); err != nil {
			return err
		}
		todoID, _ := cmd.Flags().GetString("id")
		if todoID == "" {
			return fmt.Errorf("--id 不能为空")
		}
		if _, err := na.CompleteTodo(context.Background(), todoID, "interrupted", ""); err != nil {
			return fmt.Errorf("中断失败: %w", err)
		}
		action.Println("⏹️  已中断")
		return nil
	},
}

func init() {
	todoDoneCmd.Flags().String("id", "", "待办ID（完整UUID或至少3位前缀）")
	todoDoneCmd.Flags().String("remark", "", "完成备注")
	todoSkipCmd.Flags().String("id", "", "待办ID（完整UUID或至少3位前缀）")
	todoInterruptCmd.Flags().String("id", "", "待办ID（完整UUID或至少3位前缀）")

	todoCmd.AddCommand(todoListCmd)
	todoCmd.AddCommand(todoDoneCmd)
	todoCmd.AddCommand(todoSkipCmd)
	todoCmd.AddCommand(todoInterruptCmd)
}
