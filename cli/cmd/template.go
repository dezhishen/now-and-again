package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/dezhishen/now-and-again/backend/pkg/types"
	"github.com/dezhishen/now-and-again/cli/internal/action"
	"github.com/dezhishen/now-and-again/cli/internal/resolver"
	"github.com/dezhishen/now-and-again/cli/internal/state"
	"github.com/spf13/cobra"
)

var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "任务模板",
	Long:  `查看可用模板，或通过引导式交互从模板创建任务。`,
}

// ─── template list ────────────────────────────────────────────────

var templateListCmd = &cobra.Command{
	Use:   "list",
	Short: "查看可用模板",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := autoEnsureFamily(); err != nil {
			return err
		}
		kind, _ := cmd.Flags().GetString("kind")
		templates, err := na.ListTemplates(context.Background(), kind)
		if err != nil {
			return err
		}
		if len(templates) == 0 {
			action.Println("📭 暂无可用模板")
			return nil
		}
		action.Printf("📋 可用模板 (%d个):\n", len(templates))
		for _, t := range templates {
			action.Printf("  %s %-25s  %-8s  %s", t.Icon, t.TemplateCode, t.Kind, t.Name)
		}
		return nil
	},
}

// ─── template use (interactive guided mode) ──────────────────────

var templateUseCmd = &cobra.Command{
	Use:   "use",
	Short: "从模板创建任务（支持引导式交互）",
	Long: `渲染一个模板并用结果创建任务。

引导式交互:
  na template use                     # 列出模板，逐步引导填充参数
  na template use --name "日常"       # 按名称匹配模板后引导填充
  na template use --name "日常" --yes # 跳过确认

快速模式（兼容旧用法）:
  na template use --code weekly_cleaning --params '{"area_name":"客厅"}'`,
	RunE: runTemplateUse,
}

