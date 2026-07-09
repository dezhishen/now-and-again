package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/dezhishen/now-and-again/backend/pkg/types"
	"github.com/dezhishen/now-and-again/cli/internal/action"
	"github.com/dezhishen/now-and-again/cli/internal/resolver"
	"github.com/dezhishen/now-and-again/cli/internal/state"
	"github.com/google/uuid"
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
	Short: "创建新任务（支持引导式交互）",
	Long: `在活跃家庭中创建新任务。

引导式交互（无参数时自动进入）:
  na task create

快速模式:
  na task create --name "洗碗" --schedule daily --data '{"time":"19:00"}'
  na task create --name "大扫除" --schedule weekly --data '{"time":"09:00","days":[6]}' --group 大人 --location 厨房
  na task create --name "安全检查" --kind inspection --schedule daily --data '{"time":"08:00"}' --extra '{"check_items":[...]}'
  na task create --name "搬家流程" --kind chain --schedule once --data '{"date":"2026-08-01","time":"09:00"}' --extra '{"steps":[...]}'

调度类型 (--schedule) 及对应的 --data 参数：
  daily       {"time":"09:00"}
  weekly      {"time":"09:00","days":[1,3,5]}    (1=周一,7=周日)
  monthly     {"time":"09:00","days":[1,15]}
  yearly      {"time":"09:00","day":6,"month":7}
  interval    {"days":3}
  once        {"date":"2026-07-10","time":"14:00"}

任务类型 (--kind):
  simple      简单定时任务（默认）
  inspection  巡检任务（需配合 --extra 指定 check_items）
  chain       任务链（需配合 --extra 指定 steps）`,
	RunE: runTaskCreate,
}

func runTaskCreate(cmd *cobra.Command, args []string) error {
	if err := autoEnsureFamily(); err != nil {
		return err
	}
	ctx := context.Background()
	actID := string(action.CurrentID())
	cache := resolver.NewCache()

	name, _ := cmd.Flags().GetString("name")
	schedule, _ := cmd.Flags().GetString("schedule")
	dataStr, _ := cmd.Flags().GetString("data")
	groupName, _ := cmd.Flags().GetString("group")
	locationName, _ := cmd.Flags().GetString("location")
	groupID, _ := cmd.Flags().GetString("group-id")
	locationID, _ := cmd.Flags().GetString("location-id")
	kind, _ := cmd.Flags().GetString("kind")
	extraStr, _ := cmd.Flags().GetString("extra")
	yes, _ := cmd.Flags().GetBool("yes")

	// ── Interactive mode when required flags are missing ────────
	if name == "" || schedule == "" || dataStr == "" {
		if name != "" || schedule != "" || dataStr != "" {
			return fmt.Errorf("--name, --schedule, --data 三者必须同时提供，或全部省略进入引导模式")
		}
		return runTaskCreateInteractive(ctx, cache, actID, kind, groupName, groupID, locationName, locationID, yes)
	}

	// ── Fast path: flags provided ─────────────────────────────
	var data interface{}
	if err := json.Unmarshal([]byte(dataStr), &data); err != nil {
		return fmt.Errorf("--data JSON格式错误: %w", err)
	}

	// Parse extra data for complex tasks (chain steps / inspection check items)
	var extra interface{}
	if extraStr != "" {
		if err := json.Unmarshal([]byte(extraStr), &extra); err != nil {
			return fmt.Errorf("--extra JSON格式错误: %w", err)
		}
	}

	req := &types.CreateTaskRequest{
		Task: types.Task{
			Name:         name,
			ScheduleType: schedule,
			ScheduleData: data,
			Kind:         kind,
		},
		Extra: extra,
	}

	stateArgs := map[string]string{
		"name": name, "schedule": schedule, "data": dataStr, "kind": kind,
		"group": groupName, "location": locationName,
		"group-id": groupID, "location-id": locationID,
	}

	if groupID != "" {
		req.Task.GroupID = groupID
	} else if groupName != "" {
		gid, err := cache.ResolveGroupIDInteractive(ctx, na, groupName, actID, "task create", stateArgs)
		if err != nil {
			return fmt.Errorf("--group: %w", err)
		}
		req.Task.GroupID = gid
	}

	if locationID != "" {
		req.Task.LocationID = locationID
	} else if locationName != "" {
		lid, err := cache.ResolveLocationIDInteractive(ctx, na, locationName, actID, "task create", stateArgs)
		if err != nil {
			return fmt.Errorf("--location: %w", err)
		}
		req.Task.LocationID = lid
	}

	t, err := na.CreateTask(ctx, req)
	if err != nil {
		return err
	}
	state.Delete(actID)

	action.Printf("✅ 任务已创建: %s (%s)%s%s",
		t.Name, t.ScheduleType,
		groupLabel(groupName, groupID),
		locLabel(locationName, locationID))
	return nil
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
	Short: "查看任务详情（支持名称或ID，自动显示嵌套结构）",
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

		// Build kind-specific info
		kindLabel := task.Kind
		switch task.Kind {
		case "chain":
			kindLabel = "🔗 任务链"
		case "inspection":
			kindLabel = "🔍 巡检"
		default:
			kindLabel = "📌 简单"
		}

		fmt.Printf(`
📋 任务详情
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  名称:       %s
  ID:         %s
  类型:       %s
  状态:       %s
  小组:       %s
  地址:       %s
  调度类型:   %s
  调度参数:
%s
  展示摘要:   %s
  创建时间:   %s
  更新时间:   %s
  上次待办:   %s
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
`,
			task.Name,
			task.ID,
			kindLabel,
			status,
			groupName,
			locName,
			task.ScheduleType,
			string(schedData),
			task.DisplaySummary,
			na.FormatTime(task.CreatedAt, "2006-01-02 15:04"),
			na.FormatTime(task.UpdatedAt, "2006-01-02 15:04"),
			formatLastTodo(task.LastTodoAt),
		)

		// Show nested structure for complex tasks
		if task.Kind == "chain" || task.Kind == "inspection" {
			printComplexTaskExtra(ctx, task)
		}

		return nil
	},
}

