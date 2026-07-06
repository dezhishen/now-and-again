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

func runTaskCreateInteractive(ctx context.Context, cache *resolver.Cache, actID, kind, groupName, groupID, locationName, locationID string, yes bool) error {
	action.Println("\n📝 引导式任务创建\n")

	// 1. Task name
	name := promptInput("任务名称", "")
	if name == "" {
		return fmt.Errorf("任务名称不能为空")
	}

	// 2. Task kind
	if kind == "" {
		action.Println("\n任务类型:")
		action.Println("  1. 📌 simple     — 简单定时任务")
		action.Println("  2. 🔍 inspection — 巡检任务 (检查项)")
		action.Println("  3. 🔗 chain      — 任务链 (多步骤)")
		k := promptInput("选择类型", "1")
		switch k {
		case "2":
			kind = "inspection"
		case "3":
			kind = "chain"
		default:
			kind = "simple"
		}
	}

	// 3. Schedule type
	action.Println("\n调度类型:")
	action.Println("  1. daily    — 每天")
	action.Println("  2. weekly   — 每周")
	action.Println("  3. monthly  — 每月")
	action.Println("  4. yearly   — 每年")
	action.Println("  5. interval — 间隔N天")
	action.Println("  6. once     — 仅一次")
	schedChoice := promptInput("选择调度", "1")

	schedule, schedData := buildScheduleData(schedChoice)
	if schedule == "" {
		return fmt.Errorf("无效的调度类型")
	}

	// 4. Extra data for inspection/chain
	var extra interface{}
	if kind == "inspection" {
		extra = buildInspectionExtra()
	} else if kind == "chain" {
		extra = buildChainExtra()
	}

	// 5. Group (optional)
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
		gInput := promptInput("\n分配到小组 (可选，回车跳过)", "")
		if gInput != "" {
			stateArgs := map[string]string{"name": name, "kind": kind, "schedule": schedule, "group": gInput}
			var err error
			gid, err = cache.ResolveGroupIDInteractive(ctx, na, gInput, actID, "task create", stateArgs)
			if err != nil {
				return err
			}
		}
	}

	// 6. Location (optional)
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
		lInput := promptInput("关联地址 (可选，回车跳过)", "")
		if lInput != "" {
			stateArgs := map[string]string{"name": name, "kind": kind, "schedule": schedule, "location": lInput}
			var err error
			lid, err = cache.ResolveLocationIDInteractive(ctx, na, lInput, actID, "task create", stateArgs)
			if err != nil {
				return err
			}
		}
	}

	// 7. Confirm
	schedJSON, _ := json.MarshalIndent(schedData, "  ", "  ")
	if !yes {
		fmt.Println()
		action.Println("即将创建任务:")
		action.Printf("  名称: %s", name)
		action.Printf("  类型: %s", kind)
		action.Printf("  调度: %s", schedule)
		action.Printf("  参数: %s", string(schedJSON))
		if gid != "" {
			action.Printf("  小组: %s", gid[:min(8, len(gid))])
		}
		if lid != "" {
			action.Printf("  地址: %s", lid[:min(8, len(lid))])
		}

		confirm := promptInput("\n确认创建? (Y/n)", "y")
		if strings.ToLower(confirm) != "y" && strings.ToLower(confirm) != "yes" && confirm != "" {
			state.CleanupIfDone(actID)
			action.Println("已取消")
			return nil
		}
	}

	// 8. Create
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

// buildScheduleData prompts for schedule details based on type.
func buildScheduleData(choice string) (scheduleType string, data map[string]interface{}) {
	data = make(map[string]interface{})
	switch choice {
	case "1", "daily":
		scheduleType = "daily"
		t := promptInput("  时间 (HH:MM)", "09:00")
		data["time"] = t
	case "2", "weekly":
		scheduleType = "weekly"
		t := promptInput("  时间 (HH:MM)", "09:00")
		data["time"] = t
		daysStr := promptInput("  星期几 (1-7, 逗号分隔, 1=周一)", "1,3,5")
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
		t := promptInput("  时间 (HH:MM)", "09:00")
		data["time"] = t
		daysStr := promptInput("  几号 (1-31, 逗号分隔)", "1,15")
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
		t := promptInput("  时间 (HH:MM)", "09:00")
		data["time"] = t
		mStr := promptInput("  月份 (1-12)", "7")
		m, _ := strconv.Atoi(mStr)
		data["month"] = m
		dStr := promptInput("  日期 (1-31)", "6")
		d, _ := strconv.Atoi(dStr)
		data["day"] = d
	case "5", "interval":
		scheduleType = "interval"
		dStr := promptInput("  间隔天数", "3")
		d, _ := strconv.Atoi(dStr)
		data["days"] = d
	case "6", "once":
		scheduleType = "once"
		dateStr := promptInput("  日期 (YYYY-MM-DD)", time.Now().Format("2006-01-02"))
		data["date"] = dateStr
		tStr := promptInput("  时间 (HH:MM)", "09:00")
		data["time"] = tStr
	default:
		return "", nil
	}
	return
}

