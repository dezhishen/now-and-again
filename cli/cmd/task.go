package cmd

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/dezhishen/now-and-again/shared/types"
	"github.com/spf13/cobra"
)

var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "Manage tasks and todos",
}

// ─── task create ─────────────────────────────────────────────────

var taskCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a task template",
	Example: `  na task create --family-id xxx --name "倒垃圾" --schedule daily --data '{"time":"09:00"}'
  na task create --family-id xxx --name "周报" --schedule weekly --data '{"days":[1,5],"time":"10:00"}'`,
	RunE: func(cmd *cobra.Command, args []string) error {
		familyID, _ := cmd.Flags().GetString("family-id")
		name, _ := cmd.Flags().GetString("name")
		schedule, _ := cmd.Flags().GetString("schedule")
		dataStr, _ := cmd.Flags().GetString("data")

		if familyID == "" || name == "" || schedule == "" || dataStr == "" {
			return fmt.Errorf("--family-id, --name, --schedule, --data are required")
		}

		var data interface{}
		if err := json.Unmarshal([]byte(dataStr), &data); err != nil {
			return fmt.Errorf("invalid --data JSON: %w", err)
		}

		t, err := allClients.Task.Create(familyID, &types.CreateTaskRequest{
			Name:         name,
			ScheduleType: schedule,
			ScheduleData: data,
		})
		if err != nil {
			return err
		}
		fmt.Printf("Task created: %s (%s)\n", t.Name, t.ID[:8])
		return nil
	},
}

// ─── task list ───────────────────────────────────────────────────

var taskListCmd = &cobra.Command{
	Use:   "list",
	Short: "List tasks for a family",
	RunE: func(cmd *cobra.Command, args []string) error {
		familyID, _ := cmd.Flags().GetString("family-id")
		if familyID == "" {
			return fmt.Errorf("--family-id is required")
		}
		tasks, err := allClients.Task.List(familyID)
		if err != nil {
			return err
		}
		if len(tasks) == 0 {
			fmt.Println("No tasks")
			return nil
		}
		for _, t := range tasks {
			status := "enabled"
			if !t.Enabled {
				status = "disabled"
			}
			fmt.Printf("  %s  %-20s  %-8s  %s\n", t.ID[:8], t.Name, t.ScheduleType, status)
		}
		return nil
	},
}

// ─── task todo ───────────────────────────────────────────────────

var taskTodoCmd = &cobra.Command{
	Use:   "todo",
	Short: "List pending todos for a family",
	RunE: func(cmd *cobra.Command, args []string) error {
		familyID, _ := cmd.Flags().GetString("family-id")
		if familyID == "" {
			return fmt.Errorf("--family-id is required")
		}
		todos, err := allClients.Task.ListTodos(familyID, "pending")
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

// ─── task done ───────────────────────────────────────────────────

var taskDoneCmd = &cobra.Command{
	Use:     "done",
	Short:   "Mark a todo as done or skipped",
	Example: "  na task done --id xxx --status done",
	RunE: func(cmd *cobra.Command, args []string) error {
		todoID, _ := cmd.Flags().GetString("id")
		status, _ := cmd.Flags().GetString("status")
		if todoID == "" || status == "" {
			return fmt.Errorf("--id and --status (done|skipped) are required")
		}
		t, err := allClients.Task.CompleteTodo(todoID, status)
		if err != nil {
			return err
		}
		fmt.Printf("Todo %s marked as %s\n", t.ID[:8], t.Status)
		return nil
	},
}

// ─── task enable/disable ─────────────────────────────────────────

var taskToggleCmd = &cobra.Command{
	Use:   "toggle",
	Short: "Enable or disable a task",
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID, _ := cmd.Flags().GetString("id")
		enableStr, _ := cmd.Flags().GetString("enable")
		if taskID == "" || enableStr == "" {
			return fmt.Errorf("--id and --enable (true|false) are required")
		}
		enable, _ := strconv.ParseBool(enableStr)
		t, err := allClients.Task.Update(taskID, &types.UpdateTaskRequest{Enabled: &enable})
		if err != nil {
			return err
		}
		status := "disabled"
		if t.Enabled {
			status = "enabled"
		}
		fmt.Printf("Task %s %s\n", t.Name, status)
		return nil
	},
}

// ─── task delete ─────────────────────────────────────────────────

var taskDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a task template",
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID, _ := cmd.Flags().GetString("id")
		if taskID == "" {
			return fmt.Errorf("--id is required")
		}
		if err := allClients.Task.Delete(taskID); err != nil {
			return err
		}
		fmt.Println("Task deleted")
		return nil
	},
}

func init() {
	taskCreateCmd.Flags().String("family-id", "", "Family ID (required)")
	taskCreateCmd.Flags().String("name", "", "Task name (required)")
	taskCreateCmd.Flags().String("schedule", "", "Schedule type: daily|weekly|monthly|interval")
	taskCreateCmd.Flags().String("data", "", "Schedule data JSON")

	taskListCmd.Flags().String("family-id", "", "Family ID (required)")

	taskTodoCmd.Flags().String("family-id", "", "Family ID (required)")

	taskDoneCmd.Flags().String("id", "", "Todo ID (required)")
	taskDoneCmd.Flags().String("status", "", "Status: done|skipped")

	taskToggleCmd.Flags().String("id", "", "Task ID (required)")
	taskToggleCmd.Flags().String("enable", "", "true or false")

	taskDeleteCmd.Flags().String("id", "", "Task ID (required)")

	taskCmd.AddCommand(taskCreateCmd)
	taskCmd.AddCommand(taskListCmd)
	taskCmd.AddCommand(taskTodoCmd)
	taskCmd.AddCommand(taskDoneCmd)
	taskCmd.AddCommand(taskToggleCmd)
	taskCmd.AddCommand(taskDeleteCmd)
	rootCmd.AddCommand(taskCmd)
}