// printComplexTaskExtra fetches and displays the extra data for chain/inspection tasks.
func printComplexTaskExtra(ctx context.Context, task *types.Task) {
	twe, err := na.GetTaskWithExtra(ctx, task.ID)
	if err != nil {
		action.Printf("⚠️  无法获取嵌套详情: %v", err)
		return
	}
	if twe == nil || twe.Extra == nil {
		return
	}

	extraMap, ok := twe.Extra.(map[string]interface{})
	if !ok {
		return
	}

	switch task.Kind {
	case "chain":
		printChainSteps(extraMap)
	case "inspection":
		printCheckItems(extraMap)
	}
}

// printChainSteps displays chain step details from extra data.
func printChainSteps(extra map[string]interface{}) {
	stepsRaw, ok := extra["steps"]
	if !ok {
		return
	}
	steps, ok := stepsRaw.([]interface{})
	if !ok || len(steps) == 0 {
		return
	}

	fmt.Printf("\n🔗 任务链步骤 (%d):\n", len(steps))
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	for i, s := range steps {
		step, ok := s.(map[string]interface{})
		if !ok {
			continue
		}
		name, _ := step["name"].(string)
		kind, _ := step["kind"].(string)
		if kind == "" {
			kind = "simple"
		}
		kindIcon := "📌"
		switch kind {
		case "inspection":
			kindIcon = "🔍"
		case "chain":
			kindIcon = "🔗"
		}
		fmt.Printf("  %d. %s %s  [%s]\n", i+1, kindIcon, name, kind)
	}
	fmt.Println()
}