func runTemplateUse(cmd *cobra.Command, args []string) error {
	if err := autoEnsureFamily(); err != nil {
		return err
	}

	ctx := context.Background()
	actID := action.CurrentID()
	cache := resolver.NewCache()

	code, _ := cmd.Flags().GetString("code")
	name, _ := cmd.Flags().GetString("name")
	paramsStr, _ := cmd.Flags().GetString("params")
	yes, _ := cmd.Flags().GetBool("yes")

	// Build args snapshot for action-id state
	stateArgs := map[string]string{
		"code":   code,
		"name":   name,
		"params": paramsStr,
	}

	// ── Step 1: Select template ──────────────────────────────
	var template *types.TaskTemplate

	if code != "" {
		// Direct code — fast path
		templates, err := na.ListTemplates(ctx, "")
		if err != nil {
			return err
		}
		for i := range templates {
			if templates[i].TemplateCode == code {
				template = &templates[i]
				break
			}
		}
		if template == nil {
			return fmt.Errorf("未找到模板代码: %s", code)
		}
	} else if name != "" {
		// Resolve by name with action-id state machine
		var err error
		_, template, err = cache.ResolveTemplateCodeInteractive(ctx, na, name, string(actID), "template use", stateArgs)
		if err != nil {
			return err
		}
	} else {
		// Interactive: list templates, let user choose
		templates, err := na.ListTemplates(ctx, "")
		if err != nil {
			return err
		}
		if len(templates) == 0 {
			return fmt.Errorf("📭 暂无可用模板")
		}

		action.Println("📋 可用模板:")
		for i, t := range templates {
			action.Printf("  %2d. %s %-25s  %-8s  %s", i+1, t.Icon, t.TemplateCode, t.Kind, t.Name)
		}
		fmt.Println()

		choice := promptInput("请选择模板 (编号/名称/code)", "")
		if choice == "" {
			return fmt.Errorf("已取消")
		}

		// Try as number first
		if n, err := strconv.Atoi(choice); err == nil && n >= 1 && n <= len(templates) {
			template = &templates[n-1]
		} else {
			// Resolve by name/code with action-id
			_, template, err = cache.ResolveTemplateCodeInteractive(ctx, na, choice, string(actID), "template use", stateArgs)
			if err != nil {
				return err
			}
		}
	}

	// Show template info
	action.Printf("\n📦 模板: %s %s", template.Icon, template.Name)
	if template.Description != "" {
		action.Printf("   %s", template.Description)
	}
	action.Printf("   类型: %s", template.Kind)

	// ── Step 2: Fill parameters ─────────────────────────────
	params := make(map[string]interface{})

	if paramsStr != "" {
		// Fast path: parse JSON params
		if err := json.Unmarshal([]byte(paramsStr), &params); err != nil {
			return fmt.Errorf("--params JSON格式错误: %w", err)
		}
	} else if len(template.Parameters) > 0 {
		action.Printf("\n📝 请填写模板参数 (%d个):", len(template.Parameters))
		for _, p := range template.Parameters {
			val := promptParam(p)
			if val != nil {
				params[p.Key] = val
			} else if p.Required {
				return fmt.Errorf("参数 %q 为必填", p.Label)
			}
		}
	}

	// ── Step 3: Group (optional) ────────────────────────────
	var groupID string
	gidFlag, _ := cmd.Flags().GetString("group-id")
	groupFlag, _ := cmd.Flags().GetString("group")

	if gidFlag != "" {
		groupID = gidFlag
	} else if groupFlag != "" {
		stateArgs["group"] = groupFlag
		var err error
		groupID, err = cache.ResolveGroupIDInteractive(ctx, na, groupFlag, string(actID), "template use", stateArgs)
		if err != nil {
			return err
		}
	} else if !yes {
		gInput := promptInput("分配到小组 (可选，回车跳过)", "")
		if gInput != "" {
			stateArgs["group"] = gInput
			var err error
			groupID, err = cache.ResolveGroupIDInteractive(ctx, na, gInput, string(actID), "template use", stateArgs)
			if err != nil {
				return err
			}
		}
	}

	// ── Step 4: Location (optional) ─────────────────────────
	var locationID string
	lidFlag, _ := cmd.Flags().GetString("location-id")
	locFlag, _ := cmd.Flags().GetString("location")

	if lidFlag != "" {
		locationID = lidFlag
	} else if locFlag != "" {
		stateArgs["location"] = locFlag
		var err error
		locationID, err = cache.ResolveLocationIDInteractive(ctx, na, locFlag, string(actID), "template use", stateArgs)
		if err != nil {
			return err
		}
	} else if !yes {
		lInput := promptInput("关联地址 (可选，回车跳过)", "")
		if lInput != "" {
			stateArgs["location"] = lInput
			var err error
			locationID, err = cache.ResolveLocationIDInteractive(ctx, na, lInput, string(actID), "template use", stateArgs)
			if err != nil {
				return err
			}
		}
	}

	// ── Step 5: Confirm & create ────────────────────────────
	if !yes {
		fmt.Println()
		action.Println("即将创建任务:")
		action.Printf("  模板: %s %s", template.Icon, template.Name)
		if len(params) > 0 {
			action.Printf("  参数: %v", params)
		}
		if groupID != "" {
			action.Printf("  小组: %s", groupID[:min(8, len(groupID))])
		}
		if locationID != "" {
			action.Printf("  地址: %s", locationID[:min(8, len(locationID))])
		}

		confirm := promptInput("\n确认创建? (Y/n)", "y")
		if strings.ToLower(confirm) != "y" && strings.ToLower(confirm) != "yes" && confirm != "" {
			state.CleanupIfDone(string(actID))
			action.Println("已取消")
			return nil
		}
	}

	// Create task from template (uses low-level API, no timezone conversion).
	task, err := na.CreateTaskFromTemplate(ctx, template.TemplateCode, params)
	if err != nil {
		return fmt.Errorf("创建任务失败: %w", err)
	}

	// Override group/location if user specified (post-creation update).
	if groupID != "" || locationID != "" {
		updateReq := &types.UpdateTaskRequest{
			Task: &types.Task{},
		}
		if groupID != "" {
			updateReq.Task.GroupID = groupID
		}
		if locationID != "" {
			updateReq.Task.LocationID = locationID
		}
		if _, err := na.UpdateTask(ctx, task.ID, updateReq); err != nil {
			action.Printf("⚠️  任务已创建，但设置小组/地址失败: %v", err)
		}
	}

	// Cleanup action-id state
	state.Delete(string(actID))

	action.Printf("✅ 已从模板创建: %s (%s)", task.Name, task.ID[:min(6, len(task.ID))])
	return nil
}

// ─── Interactive helpers ──────────────────────────────────────────

func promptInput(label, defaultVal string) string {
	if defaultVal != "" {
		fmt.Printf("→ %s [default: %s]: ", label, defaultVal)
	} else {
		fmt.Printf("→ %s: ", label)
	}
	var input string
	fmt.Scanln(&input)
	if input == "" {
		return defaultVal
	}
	return strings.TrimSpace(input)
}

