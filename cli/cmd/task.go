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
	Short: "管理任务",
	Long:  `创建、列出和删除活跃家庭中的任务。`,
}

var taskCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "创建新任务",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := autoEnsureFamily(); err != nil {
			return err
		}
		name, _ := cmd.Flags().GetString("name")
		schedule, _ := cmd.Flags().GetString("schedule")
		dataStr, _ := cmd.Flags().GetString("data")
		if name == "" || schedule == "" || dataStr == "" {
			return fmt.Errorf("--name, --schedule, --data 不能为空")
		}
		var data interface{}
		if err := json.Unmarshal([]byte(dataStr), &data); err != nil {
			return fmt.Errorf("--data JSON格式错误: %w", err)
		}
		t, err := na.CreateTask(context.Background(), &types.CreateTaskRequest{
			Task: types.Task{Name: name, ScheduleType: schedule, ScheduleData: data},
		})
		if err != nil {
			return err
		}
		fmt.Printf("✅ 任务已创建: %s (%s)\n", t.Name, t.ID[:6])
		return nil
	},
}

var taskListCmd = &cobra.Command{
	Use:   "list",
	Short: "列出活跃家庭的任务（默认排除已归档和已禁用的任务）",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := autoEnsureFamily(); err != nil {
			return err
		}
		all, _ := cmd.Flags().GetBool("all")

		var tasks []types.Task
		var err error
		if all {
			tasks, err = na.ListTasksFiltered(context.Background(), true, true)
		} else {
			tasks, err = na.ListTasks(context.Background())
		}
		if err != nil {
			return err
		}

		if len(tasks) == 0 {
			fmt.Println("📭 暂无活跃任务")
			return nil
		}
		fmt.Printf("📋 任务列表 (%d项):\n\n", len(tasks))
		for _, t := range tasks {
			s := "✅"
			if t.Archived {
				s = "📦"
			} else if !t.Enabled {
				s = "⏸️"
			}
			fmt.Printf("  %s [%s] %-25s %-8s\n", s, t.ID[:6], t.Name, t.ScheduleType)
		}
		return nil
	},
}

var taskDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "删除任务",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := autoEnsureFamily(); err != nil {
			return err
		}
		taskID, _ := cmd.Flags().GetString("id")
		if taskID == "" {
			return fmt.Errorf("--id 不能为空")
		}
		return na.DeleteTask(context.Background(), taskID)
	},
}

func init() {
	taskCreateCmd.Flags().String("name", "", "任务名称")
	taskCreateCmd.Flags().String("schedule", "", "调度类型: daily|weekly|monthly|interval|once")
	taskCreateCmd.Flags().String("data", "", "调度数据 JSON，如 '{\"time\":\"09:00\"}'")
	taskDeleteCmd.Flags().String("id", "", "任务ID")

	taskListCmd.Flags().Bool("all", false, "显示全部任务（包括已归档和已禁用的）")

	taskCmd.AddCommand(taskCreateCmd)
	taskCmd.AddCommand(taskListCmd)
	taskCmd.AddCommand(taskDeleteCmd)
}
