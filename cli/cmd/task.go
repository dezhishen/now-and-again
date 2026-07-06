package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/dezhishen/now-and-again/backend/pkg/types"
	"github.com/dezhishen/now-and-again/cli/internal/action"
	"github.com/dezhishen/now-and-again/cli/internal/resolver"
	"github.com/spf13/cobra"
)

var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "管理任务",
	Long: `创建、查看、更新和删除活跃家庭中的任务。

所有命令均支持通过 --name 指定任务名称（而非短ID），CLI 会自动查找对应的任务。
也支持通过 --id 使用完整 UUID 或短前缀。`,
}

// ─── task create ──────────────────────────────────────────────────

var taskCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "创建新任务",
	Long: `在活跃家庭中创建新任务。

调度类型 (--schedule) 及对应的 --data 参数：

  daily       {"time":"09:00"}                         每天指定时间
  weekly      {"time":"09:00","days":[1,3,5]}          每周一三五 (1=周一,7=周日)
  monthly     {"time":"09:00","days":[1,15]}            每月1号和15号
  yearly      {"time":"09:00","day":6,"month":7}       每年7月6日
  interval    {"days":3}                                每3天一次
  once        {"date":"2026-07-10","time":"14:00"}     仅一次

示例:
  na task create --name "洗碗" --schedule daily --data '{"time":"19:00"}'
  na task create --name "大扫除" --schedule weekly --data '{"time":"09:00","days":[6]}' --group 大人 --location 厨房`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := autoEnsureFamily(); err != nil {
			return err
		}
		ctx := context.Background()

		name, _ := cmd.Flags().GetString("name")
		schedule, _ := cmd.Flags().GetString("schedule")
		dataStr, _ := cmd.Flags().GetString("data")
		groupName, _ := cmd.Flags().GetString("group")
		locationName, _ := cmd.Flags().GetString("location")
		kind, _ := cmd.Flags().GetString("kind")

		if name == "" || schedule == "" || dataStr == "" {
			return fmt.Errorf("--name, --schedule, --data 不能为空")
		}

		var data interface{}
		if err := json.Unmarshal([]byte(dataStr), &data); err != nil {
			return fmt.Errorf("--data JSON格式错误: %w", err)
		}

		req := &types.CreateTaskRequest{
			Task: types.Task{
				Name:         name,
				ScheduleType: schedule,
				ScheduleData: data,
				Kind:         kind,
			},
		}

		// Resolve group name → ID
		cache := resolver.NewCache()
		if groupName != "" {
			gid, err := cache.ResolveGroupID(ctx, na, groupName)
			if err != nil {
				return fmt.Errorf("--group: %w", err)
			}
			req.Task.GroupID = gid
		}

		// Resolve location name → ID
		if locationName != "" {
			lid, err := cache.ResolveLocationID(ctx, na, locationName)
			if err != nil {
				return fmt.Errorf("--location: %w", err)
			}
			req.Task.LocationID = lid
		}

		t, err := na.CreateTask(ctx, req)
		if err != nil {
			return err
		}

		// Show friendly names in output
		groupLabel := ""
		if groupName != "" {
			groupLabel = fmt.Sprintf(" 小组: %s", groupName)
		}
		locLabel := ""
		if locationName != "" {
			locLabel = fmt.Sprintf(" 地址: %s", locationName)
		}
		action.Printf("✅ 任务已创建: %s (%s)%s%s", t.Name, t.ScheduleType, groupLabel, locLabel)
		return nil
	},
}

// ─── task list ────────────────────────────────────────────────────

var taskListCmd = &cobra.Command{
	Use:   "list",
	Short: "列出活跃家庭的任务（默认排除已归档和已禁用的任务）",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := autoEnsureFamily(); err != nil {
			return err
		}
		ctx := context.Background()
		all, _ := cmd.Flags().GetBool("all")

		var tasks []types.Task
		var err error
		if all {
			tasks, err = na.ListTasksFiltered(ctx, true, true, "")
		} else {
			tasks, err = na.ListTasks(ctx)
		}
		if err != nil {
			return err
		}

		if len(tasks) == 0 {
			action.Println("📭 暂无活跃任务")
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

		fmt.Printf("📋 任务列表 (%d项):\n\n", len(tasks))
		for _, t := range tasks {
			s := "✅"
			if t.Archived {
				s = "📦"
			} else if !t.Enabled {
				s = "⏸️"
			}

			extra := ""
			if gn, ok := groupMap[t.GroupID]; ok {
				extra += fmt.Sprintf(" [%s]", gn)
			}
			if ln, ok := locMap[t.LocationID]; ok {
				extra += fmt.Sprintf(" @%s", ln)
			}

			fmt.Printf("  %s %-30s %-8s%s\n", s, t.Name, t.ScheduleType, extra)
		}

		fmt.Println("\n💡 使用 na task info --name \"任务名\" 查看详情")
		fmt.Println("💡 使用 na task delete --name \"任务名\" 删除任务")
		return nil
	},
}

