package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dezhishen/now-and-again/backend/pkg/types"
	"github.com/spf13/cobra"
)

var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "Manage tasks",
	Long:  `Create, list, and delete tasks in the active family.`,
}

var taskCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new task",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		schedule, _ := cmd.Flags().GetString("schedule")
		dataStr, _ := cmd.Flags().GetString("data")

		if name == "" || schedule == "" || dataStr == "" {
			return fmt.Errorf("--name, --schedule, --data are required")
		}

		var data interface{}
		if err := json.Unmarshal([]byte(dataStr), &data); err != nil {
			return fmt.Errorf("invalid --data JSON: %w", err)
		}

		t, err := na.CreateTask(context.Background(), &types.CreateTaskRequest{
			Task: types.Task{Name: name, ScheduleType: schedule, ScheduleData: data},
		})
		if err != nil {
			return err
		}
		fmt.Printf("Task created: %s (%s)\n", t.Name, t.ID[:8])
		return nil
	},
}

var taskListCmd = &cobra.Command{
	Use:   "list",
	Short: "List tasks in the active family",
	RunE: func(cmd *cobra.Command, args []string) error {
		tasks, err := na.ListTasks(context.Background())
		if err != nil {
			return err
		}
		if len(tasks) == 0 {
			fmt.Println("No tasks")
			return nil
		}
		for _, t := range tasks {
			s := "enabled"
			if !t.Enabled {
				s = "disabled"
			}
			fmt.Printf("  %s  %-20s  %-8s  %s\n", t.ID[:8], t.Name, t.ScheduleType, s)
		}
		return nil
	},
}

var taskDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a task",
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID, _ := cmd.Flags().GetString("id")
		if taskID == "" {
			return fmt.Errorf("--id is required")
		}
		return na.DeleteTask(context.Background(), taskID)
	},
}

func init() {
	taskCreateCmd.Flags().String("name", "", "Task name (required)")
	taskCreateCmd.Flags().String("schedule", "", "Schedule type: daily|weekly|monthly|interval|once")
	taskCreateCmd.Flags().String("data", "", "Schedule data JSON")
	taskDeleteCmd.Flags().String("id", "", "Task ID (required)")

	taskCmd.AddCommand(taskCreateCmd)
	taskCmd.AddCommand(taskListCmd)
	taskCmd.AddCommand(taskDeleteCmd)
}