// promptParam guides the user through filling a single template parameter.
func promptParam(p types.TemplateParameter) interface{} {
	required := ""
	if p.Required {
		required = " (必填)"
	}
	fmt.Printf("\n  ▸ %s [%s]%s\n", p.Label, p.Type, required)
	if p.Description != "" {
		fmt.Printf("    %s\n", p.Description)
	}
	if p.Default != nil {
		fmt.Printf("    默认: %v\n", p.Default)
	}

	switch p.Type {
	case "select":
		if len(p.Options) > 0 {
			for i, opt := range p.Options {
				label := fmt.Sprintf("    %d. %s", i+1, opt.Label)
				if opt.Value != "" && opt.Value != opt.Label {
					label += fmt.Sprintf("  (%s)", opt.Value)
				}
				fmt.Println(label)
			}
			choice := promptInput("  选择", "")
			if choice == "" {
				return p.Default
			}
			if n, err := strconv.Atoi(choice); err == nil && n >= 1 && n <= len(p.Options) {
				return p.Options[n-1].Value
			}
			for _, opt := range p.Options {
				if strings.EqualFold(opt.Label, choice) || strings.EqualFold(opt.Value, choice) {
					return opt.Value
				}
			}
			return choice
		}
		fallthrough
	case "string":
		prompt := "  输入"
		if p.Placeholder != "" {
			prompt += " (" + p.Placeholder + ")"
		}
		defStr := ""
		if p.Default != nil {
			defStr = fmt.Sprintf("%v", p.Default)
		}
		val := promptInput(prompt, defStr)
		if val == "" {
			return p.Default
		}
		return val
	case "int":
		defStr := ""
		if p.Default != nil {
			defStr = fmt.Sprintf("%v", p.Default)
		}
		val := promptInput("  输入整数", defStr)
		if val == "" {
			return p.Default
		}
		if n, err := strconv.Atoi(val); err == nil {
			return n
		}
		fmt.Println("  ⚠ 无效整数，使用默认值")
		return p.Default
	case "float":
		defStr := ""
		if p.Default != nil {
			defStr = fmt.Sprintf("%v", p.Default)
		}
		val := promptInput("  输入数字", defStr)
		if val == "" {
			return p.Default
		}
		if n, err := strconv.ParseFloat(val, 64); err == nil {
			return n
		}
		fmt.Println("  ⚠ 无效数字，使用默认值")
		return p.Default
	case "bool":
		defStr := ""
		if p.Default != nil {
			defStr = fmt.Sprintf("%v", p.Default)
		}
		val := promptInput("  输入 (true/false)", defStr)
		if val == "" {
			return p.Default
		}
		switch strings.ToLower(val) {
		case "true", "yes", "y", "1":
			return true
		case "false", "no", "n", "0":
			return false
		}
		fmt.Println("  ⚠ 无效布尔值，使用默认值")
		return p.Default
	case "time":
		defStr := ""
		if p.Default != nil {
			defStr = fmt.Sprintf("%v", p.Default)
		}
		val := promptInput("  输入时间 (HH:MM)", defStr)
		if val == "" {
			return p.Default
		}
		return val
	default:
		defStr := ""
		if p.Default != nil {
			defStr = fmt.Sprintf("%v", p.Default)
		}
		val := promptInput("  输入", defStr)
		if val == "" {
			return p.Default
		}
		return val
	}
}

func init() {
	templateListCmd.Flags().String("kind", "", "按类型筛选: simple|inspection|chain")

	templateUseCmd.Flags().String("code", "", "模板代码 (精确匹配，跳过选择)")
	templateUseCmd.Flags().String("name", "", "模板名称 (支持模糊匹配)")
	templateUseCmd.Flags().String("params", "", "模板参数 JSON (跳过交互式填充)")
	templateUseCmd.Flags().String("group", "", "分配到小组 (名称)")
	templateUseCmd.Flags().String("group-id", "", "分配到小组 (UUID)")
	templateUseCmd.Flags().String("location", "", "关联地址 (名称)")
	templateUseCmd.Flags().String("location-id", "", "关联地址 (UUID)")
	templateUseCmd.Flags().BoolP("yes", "y", false, "跳过确认提示")

	templateCmd.AddCommand(templateListCmd)
	templateCmd.AddCommand(templateUseCmd)
}