// ─── task info ────────────────────────────────────────────────────

var taskInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "查看任务详情（支持名称或ID）",
	Example: `  na task info --name "洗碗"
  na task info --id abc123`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := autoEnsureFamily(); err != nil {
			return err
		}
		ctx := context.Background()
		cache := resolver.NewCache()

		input, err := resolveTaskInput(cmd)
		if err != nil {
			return err
		}

		task, err := cache.ResolveTask(ctx, na, input)
		if err != nil {
			return err
		}

		// Resolve group/location names for display
		groupName := "-"
		if task.GroupID != "" {
			groups, _ := na.ListGroups(ctx, "")
			for _, g := range groups {
				if g.ID == task.GroupID {
					groupName = g.Name
					break
				}
			}
		}
		locName := "-"
		if task.LocationID != "" {
			locs, _ := na.ListLocations(ctx, "")
			for _, l := range locs {
				if l.ID == task.LocationID {
					locName = l.Name
					break
				}
			}
		}

		schedData, _ := json.MarshalIndent(task.ScheduleData, "  ", "  ")

		status := "✅ 启用"
		if task.Archived {
			status = "📦 已归档"
		} else if !task.Enabled {
			status = "⏸️  已禁用"
		}

		fmt.Printf(`
📋 任务详情
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  名称:       %s
  ID:         %s
  类型:       %s (%s)
  状态:       %s
  小组:       %s
  地址:       %s
  调度类型:   %s
  调度参数:
%s
  创建时间:   %s
  更新时间:   %s
  上次待办:   %s
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
`,
			task.Name,
			task.ID,
			task.Kind, task.ScheduleType,
			status,
			groupName,
			locName,
			task.ScheduleType,
			string(schedData),
			na.FormatTime(task.CreatedAt, "2006-01-02 15:04"),
			na.FormatTime(task.UpdatedAt, "2006-01-02 15:04"),
			formatLastTodo(task.LastTodoAt),
		)
		return nil
	},
}

// ─── task update ──────────────────────────────────────────────────

var taskUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "更新任务属性（支持名称或ID定位）",
	Long: `更新任务的名称、调度参数、小组或地址。

示例:
  na task update --name "洗碗" --new-name "晚餐洗碗"
  na task update --name "大扫除" --schedule weekly --data '{"time":"10:00","days":[6,7]}'
  na task update --name "浇花" --group "大人" --location "阳台"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := autoEnsureFamily(); err != nil {
			return err
		}
		ctx := context.Background()
		cache := resolver.NewCache()

		input, err := resolveTaskInput(cmd)
		if err != nil {
			return err
		}

		task, err := cache.ResolveTask(ctx, na, input)
		if err != nil {
			return err
		}

		// Build update request from flags that are explicitly set
		updateTask := &types.Task{}
		hasUpdate := false

		if cmd.Flags().Changed("new-name") {
			n, _ := cmd.Flags().GetString("new-name")
			updateTask.Name = n
			hasUpdate = true
		}
		if cmd.Flags().Changed("schedule") {
			s, _ := cmd.Flags().GetString("schedule")
			updateTask.ScheduleType = s
			hasUpdate = true
		}
		if cmd.Flags().Changed("data") {
			d, _ := cmd.Flags().GetString("data")
			var sd interface{}
			if err := json.Unmarshal([]byte(d), &sd); err != nil {
				return fmt.Errorf("--data JSON格式错误: %w", err)
			}
			updateTask.ScheduleData = sd
			hasUpdate = true
		}
		if cmd.Flags().Changed("group") {
			gn, _ := cmd.Flags().GetString("group")
			if gn == "" {
				updateTask.GroupID = "" // clear group
			} else {
				gid, err := cache.ResolveGroupID(ctx, na, gn)
				if err != nil {
					return fmt.Errorf("--group: %w", err)
				}
				updateTask.GroupID = gid
			}
			hasUpdate = true
		}
		if cmd.Flags().Changed("location") {
			ln, _ := cmd.Flags().GetString("location")
			if ln == "" {
				updateTask.LocationID = "" // clear location
			} else {
				lid, err := cache.ResolveLocationID(ctx, na, ln)
				if err != nil {
					return fmt.Errorf("--location: %w", err)
				}
				updateTask.LocationID = lid
			}
			hasUpdate = true
		}

		if !hasUpdate {
			return fmt.Errorf("请指定至少一项要更新的属性: --new-name, --schedule, --data, --group, --location")
		}

		_, err = na.UpdateTask(ctx, task.ID, &types.UpdateTaskRequest{Task: updateTask})
		if err != nil {
			return err
		}
		action.Printf("✅ 已更新任务: %s", task.Name)
		return nil
	},
}

// ─── task delete ──────────────────────────────────────────────────

var taskDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "删除任务（支持名称或ID）",
	Long: `删除任务。使用 -y 或 --output json 可跳过交互式确认。

