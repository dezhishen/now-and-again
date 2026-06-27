package cmd

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/dezhishen/now-and-again/backend/pkg/types"
	"github.com/spf13/cobra"
)

var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "Manage tasks and todos",
	Long:  `Create, list, enable/disable, and delete tasks. View and complete todos.`,
}

// ─── task create ─────────────────────────────────────────────────

var taskCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new task",
	Long: `Create a task that generates todos on a schedule.

Schedule types:
  once      One-time task (needs date+time in --data)
  daily     Every day at specified time
  weekly    Every week on specified days (1=Mon, 7=Sun)
  monthly   Every month on specified days (1-31)
  interval  Every N days`,
	Example: `  # Daily task at 9:00
  na task create --family-id xxx --name "倒垃圾" --schedule daily --data '{"time":"09:00"}'

  # One-time task for tomorrow
  na task create --family-id xxx --name "取快递" --schedule once --data '{"date":"2026-06-28","time":"18:00"}'

  # Weekly task on Mon/Wed/Fri
  na task create --family-id xxx --name "周报" --schedule weekly --data '{"days":[1,3,5],"time":"10:00"}'`,
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
			Task: types.TaskFields{
				Name:         name,
				ScheduleType: schedule,
				ScheduleData: data,
			},
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
	Use:     "list",
	Short:   "List tasks in a family",
	Example: "  na task list --family-id abc123",
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
	Use:     "todo",
	Short:   "List pending todos in a family",
	Example: "  na task todo --family-id abc123",
	RunE: func(cmd *cobra.Command, args []string) error {
		familyID, _ := cmd.Flags().GetString("family-id")
		if familyID == "" {
			return fmt.Errorf("--family-id is required")
		}
		todos, err := allClients.Task.ListTodosSimple(familyID, "pending")
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
		t, err := allClients.Task.CompleteTodoSimple(todoID, status)
		if err != nil {
			return err
		}
		fmt.Printf("Todo %s marked as %s\n", t.ID[:8], t.Status)
		return nil
	},
}

// ─── task enable/disable ─────────────────────────────────────────

var taskToggleCmd = &cobra.Command{
	Use:     "toggle",
	Short:   "Enable or disable a task",
	Example: "  na task toggle --id abc123 --enable false",
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID, _ := cmd.Flags().GetString("id")
		enableStr, _ := cmd.Flags().GetString("enable")
		if taskID == "" || enableStr == "" {
			return fmt.Errorf("--id and --enable (true|false) are required")
		}
		enable, _ := strconv.ParseBool(enableStr)
		t, err := allClients.Task.Update(taskID, &types.UpdateTaskRequest{
			Task: &types.UpdateTaskFields{Enabled: &enable},
		})
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