// buildInspectionExtra builds check_items interactively for inspection tasks.
func buildInspectionExtra() interface{} {
	action.Println("\n🔍 巡检项目 (每行一个，空行结束):")
	var items []map[string]interface{}
	for i := 1; ; i++ {
		itemName := promptInput(fmt.Sprintf("  项目%d 名称", i), "")
		if itemName == "" {
			break
		}
		item := map[string]interface{}{"name": itemName}

		action.Printf("    结果选项 (逗号分隔，如: 合格,不合格):")
		branchesStr := promptInput("    ", "合格,不合格")
		var branches []map[string]interface{}
		for _, b := range strings.Split(branchesStr, ",") {
			b = strings.TrimSpace(b)
			if b != "" {
				branches = append(branches, map[string]interface{}{
					"name":        b,
					"create_todo": false,
				})
			}
		}
		item["branches"] = branches
		items = append(items, item)
	}
	if len(items) == 0 {
		return nil
	}
	return map[string]interface{}{"check_items": items}
}

// buildChainExtra builds steps interactively for chain tasks.
// Each step can be simple or inspection with its own config.
func buildChainExtra() interface{} {
	action.Println("\n🔗 任务链步骤 (每次输入步骤名称，空行结束)")
	action.Println("   提示: 可在名称后附加 |inspection 指定为巡检步骤，如: 水电检查|inspection")
	var steps []map[string]interface{}
	for i := 1; ; i++ {
		stepName := promptInput(fmt.Sprintf("  步骤%d 名称", i), "")
		if stepName == "" {
			break
		}

		stepKind := "simple"
		// Support inline kind specifier: "name|inspection" or "name|chain"
		if idx := strings.LastIndex(stepName, "|"); idx > 0 {
			suffix := strings.ToLower(stepName[idx+1:])
			switch suffix {
			case "inspection", "insp", "巡检":
				stepKind = "inspection"
				stepName = stepName[:idx]
			case "chain", "链":
				stepKind = "chain"
				stepName = stepName[:idx]
			case "simple", "简单":
				stepKind = "simple"
				stepName = stepName[:idx]
			}
		}

		step := map[string]interface{}{
			"name": stepName,
			"kind": stepKind,
		}

		// For inspection steps, optionally configure check items
		if stepKind == "inspection" {
			action.Printf("    [%s 为巡检步骤] 可选配置检查项 (空行跳过):", stepName)
			step["extra"] = buildInspectionExtraSimple()
		}

		steps = append(steps, step)
	}
	if len(steps) == 0 {
		return nil
	}
	return map[string]interface{}{"steps": steps}
}

// buildInspectionExtraSimple builds a simplified inspection extra without too many prompts.
func buildInspectionExtraSimple() interface{} {
	var items []map[string]interface{}
	for i := 1; i <= 5; i++ {
		itemName := promptInput(fmt.Sprintf("    检查项%d 名称", i), "")
		if itemName == "" {
			break
		}
		item := map[string]interface{}{"name": itemName}

		branchesStr := promptInput(fmt.Sprintf("    结果选项 (逗号分隔，默认: 合格,不合格)"), "合格,不合格")
		var branches []map[string]interface{}
		for _, b := range strings.Split(branchesStr, ",") {
			b = strings.TrimSpace(b)
			if b != "" {
				branches = append(branches, map[string]interface{}{
					"name":        b,
					"create_todo": false,
				})
			}
		}
		item["branches"] = branches
		items = append(items, item)
	}
	if len(items) == 0 {
		return nil
	}
	return map[string]interface{}{"check_items": items}
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

	// task children
	taskChildrenCmd.Flags().String("name", "", "父任务名称")
	taskChildrenCmd.Flags().String("id", "", "父任务ID（完整UUID或≥3位前缀）")

	taskCmd.AddCommand(taskCreateCmd)
	taskCmd.AddCommand(taskListCmd)
	taskCmd.AddCommand(taskInfoCmd)
	taskCmd.AddCommand(taskUpdateCmd)
	taskCmd.AddCommand(taskDeleteCmd)
	taskCmd.AddCommand(taskTriggerCmd)
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