AI/脚本用法:
  na task delete --name "洗碗" -y
  na task delete --name "洗碗" --output json`,
	Example: `  na task delete --name "洗碗" -y    # 非交互删除
  na task delete --name "洗碗" --output json  # JSON 输出，自动跳过确认`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := autoEnsureFamily(); err != nil {
			return err
		}
		ctx := context.Background()
		cache := resolver.NewCache()

		input, err := resolveTaskInput(cmd)
		if err != nil {
			return err
		}

		task, err := cache.ResolveTask(ctx, na, input)
		if err != nil {
			return err
		}

		skipConfirm, _ := cmd.Flags().GetBool("yes")
		// In structured output mode (AI usage), skip interactive confirmation
		if !skipConfirm && action.OutputFormat() == "text" {
			fmt.Printf("⚠️  确认删除任务 %q ? (y/N): ", task.Name)
			var response string
			fmt.Scanln(&response)
			if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
				action.Println("已取消")
				return nil
			}
		}

		if err := na.DeleteTask(ctx, task.ID); err != nil {
			return err
		}
		action.Printf("🗑️  已删除任务: %s", task.Name)
		return nil
	},
}

// ─── task trigger ─────────────────────────────────────────────────

var taskTriggerCmd = &cobra.Command{
	Use:     "trigger",
	Short:   "手动触发任务（立即生成一条待办，支持名称或ID）",
	Example: `  na task trigger --name "洗碗"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := autoEnsureFamily(); err != nil {
			return err
		}
		ctx := context.Background()
		cache := resolver.NewCache()

		input, err := resolveTaskInput(cmd)
		if err != nil {
			return err
		}

		task, err := cache.ResolveTask(ctx, na, input)
		if err != nil {
			return err
		}

		if err := na.TriggerTask(ctx, task.ID); err != nil {
			return err
		}
		action.Printf("⚡ 已触发任务: %s", task.Name)
		return nil
	},
}

// ─── Flag registration ────────────────────────────────────────────

func init() {
	// task create
	taskCreateCmd.Flags().String("name", "", "任务名称（必填）")
	taskCreateCmd.Flags().String("schedule", "", "调度类型: daily|weekly|monthly|yearly|interval|once")
	taskCreateCmd.Flags().String("data", "", "调度数据 JSON（见 --help 示例）")
	taskCreateCmd.Flags().String("group", "", "分配小组名称")
	taskCreateCmd.Flags().String("location", "", "分配地址名称")
	taskCreateCmd.Flags().String("kind", "simple", "任务类型: simple|inspection|chain")

	// task list
	taskListCmd.Flags().Bool("all", false, "显示全部任务（包括已归档和已禁用的）")

	// task info
	taskInfoCmd.Flags().String("name", "", "任务名称")
	taskInfoCmd.Flags().String("id", "", "任务ID（完整UUID或≥3位前缀）")

	// task update
	taskUpdateCmd.Flags().String("name", "", "要更新的任务名称")
	taskUpdateCmd.Flags().String("id", "", "要更新的任务ID（完整UUID或≥3位前缀）")
	taskUpdateCmd.Flags().String("new-name", "", "新的任务名称")
	taskUpdateCmd.Flags().String("schedule", "", "新的调度类型")
	taskUpdateCmd.Flags().String("data", "", "新的调度数据 JSON")
	taskUpdateCmd.Flags().String("group", "", "新的小组名称（传空字符串清除）")
	taskUpdateCmd.Flags().String("location", "", "新的地址名称（传空字符串清除）")

	// task delete
	taskDeleteCmd.Flags().String("name", "", "任务名称")
	taskDeleteCmd.Flags().String("id", "", "任务ID（完整UUID或≥3位前缀）")
	taskDeleteCmd.Flags().BoolP("yes", "y", false, "跳过确认提示")

	// task trigger
	taskTriggerCmd.Flags().String("name", "", "任务名称")
	taskTriggerCmd.Flags().String("id", "", "任务ID（完整UUID或≥3位前缀）")

	taskCmd.AddCommand(taskCreateCmd)
	taskCmd.AddCommand(taskListCmd)
	taskCmd.AddCommand(taskInfoCmd)
	taskCmd.AddCommand(taskUpdateCmd)
	taskCmd.AddCommand(taskDeleteCmd)
	taskCmd.AddCommand(taskTriggerCmd)
}

// ─── Helpers ──────────────────────────────────────────────────────

// resolveTaskInput reads --name or --id flag (name takes priority) and returns a non-empty input.
func resolveTaskInput(cmd *cobra.Command) (string, error) {
	name, _ := cmd.Flags().GetString("name")
	if name != "" {
		return name, nil
	}
	id, _ := cmd.Flags().GetString("id")
	if id != "" {
		return id, nil
	}
	return "", fmt.Errorf("请指定 --name 或 --id")
}

func formatLastTodo(t *time.Time) string {
	if t == nil || t.IsZero() {
		return "-"
	}
	return t.Format("2006-01-02 15:04")
}
