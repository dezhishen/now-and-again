package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// ─── task ────────────────────────────────────────────────────────

var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "Manage tasks (list, create, update, assign)",
}

var taskListCmd = &cobra.Command{
	Use:   "list",
	Short: "List tasks in a family",
	RunE: func(cmd *cobra.Command, args []string) error {
		familyID, _ := cmd.Flags().GetString("family-id")
		fmt.Printf("TODO: list tasks for family %s\n", familyID)
		return nil
	},
}

var taskCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new task",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("TODO: create task")
		return nil
	},
}

var taskUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update task status or fields",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("TODO: update task")
		return nil
	},
}

var taskAssignCmd = &cobra.Command{
	Use:   "assign",
	Short: "Assign users to a task",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("TODO: assign task")
		return nil
	},
}

func init() {
	taskListCmd.Flags().String("family-id", "", "Family ID (required)")
	taskListCmd.Flags().String("status", "", "Filter by status")
	taskListCmd.Flags().String("assignee", "", "Filter by assignee user ID")

	taskCreateCmd.Flags().String("family-id", "", "Family ID (required)")
	taskCreateCmd.Flags().String("title", "", "Task title (required)")
	taskCreateCmd.Flags().String("type", "chore_general", "Task type code")
	taskCreateCmd.Flags().String("priority", "medium", "Priority: low, medium, high, urgent")
	taskCreateCmd.Flags().String("due", "", "Due date (ISO 8601)")
	taskCreateCmd.Flags().String("description", "", "Task description")
	taskCreateCmd.Flags().StringSlice("assignee", nil, "Assignee user IDs (comma-separated)")

	taskUpdateCmd.Flags().String("task-id", "", "Task ID (required)")
	taskUpdateCmd.Flags().String("status", "", "New status: todo, in_progress, done")

	taskAssignCmd.Flags().String("task-id", "", "Task ID (required)")
	taskAssignCmd.Flags().StringSlice("user-id", nil, "User IDs to assign (comma-separated)")

	taskCmd.AddCommand(taskListCmd)
	taskCmd.AddCommand(taskCreateCmd)
	taskCmd.AddCommand(taskUpdateCmd)
	taskCmd.AddCommand(taskAssignCmd)
}