// printCheckItems displays inspection check items from extra data.
func printCheckItems(extra map[string]interface{}) {
	itemsRaw, ok := extra["check_items"]
	if !ok {
		return
	}
	items, ok := itemsRaw.([]interface{})
	if !ok || len(items) == 0 {
		return
	}

	fmt.Printf("\n🔍 巡检项目 (%d):\n", len(items))
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	for i, ci := range items {
		item, ok := ci.(map[string]interface{})
		if !ok {
			continue
		}
		name, _ := item["name"].(string)
		fmt.Printf("  %d. %s\n", i+1, name)

		branchesRaw, ok := item["branches"]
		if !ok {
			continue
		}
		branches, ok := branchesRaw.([]interface{})
		if !ok {
			continue
		}
		for _, b := range branches {
			branch, ok := b.(map[string]interface{})
			if !ok {
				continue
			}
			bName, _ := branch["name"].(string)
			createTodo, _ := branch["create_todo"].(bool)
			mark := "  "
			if createTodo {
				mark = "📋"
			}
			fmt.Printf("     %s %s", mark, bName)
			if createTodo {
				if bt, ok := branch["branch_task"].(map[string]interface{}); ok {
					if btTask, ok := bt["task"].(map[string]interface{}); ok {
						if tname, ok := btTask["name"].(string); ok {
							fmt.Printf(" → 创建任务: %s", tname)
						}
					}
				}
			}
			fmt.Println()
		}
	}
	fmt.Println()
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
		actID := string(action.CurrentID())
		cache := resolver.NewCache()

		input, err := resolveTaskInput(cmd)
		if err != nil {
			return err
		}

		task, err := cache.ResolveTask(ctx, na, input)
		if err != nil {
			return err
		}

		// Build args snapshot for state persistence.
		stateArgs := map[string]string{"name": input}
		if cmd.Flags().Changed("new-name") {
			n, _ := cmd.Flags().GetString("new-name")
			stateArgs["new-name"] = n
		}
		if cmd.Flags().Changed("schedule") {
			s, _ := cmd.Flags().GetString("schedule")
			stateArgs["schedule"] = s
		}
		if cmd.Flags().Changed("data") {
			d, _ := cmd.Flags().GetString("data")
			stateArgs["data"] = d
		}
		if cmd.Flags().Changed("group") {
			stateArgs["group"], _ = cmd.Flags().GetString("group")
		}
		if cmd.Flags().Changed("group-id") {
			stateArgs["group-id"], _ = cmd.Flags().GetString("group-id")
		}
		if cmd.Flags().Changed("location") {
			stateArgs["location"], _ = cmd.Flags().GetString("location")
		}
		if cmd.Flags().Changed("location-id") {
			stateArgs["location-id"], _ = cmd.Flags().GetString("location-id")
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
		if cmd.Flags().Changed("group") || cmd.Flags().Changed("group-id") {
			if cmd.Flags().Changed("group-id") {
				gid, _ := cmd.Flags().GetString("group-id")
				if gid == "" {
					updateTask.GroupID = "" // clear group
				} else {
					updateTask.GroupID = gid
				}
			} else {
				gn, _ := cmd.Flags().GetString("group")
				if gn == "" {
					updateTask.GroupID = "" // clear group
				} else {
					gid, err := cache.ResolveGroupIDInteractive(ctx, na, gn, actID, "task update", stateArgs)
					if err != nil {
						return fmt.Errorf("--group: %w", err)
					}
					updateTask.GroupID = gid
				}
			}
			hasUpdate = true
		}
		if cmd.Flags().Changed("location") || cmd.Flags().Changed("location-id") {
			if cmd.Flags().Changed("location-id") {
				lid, _ := cmd.Flags().GetString("location-id")
				if lid == "" {
					updateTask.LocationID = "" // clear location
				} else {
					updateTask.LocationID = lid
				}
			} else {
				ln, _ := cmd.Flags().GetString("location")
				if ln == "" {
					updateTask.LocationID = "" // clear location
				} else {
					lid, err := cache.ResolveLocationIDInteractive(ctx, na, ln, actID, "task update", stateArgs)
					if err != nil {
						return fmt.Errorf("--location: %w", err)
					}
					updateTask.LocationID = lid
				}
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
		// Clean up state file on success.
		state.Delete(actID)
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

// ─── task enable ──────────────────────────────────────────────────

var taskEnableCmd = &cobra.Command{
	Use:   "enable",
	Short: "启用任务（支持名称或ID）",
	Example: `  na task enable --name "洗碗"
  na task enable --id abc123`,
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

		if _, err := na.SetTaskEnabled(ctx, task.ID, true); err != nil {
			return err
		}
		action.Printf("✅ 已启用任务: %s", task.Name)
		return nil
	},
}

// ─── task disable ─────────────────────────────────────────────────

var taskDisableCmd = &cobra.Command{
	Use:   "disable",
	Short: "禁用任务（支持名称或ID）",
	Example: `  na task disable --name "洗碗"
  na task disable --id abc123`,
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

		if _, err := na.SetTaskEnabled(ctx, task.ID, false); err != nil {
			return err
		}
		action.Printf("⏸️  已禁用任务: %s", task.Name)
		return nil
	},
}

// ─── task logs ────────────────────────────────────────────────────

var taskLogsCmd = &cobra.Command{
	Use:   "logs",
	Short: "查看任务日志（支持名称或ID）",
	Long: `列出指定任务的执行日志，包括系统自动记录和用户手动记录的日志。

示例:
  na task logs --name "洗碗"
  na task logs --id abc123 --limit 20
  na task logs --name "大扫除" --user`,
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

		limit, _ := cmd.Flags().GetInt("limit")
		userOnly, _ := cmd.Flags().GetBool("user")

		logs, err := na.Task.ListTaskLogs(ctx, uuid.MustParse(task.ID), limit, userOnly)
		if err != nil {
			return err
		}

		if len(logs) == 0 {
			action.Println("📭 暂无日志")
			return nil
		}

		fmt.Printf("📋 任务 %q 的日志 (%d条):\n\n", task.Name, len(logs))
		for _, l := range logs {
			ts := na.FormatTime(l.CreatedAt, "01-02 15:04:05")
			lvl := l.LogType
			fmt.Printf("  [%s] %-5s %s\n", ts, lvl, l.Message)
		}
		fmt.Println()
		return nil
	},
}

// ─── task children ────────────────────────────────────────────────

var taskChildrenCmd = &cobra.Command{
	Use:   "children",
	Short: "列出任务的子任务（适用于任务链/巡检等复杂任务）",
	Long: `列出指定任务的直接子任务。对于任务链，显示每个步骤对应的子任务；
对于巡检任务，显示各检查项分支创建的子任务。

示例:
  na task children --name "大扫除"
  na task children --id abc123`,
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

		// List all tasks and filter by parent_task_id
		allTasks, err := na.ListTasksFiltered(ctx, true, true, "")
		if err != nil {
			return fmt.Errorf("获取任务列表失败: %w", err)
		}

		var children []types.Task
		for _, t := range allTasks {
			if t.ParentTaskID == task.ID {
				children = append(children, t)
			}
		}

		if len(children) == 0 {
			fmt.Printf("\n📭 任务 %q 没有子任务\n", task.Name)
			if task.Kind == "simple" {
				fmt.Println("💡 简单任务没有子任务。使用 na task create --kind chain 创建任务链")
			} else {
				fmt.Println("💡 子任务在触发待办后才会创建。使用 na task trigger --name \"" + task.Name + "\" 触发")
			}
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

		fmt.Printf("\n📋 %q 的子任务 (%d):\n\n", task.Name, len(children))
		for i, child := range children {
			s := "✅"
			if child.Archived {
				s = "📦"
			} else if !child.Enabled {
				s = "⏸️"
			}

			kindIcon := "📌"
			switch child.Kind {
			case "chain":
				kindIcon = "🔗"
			case "inspection":
				kindIcon = "🔍"
			}

			extra := ""
			if gn, ok := groupMap[child.GroupID]; ok {
				extra += fmt.Sprintf(" [%s]", gn)
			}
			if ln, ok := locMap[child.LocationID]; ok {
				extra += fmt.Sprintf(" @%s", ln)
			}
			if child.DisplaySummary != "" {
				extra += fmt.Sprintf("  %s", child.DisplaySummary)
			}

			fmt.Printf("  %2d. %s %s %-25s %s%s\n", i+1, s, kindIcon, child.Name, child.ScheduleType, extra)
		}
		fmt.Println()
		return nil
	},
}

// ─── Flag registration ────────────────────────────────────────────

// ─── Interactive task creation ───────────────────────────────────
//
// AI 友好设计原则:
//   - 每条输出带 [action: xxx] 前缀，AI 可通过 action-id 跟踪
//   - 输入提示统一 "→ QUESTION: ... [default: ...]" 格式
//   - 复杂任务分步引导，每步确认后进入下一步
//   - 最终确认展示完整嵌套结构，支持回退修改

func runTaskCreateInteractive(ctx context.Context, cache *resolver.Cache, actID, kind, groupName, groupID, locationName, locationID string, yes bool) error {
	action.Println("📝 引导式任务创建")
	action.Printf("💡 提示: 所有输入回车为空则使用默认值或跳过")

	// ── STEP 1: 任务名称 ─────────────────────────────────────
	name := promptAI("任务名称", "例如: 洗碗、安全检查", "")
	if name == "" {
		return fmt.Errorf("任务名称不能为空")
	}
	action.Printf("✅ 名称: %s", name)

	// ── STEP 2: 任务类型 ─────────────────────────────────────
	if kind == "" || kind == "simple" {
		action.Println("")
		action.Println("任务类型:")
		action.Println("  1. 📌 simple     — 简单定时任务")
		action.Println("  2. 🔍 inspection — 巡检任务 (含检查项+分支+分支任务)")
		action.Println("  3. 🔗 chain      — 任务链 (多步骤，每步可独立配置)")
		k := promptAI("选择类型", "1/2/3", "1")
		switch k {
		case "2":
			kind = "inspection"
		case "3":
			kind = "chain"
		default:
			kind = "simple"
		}
	}
	action.Printf("✅ 类型: %s", kindLabel(kind))

	// ── STEP 3: 调度配置 ─────────────────────────────────────
	action.Println("")
	action.Println("调度类型:")
	action.Println("  1. daily    — 每天")
	action.Println("  2. weekly   — 每周")
	action.Println("  3. monthly  — 每月")
	action.Println("  4. yearly   — 每年")
	action.Println("  5. interval — 间隔N天")
	action.Println("  6. once     — 仅一次")
	schedChoice := promptAI("选择调度", "1/2/3/4/5/6", "1")

	schedule, schedData := buildScheduleData(schedChoice)
	if schedule == "" {
		return fmt.Errorf("无效的调度类型")
	}
	action.Printf("✅ 调度: %s", schedule)

	// ── STEP 4: 复杂任务配置 (inspection / chain) ────────────
	var extra interface{}
	if kind == "inspection" {
		extra = buildInspectionExtra()
		if extra == nil {
			action.Println("⚠️  未配置检查项，将创建空巡检任务")
		}
	} else if kind == "chain" {
		extra = buildChainExtra()
		if extra == nil {
			action.Println("⚠️  未配置步骤，将创建空任务链")
		}
	}

	// ── STEP 5: 小组 (可选) ──────────────────────────────────
	var gid string
	if groupID != "" {
		gid = groupID
	} else if groupName != "" {
		stateArgs := map[string]string{"name": name, "kind": kind, "schedule": schedule, "group": groupName}
		var err error
		gid, err = cache.ResolveGroupIDInteractive(ctx, na, groupName, actID, "task create", stateArgs)
		if err != nil {
			return err
		}
	} else if !yes {
		gInput := promptAI("分配到小组 (可选)", "回车跳过", "")
		if gInput != "" {
			stateArgs := map[string]string{"name": name, "kind": kind, "schedule": schedule, "group": gInput}
			var err error
			gid, err = cache.ResolveGroupIDInteractive(ctx, na, gInput, actID, "task create", stateArgs)
			if err != nil {
				return err
			}
		}
	}
	if gid != "" {
		action.Printf("✅ 小组: %s", gid[:min(8, len(gid))])
	}

	// ── STEP 6: 地址 (可选) ──────────────────────────────────
	var lid string
	if locationID != "" {
		lid = locationID
	} else if locationName != "" {
		stateArgs := map[string]string{"name": name, "kind": kind, "schedule": schedule, "location": locationName}
		var err error
		lid, err = cache.ResolveLocationIDInteractive(ctx, na, locationName, actID, "task create", stateArgs)
		if err != nil {
			return err
		}
	} else if !yes {
		lInput := promptAI("关联地址 (可选)", "回车跳过", "")
		if lInput != "" {
			stateArgs := map[string]string{"name": name, "kind": kind, "schedule": schedule, "location": lInput}
			var err error
			lid, err = cache.ResolveLocationIDInteractive(ctx, na, lInput, actID, "task create", stateArgs)
			if err != nil {
				return err
			}
		}
	}
	if lid != "" {
		action.Printf("✅ 地址: %s", lid[:min(8, len(lid))])
	}

	// ── STEP 7: 最终确认 (增强版) ────────────────────────────
	schedJSON, _ := json.MarshalIndent(schedData, "  ", "  ")
	if !yes {
		fmt.Println()
		fmt.Println("┌──────────────────────────────────────────────┐")
		action.Printf("│ 📋 确认任务摘要                                │")
		fmt.Println("│                                              │")
		fmt.Printf("│  名称:   %-36s │\n", truncateStr(name, 36))
		fmt.Printf("│  类型:   %-36s │\n", kindLabel(kind))
		fmt.Printf("│  调度:   %-36s │\n", schedule)
		schedLines := strings.Split(string(schedJSON), "\n")
		for _, line := range schedLines {
			fmt.Printf("│          %-36s │\n", truncateStr(line, 36))
		}
		if gid != "" {
			fmt.Printf("│  小组:   %-36s │\n", gid[:min(8, len(gid))])
		}
		if lid != "" {
			fmt.Printf("│  地址:   %-36s │\n", lid[:min(8, len(lid))])
		}
		fmt.Println("│                                              │")

		// 展示复杂任务嵌套结构
		if kind == "chain" {
			printChainSummary(extra)
		} else if kind == "inspection" {
			printInspectionSummary(extra)
		}

		fmt.Println("└──────────────────────────────────────────────┘")

		confirm := promptAI("确认创建?", "Y/n/s=改步骤/c=改检查项/q=取消", "y")
		switch strings.ToLower(confirm) {
		case "q", "quit", "exit":
			state.CleanupIfDone(actID)
			action.Println("已取消")
			return nil
		case "y", "yes", "":
			// proceed
		default:
			action.Println("已取消")
			return nil
		}
	}

	// ── STEP 8: 创建 ────────────────────────────────────────
	req := &types.CreateTaskRequest{
		Task: types.Task{
			Name:         name,
			ScheduleType: schedule,
			ScheduleData: schedData,
			Kind:         kind,
			GroupID:      gid,
			LocationID:   lid,
		},
		Extra: extra,
	}

	t, err := na.CreateTask(ctx, req)
	if err != nil {
		return err
	}
	state.Delete(actID)

	action.Printf("✅ 任务已创建: %s (%s) [%s]", t.Name, t.ScheduleType, kind)
	return nil
}

// ─── AI-friendly prompt helper ────────────────────────────────────
// promptAI 输出统一格式: → QUESTION: <label> [default: <defaultVal>]
func promptAI(label, hint, defaultVal string) string {
	if defaultVal != "" {
		fmt.Printf("→ %s [default: %s]: ", label, defaultVal)
	} else if hint != "" {
		fmt.Printf("→ %s (%s): ", label, hint)
	} else {
		fmt.Printf("→ %s: ", label)
	}
	var input string
	fmt.Scanln(&input)
	input = strings.TrimSpace(input)
	if input == "" {
		return defaultVal
	}
	return input
}

// ─── buildScheduleData ────────────────────────────────────────────

func buildScheduleData(choice string) (scheduleType string, data map[string]interface{}) {
	data = make(map[string]interface{})
	switch choice {
	case "1", "daily":
		scheduleType = "daily"
		t := promptAI("时间", "HH:MM", "09:00")
		data["time"] = t
	case "2", "weekly":
		scheduleType = "weekly"
		t := promptAI("时间", "HH:MM", "09:00")
		data["time"] = t
		daysStr := promptAI("星期几", "1-7逗号分隔,1=周一", "1,3,5")
		var days []int
		for _, s := range strings.Split(daysStr, ",") {
			s = strings.TrimSpace(s)
			if n, err := strconv.Atoi(s); err == nil && n >= 1 && n <= 7 {
				days = append(days, n)
			}
		}
		data["days"] = days
	case "3", "monthly":
		scheduleType = "monthly"
		t := promptAI("时间", "HH:MM", "09:00")
		data["time"] = t
		daysStr := promptAI("几号", "1-31逗号分隔", "1,15")
		var days []int
		for _, s := range strings.Split(daysStr, ",") {
			s = strings.TrimSpace(s)
			if n, err := strconv.Atoi(s); err == nil && n >= 1 && n <= 31 {
				days = append(days, n)
			}
		}
		data["days"] = days
	case "4", "yearly":
		scheduleType = "yearly"
		t := promptAI("时间", "HH:MM", "09:00")
		data["time"] = t
		mStr := promptAI("月份", "1-12", "7")
		m, _ := strconv.Atoi(mStr)
		data["month"] = m
		dStr := promptAI("日期", "1-31", "6")
		d, _ := strconv.Atoi(dStr)
		data["day"] = d
	case "5", "interval":
		scheduleType = "interval"
		dStr := promptAI("间隔天数", "", "3")
		d, _ := strconv.Atoi(dStr)
		data["days"] = d
	case "6", "once":
		scheduleType = "once"
		dateStr := promptAI("日期", "YYYY-MM-DD", time.Now().Format("2006-01-02"))
		data["date"] = dateStr
		tStr := promptAI("时间", "HH:MM", "09:00")
		data["time"] = tStr
	default:
		return "", nil
	}
	return
}

// ─── Inspection interactive builder ───────────────────────────────
//
// 流程:
//   1. 逐个询问检查项
//   2. 每个检查项 → 询问分支名称
//   3. 每个分支 → 询问是否创建任务 → 任务类型 → 任务调度
//   4. 如果分支任务是 chain → 递归引导步骤
//   5. 检查项配置完毕 → 确认进入下一个检查项

func buildInspectionExtra() interface{} {
	action.Println("")
	action.Println("🔍 巡检项配置")
	action.Println("   为每个检查项配置结果分支，不足/异常分支可自动创建任务")

	var items []map[string]interface{}
	for i := 1; ; i++ {
		itemName := promptAI(fmt.Sprintf("检查项%d名称", i), "空行结束", "")
		if itemName == "" {
			if i == 1 {
				action.Println("   (跳过，不配置检查项)")
			}
			break
		}

		action.Printf("  📋 检查项: %s", itemName)
		action.Println("     配置结果分支 (逗号分隔多个结果):")

		branchesStr := promptAI(fmt.Sprintf("    %s的结果选项", itemName), "逗号分隔", "正常,不足")
		var branches []map[string]interface{}
		branchNames := strings.Split(branchesStr, ",")
		for _, bn := range branchNames {
			bn = strings.TrimSpace(bn)
			if bn == "" {
				continue
			}

			action.Printf("      ── 分支 %q ──", bn)
			createTodoStr := promptAI(fmt.Sprintf("      %q是否创建任务?", bn), "y/N", "n")
			createTodo := strings.ToLower(createTodoStr) == "y" || strings.ToLower(createTodoStr) == "yes"

			branch := map[string]interface{}{
				"name":        bn,
				"create_todo": createTodo,
			}

			if createTodo {
				branch["branch_task"] = buildBranchTask(bn)
			}
			branches = append(branches, branch)
		}

		item := map[string]interface{}{
			"name":     itemName,
			"branches": branches,
		}
		items = append(items, item)

		todoBranches := 0
		for _, b := range branches {
			if b["create_todo"].(bool) {
				todoBranches++
			}
		}
		action.Printf("  ✅ 检查项 %q: %d个分支, %d个创建任务", itemName, len(branches), todoBranches)
	}
	if len(items) == 0 {
		return nil
	}
	action.Printf("✅ 巡检项配置完成: %d 项", len(items))
	return map[string]interface{}{"check_items": items}
}

// buildBranchTask 引导配置分支任务 (支持 simple / inspection / chain)
func buildBranchTask(branchName string) map[string]interface{} {
	action.Println("        📋 分支任务配置:")

	// 任务名称
	taskName := promptAI("        任务名称", fmt.Sprintf("修复%s", branchName), fmt.Sprintf("处理%s", branchName))

	// 任务类型
	action.Println("        任务类型:")
	action.Println("          1. simple     — 简单任务")
	action.Println("          2. inspection — 巡检任务")
	action.Println("          3. chain      — 任务链 (多步骤)")
	taskKind := promptAI("        选择类型", "1/2/3", "1")
	kind := "simple"
	var subExtra interface{}
	switch taskKind {
	case "2":
		kind = "inspection"
		subExtra = buildInspectionExtraSimple()
	case "3":
		kind = "chain"
		subExtra = buildChainExtraSimple()
	}

	// 调度
	schedule, schedData := buildSimpleSchedule()

	task := map[string]interface{}{
		"name":          taskName,
		"kind":          kind,
		"schedule_type": schedule,
		"schedule_data": schedData,
	}

	result := map[string]interface{}{
		"task": task,
	}
	if subExtra != nil {
		result["extra"] = subExtra
	}
	return result
}

// buildSimpleSchedule 引导配置简单调度（用于分支任务/链步骤）
func buildSimpleSchedule() (string, map[string]interface{}) {
	action.Println("        调度:")
	action.Println("          1. daily    2. weekly    3. once")
	c := promptAI("        选择", "1/2/3", "3")
	switch c {
	case "1":
		t := promptAI("        时间", "HH:MM", "09:00")
		return "daily", map[string]interface{}{"time": t}
	case "2":
		t := promptAI("        时间", "HH:MM", "09:00")
		ds := promptAI("        星期几", "1-7逗号分隔", "1,3,5")
		var days []int
		for _, s := range strings.Split(ds, ",") {
			s = strings.TrimSpace(s)
			if n, err := strconv.Atoi(s); err == nil && n >= 1 && n <= 7 {
				days = append(days, n)
			}
		}
		return "weekly", map[string]interface{}{"time": t, "days": days}
	default:
		d := promptAI("        日期", "YYYY-MM-DD", time.Now().AddDate(0, 0, 1).Format("2006-01-02"))
		t := promptAI("        时间", "HH:MM", "09:00")
		return "once", map[string]interface{}{"date": d, "time": t}
	}
}

// ─── Chain interactive builder ────────────────────────────────────
//
// 流程:
//   1. 逐个询问步骤
//   2. 每个步骤 → 询问类型 (simple/inspection/chain)
//   3. 基于类型引导配置:
//      simple   → 调度
//      inspection → 检查项 + 分支任务
//      chain    → 递归子步骤 (最多1层)
//   4. 确认当前步骤 → 进入下一个步骤
//   5. 所有步骤配置完毕 → 返回

func buildChainExtra() interface{} {
	action.Println("")
	action.Println("🔗 任务链步骤配置")
	action.Println("   每步可独立设置类型和调度，链式执行")

	var steps []map[string]interface{}
	for i := 1; ; i++ {
		action.Println("")
		stepName := promptAI(fmt.Sprintf("步骤%d名称", i), "空行结束", "")
		if stepName == "" {
			if i == 1 {
				action.Println("   (跳过，不配置步骤)")
			}
			break
		}

		// 简化交互: 用编号选择类型
		action.Println("  步骤类型:")
		action.Println("    1. 📌 simple     — 简单任务")
		action.Println("    2. 🔍 inspection — 巡检任务")
		action.Println("    3. 🔗 chain      — 子任务链")
		kindChoice := promptAI("  选择类型", "1/2/3", "1")

		var stepKind string
		var stepExtra interface{}
		switch kindChoice {
		case "2":
			stepKind = "inspection"
			stepExtra = buildInspectionExtraSimple()
		case "3":
			stepKind = "chain"
			stepExtra = buildChainExtraSimple()
		default:
			stepKind = "simple"
		}

		// 步骤调度
		schedule, schedData := buildSimpleSchedule()

		step := map[string]interface{}{
			"name":          stepName,
			"kind":          stepKind,
			"schedule_type": schedule,
			"schedule_data": schedData,
		}
		if stepExtra != nil {
			step["extra"] = stepExtra
		}

		steps = append(steps, step)
		action.Printf("  ✅ 步骤%d: %s %s (%s)", i, kindIcon(stepKind), stepName, schedule)
	}
	if len(steps) == 0 {
		return nil
	}
	action.Printf("✅ 任务链配置完成: %d 步", len(steps))
	return map[string]interface{}{"steps": steps}
}

// buildInspectionExtraSimple 简化版巡检配置（链步骤内使用）
func buildInspectionExtraSimple() interface{} {
	action.Println("    🔍 子巡检项配置:")
	var items []map[string]interface{}
	for i := 1; i <= 5; i++ {
		itemName := promptAI(fmt.Sprintf("    检查项%d名称", i), "空行跳过", "")
		if itemName == "" {
			break
		}

		branchesStr := promptAI(fmt.Sprintf("    %q的结果选项", itemName), "逗号分隔", "合格,不合格")
		var branches []map[string]interface{}
		for _, bn := range strings.Split(branchesStr, ",") {
			bn = strings.TrimSpace(bn)
			if bn == "" {
				continue
			}
			createTodoStr := promptAI(fmt.Sprintf("    %q是否创建任务?", bn), "y/N", "n")
			createTodo := strings.ToLower(createTodoStr) == "y" || strings.ToLower(createTodoStr) == "yes"

			branch := map[string]interface{}{
				"name":        bn,
				"create_todo": createTodo,
			}
			if createTodo {
				branch["branch_task"] = buildBranchTask(bn)
			}
			branches = append(branches, branch)
		}
		items = append(items, map[string]interface{}{
			"name":     itemName,
			"branches": branches,
		})
	}
	if len(items) == 0 {
		return nil
	}
	return map[string]interface{}{"check_items": items}
}

// buildChainExtraSimple 简化版任务链配置（分支任务内使用）
func buildChainExtraSimple() interface{} {
	action.Println("    🔗 子任务链步骤:")
	var steps []map[string]interface{}
	for i := 1; i <= 10; i++ {
		stepName := promptAI(fmt.Sprintf("    子步骤%d名称", i), "空行结束", "")
		if stepName == "" {
			break
		}
		kindChoice := promptAI("    类型", "1=simple 2=inspection 3=chain", "1")
		var stepKind string
		switch kindChoice {
		case "2":
			stepKind = "inspection"
		case "3":
			stepKind = "chain"
		default:
			stepKind = "simple"
		}
		steps = append(steps, map[string]interface{}{
			"name": stepName,
			"kind": stepKind,
		})
	}
	if len(steps) == 0 {
		return nil
	}
	return map[string]interface{}{"steps": steps}
}

// ─── Summary printers (for confirm step) ──────────────────────────

func printChainSummary(extra interface{}) {
	if extra == nil {
		return
	}
	em, ok := extra.(map[string]interface{})
	if !ok {
		return
	}
	stepsRaw, ok := em["steps"]
	if !ok {
		return
	}
	steps, ok := stepsRaw.([]interface{})
	if !ok || len(steps) == 0 {
		return
	}
	fmt.Printf("│  步骤 (%d):                                     │\n", len(steps))
	for i, s := range steps {
		step, ok := s.(map[string]interface{})
		if !ok {
			continue
		}
		sn, _ := step["name"].(string)
		sk, _ := step["kind"].(string)
		if sk == "" {
			sk = "simple"
		}
		st, _ := step["schedule_type"].(string)
		fmt.Printf("│    %d. %s %-31s │\n", i+1, kindIcon(sk), truncateStr(sn, 28))
		if st != "" {
			fmt.Printf("│       调度: %-33s │\n", truncateStr(st, 33))
		}
	}
}

func printInspectionSummary(extra interface{}) {
	if extra == nil {
		return
	}
	em, ok := extra.(map[string]interface{})
	if !ok {
		return
	}
	itemsRaw, ok := em["check_items"]
	if !ok {
		return
	}
	items, ok := itemsRaw.([]interface{})
	if !ok || len(items) == 0 {
		return
	}
	fmt.Printf("│  检查项 (%d):                                    │\n", len(items))
	for _, ci := range items {
		item, ok := ci.(map[string]interface{})
		if !ok {
			continue
		}
		in, _ := item["name"].(string)
		fmt.Printf("│    📋 %-37s │\n", truncateStr(in, 37))
		branches, _ := item["branches"].([]interface{})
		for _, b := range branches {
			br, ok := b.(map[string]interface{})
			if !ok {
				continue
			}
			bn, _ := br["name"].(string)
			ct, _ := br["create_todo"].(bool)
			if ct {
				bt, _ := br["branch_task"].(map[string]interface{})
				tn := ""
				if btt, ok := bt["task"].(map[string]interface{}); ok {
					tn, _ = btt["name"].(string)
				}
				if tn != "" {
					fmt.Printf("│      %s → 📋 %-31s │\n", truncateStr(bn, 6), truncateStr(tn, 31))
				} else {
					fmt.Printf("│      %s → 📋 (创建任务)%-22s │\n", truncateStr(bn, 6), "")
				}
			} else {
				fmt.Printf("│      %s (无任务)%-29s │\n", truncateStr(bn, 8), "")
			}
		}
	}
}

// ─── Label / icon helpers ─────────────────────────────────────────

func kindLabel(kind string) string {
	switch kind {
	case "chain":
		return "🔗 任务链 (chain)"
	case "inspection":
		return "🔍 巡检 (inspection)"
	default:
		return "📌 简单 (simple)"
	}
}

func kindIcon(kind string) string {
	switch kind {
	case "chain":
		return "🔗"
	case "inspection":
		return "🔍"
	default:
		return "📌"
	}
}

func truncateStr(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen-1]) + "…"
}

