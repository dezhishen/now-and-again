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

调度类型 (--schedule) 及对应的 --data 参数：
  daily       {"time":"09:00"}
  weekly      {"time":"09:00","days":[1,3,5]}    (1=周一,7=周日)
  monthly     {"time":"09:00","days":[1,15]}
  yearly      {"time":"09:00","day":6,"month":7}
  interval    {"days":3}
  once        {"date":"2026-07-10","time":"14:00"}`,
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

	req := &types.CreateTaskRequest{
		Task: types.Task{
			Name:         name,
			ScheduleType: schedule,
			ScheduleData: data,
			Kind:         kind,
		},
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
func buildChainExtra() interface{} {
	action.Println("\n🔗 任务链步骤 (每行一个步骤名，空行结束):")
	var steps []map[string]interface{}
	for i := 1; ; i++ {
		stepName := promptInput(fmt.Sprintf("  步骤%d 名称", i), "")
		if stepName == "" {
			break
		}
		steps = append(steps, map[string]interface{}{
			"name": stepName,
			"kind": "simple",
		})
	}
	if len(steps) == 0 {
		return nil
	}
	return map[string]interface{}{"steps": steps}
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
