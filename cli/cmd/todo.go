package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

var todoCmd = &cobra.Command{
	Use:   "todo",
	Short: "Manage todos",
	Long:  `List and complete pending todos in the active family.`,
}

var todoListCmd = &cobra.Command{
	Use:   "list",
	Short: "List pending todos",
	RunE: func(cmd *cobra.Command, args []string) error {
		todos, err := na.GetPendingTodos(context.Background())
		if err != nil {
			return err
		}
		if len(todos) == 0 {
			fmt.Println("No pending todos")
			return nil
		}
		for _, t := range todos {
			name := t.TaskID
			if t.Task != nil {
				name = t.Task.Name
			}
			fmt.Printf("  %s  %-20s  due: %s\n", t.ID[:8], name, t.DueDate.Format("2006-01-02 15:04"))
		}
		return nil
	},
}

var todoDoneCmd = &cobra.Command{
	Use:   "done",
	Short: "Mark a todo as done with optional remark",
	RunE: func(cmd *cobra.Command, args []string) error {
		todoID, _ := cmd.Flags().GetString("id")
		remark, _ := cmd.Flags().GetString("remark")
		if todoID == "" {
			return fmt.Errorf("--id is required")
		}
		if _, err := na.DoneTodo(context.Background(), todoID, remark); err != nil {
			return err
		}
		fmt.Printf("Todo %s marked as done\n", todoID[:8])
		return nil
	},
}

var todoSkipCmd = &cobra.Command{
	Use:   "skip",
	Short: "Skip a todo",
	RunE: func(cmd *cobra.Command, args []string) error {
		todoID, _ := cmd.Flags().GetString("id")
		if todoID == "" {
			return fmt.Errorf("--id is required")
		}
		if _, err := na.SkipTodo(context.Background(), todoID); err != nil {
			return err
		}
		fmt.Printf("Todo %s skipped\n", todoID[:8])
		return nil
	},
}

func init() {
	todoDoneCmd.Flags().String("id", "", "Todo ID (required)")
	todoDoneCmd.Flags().String("remark", "", "Completion remark")
	todoSkipCmd.Flags().String("id", "", "Todo ID (required)")

	todoCmd.AddCommand(todoListCmd)
	todoCmd.AddCommand(todoDoneCmd)
	todoCmd.AddCommand(todoSkipCmd)
}