// groupLabel returns a human-readable group label.
func groupLabel(name, id string) string {
	if name != "" {
		return fmt.Sprintf(" 小组: %s", name)
	}
	if id != "" {
		return fmt.Sprintf(" 小组: %s", id[:min(8, len(id))])
	}
	return ""
}

// locLabel returns a human-readable location label.
func locLabel(name, id string) string {
	if name != "" {
		return fmt.Sprintf(" 地址: %s", name)
	}
	if id != "" {
		return fmt.Sprintf(" 地址: %s", id[:min(8, len(id))])
	}
	return ""
}

func init() {
	// task create
	taskCreateCmd.Flags().String("name", "", "任务名称（快速模式必填）")
	taskCreateCmd.Flags().String("schedule", "", "调度类型: daily|weekly|monthly|yearly|interval|once")
	taskCreateCmd.Flags().String("data", "", "调度数据 JSON（见 --help 示例）")
	taskCreateCmd.Flags().String("group", "", "分配小组名称")
	taskCreateCmd.Flags().String("group-id", "", "分配小组ID（精确UUID，跳过名称解析）")
	taskCreateCmd.Flags().String("location", "", "分配地址名称")
	taskCreateCmd.Flags().String("location-id", "", "分配地址ID（精确UUID，跳过名称解析）")
	taskCreateCmd.Flags().String("kind", "simple", "任务类型: simple|inspection|chain")
	taskCreateCmd.Flags().String("extra", "", "额外数据 JSON（链步骤/巡检项，快速模式使用）")
	taskCreateCmd.Flags().BoolP("yes", "y", false, "跳过确认提示（引导模式）")

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

	// task enable
	taskEnableCmd.Flags().String("name", "", "任务名称")
	taskEnableCmd.Flags().String("id", "", "任务ID（完整UUID或≥3位前缀）")

	// task disable
	taskDisableCmd.Flags().String("name", "", "任务名称")
	taskDisableCmd.Flags().String("id", "", "任务ID（完整UUID或≥3位前缀）")

	// task logs
	taskLogsCmd.Flags().String("name", "", "任务名称")
	taskLogsCmd.Flags().String("id", "", "任务ID（完整UUID或≥3位前缀）")
	taskLogsCmd.Flags().Int("limit", 20, "日志条数上限")
	taskLogsCmd.Flags().Bool("user", false, "仅显示用户操作日志")

	// task children
	taskChildrenCmd.Flags().String("name", "", "父任务名称")
	taskChildrenCmd.Flags().String("id", "", "父任务ID（完整UUID或≥3位前缀）")

	taskCmd.AddCommand(taskCreateCmd)
	taskCmd.AddCommand(taskListCmd)
	taskCmd.AddCommand(taskInfoCmd)
	taskCmd.AddCommand(taskUpdateCmd)
	taskCmd.AddCommand(taskDeleteCmd)
	taskCmd.AddCommand(taskEnableCmd)
	taskCmd.AddCommand(taskDisableCmd)
	taskCmd.AddCommand(taskTriggerCmd)
	taskCmd.AddCommand(taskLogsCmd)
	taskCmd.AddCommand(taskChildrenCmd)
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
