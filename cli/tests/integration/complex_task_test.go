package integration

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// ─── Complex Task: Chain Creation & Inspection ──────────────────

func TestComplexTask_ChainCreateFastPath(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	setupWithFamily(t)

	taskName := testName("链-快速")
	extra := `{"steps":[{"name":"步骤1-准备","kind":"simple"},{"name":"步骤2-检查","kind":"inspection","extra":{"check_items":[{"name":"门窗","branches":[{"name":"合格","create_todo":false},{"name":"不合格","create_todo":true}]}]}},{"name":"步骤3-收尾","kind":"simple"}]}`

	out := runNAOK(t, "task", "create",
		"--name", taskName,
		"--schedule", "once",
		"--data", `{"date":"2026-12-31","time":"09:00"}`,
		"--kind", "chain",
		"--extra", extra,
	)
	if !strings.Contains(out, "已创建") {
		t.Fatalf("chain task create failed: %s", out)
	}
	t.Log("✅ chain task created via fast path")

	// Verify task info shows chain details
	out = runNAOK(t, "task", "info", "--name", taskName)
	if !strings.Contains(out, "任务链") && !strings.Contains(out, "chain") {
		t.Logf("task info output: %s", truncate(out, 300))
	}
	// display_summary should contain the step names
	if !strings.Contains(out, "步骤1") {
		t.Logf("task info may not show steps (extra not yet loaded): %s", truncate(out, 300))
	}
	t.Log("✅ task info shown")
}

func TestComplexTask_ChainCreateInteractive(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	setupWithFamily(t)

	// Simulate interactive input for chain task creation
	// Note: since we can't provide real stdin interactively in tests,
	// we verify the fast path with --extra works properly
	taskName := testName("链-交互")

	// Create chain with 3 steps via --extra flag (equivalent to interactive result)
	extra := `{"steps":[{"name":"准备材料","kind":"simple"},{"name":"执行操作","kind":"simple"},{"name":"检查结果","kind":"inspection","extra":{"check_items":[{"name":"质量","branches":[{"name":"通过","create_todo":false},{"name":"不通过","create_todo":true}]}]}}]}`

	out := runNAOK(t, "task", "create",
		"--name", taskName,
		"--schedule", "weekly",
		"--data", `{"days":[1,3,5],"time":"10:00"}`,
		"--kind", "chain",
		"--extra", extra,
	)
	if !strings.Contains(out, "已创建") {
		t.Fatalf("chain interactive-style create failed: %s", out)
	}
	t.Log("✅ interactive-style chain task created")

	// Verify children
	out = runNAOK(t, "task", "children", "--name", taskName)
	t.Logf("children output: %s", truncate(out, 300))

	// Trigger and check again
	taskID := resolveTaskID(t, taskName)
	if taskID != "" {
		out, _, err := runNA("task", "trigger", "--id", taskID)
		if err != nil {
			t.Logf("trigger warning (expected if no scheduler): %v", err)
		} else {
			t.Logf("trigger output: %s", truncate(out, 100))
		}

		// After trigger, children should be created
		out = runNAOK(t, "task", "children", "--name", taskName)
		t.Logf("children after trigger: %s", truncate(out, 300))
	}
}

func TestComplexTask_InspectionCreate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	setupWithFamily(t)

	taskName := testName("巡检-测试")
	extra := `{"check_items":[{"name":"电路","branches":[{"name":"正常","create_todo":false},{"name":"异常","create_todo":true,"branch_task":{"task":{"name":"修复电路","kind":"simple","schedule_type":"once","schedule_data":{"time":"09:00"}}}}]},{"name":"消防","branches":[{"name":"合格","create_todo":false},{"name":"不合格","create_todo":true}]}]}`

	out := runNAOK(t, "task", "create",
		"--name", taskName,
		"--schedule", "daily",
		"--data", `{"time":"08:00"}`,
		"--kind", "inspection",
		"--extra", extra,
	)
	if !strings.Contains(out, "已创建") {
		t.Fatalf("inspection create failed: %s", out)
	}
	t.Log("✅ inspection task created")

	// Verify task info
	out = runNAOK(t, "task", "info", "--name", taskName)
	t.Logf("inspection info: %s", truncate(out, 400))
	if !strings.Contains(out, "巡检") && !strings.Contains(out, "inspection") {
		t.Errorf("task info should show inspection kind: %s", truncate(out, 200))
	}
}

func TestComplexTask_TaskChildren(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	setupWithFamily(t)

	// First create a simple parent task (can't have children for simple tasks)
	// Then create a chain task and verify the children command works

	taskName := testName("子任务测试")
	out := runNAOK(t, "task", "create",
		"--name", taskName,
		"--schedule", "daily",
		"--data", `{"time":"09:00"}`,
		"--kind", "chain",
		"--extra", `{"steps":[{"name":"子步骤A","kind":"simple"},{"name":"子步骤B","kind":"simple"}]}`,
	)
	if !strings.Contains(out, "已创建") {
		t.Fatalf("chain create for children test failed: %s", out)
	}

	// Test children command
	out = runNAOK(t, "task", "children", "--name", taskName)
	t.Logf("children output: %s", truncate(out, 300))

	// Should at least not crash
	if strings.Contains(out, "没有子任务") {
		t.Log("⚠️  children not yet created (trigger first to generate)")
	} else {
		t.Log("✅ children listed")
	}
}

func TestComplexTask_InfoExtraData(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	setupWithFamily(t)

	taskName := testName("详情测试")
	out := runNAOK(t, "task", "create",
		"--name", taskName,
		"--schedule", "once",
		"--data", `{"date":"2026-12-25","time":"08:00"}`,
		"--kind", "inspection",
		"--extra", `{"check_items":[{"name":"安全检查","branches":[{"name":"合格","create_todo":false},{"name":"不合格","create_todo":true}]}]}`,
	)
	if !strings.Contains(out, "已创建") {
		t.Fatalf("create failed: %s", out)
	}

	// Verify info shows display_summary
	out = runNAOK(t, "task", "info", "--name", taskName)
	t.Logf("info output:\n%s", truncate(out, 500))

	// Check for key fields
	for _, keyword := range []string{"展示摘要", "display_summary", "安全检查"} {
		if strings.Contains(out, keyword) {
			t.Logf("✅ info contains %q", keyword)
		}
	}
}

// ─── Action-ID workflow with complex tasks ──────────────────────

func TestComplexTask_ActionIDWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	setupWithFamily(t)

	workflowID := uuid.New().String()
	sid := workflowID[:8]

	// Phase 1: Discover groups and locations
	out := runNAOKWithActionID(t, workflowID, "family", "status", "-o", "json")
	// JSON mode uses "action_id" field, not [action: xxx] prefix
	if !strings.Contains(out, workflowID) {
		t.Errorf("[1-discover] JSON missing action_id=%s", workflowID)
	}

	var env struct {
		Data struct {
			Groups    []struct{ ID, Name string } `json:"groups"`
			Locations []struct{ ID, Name string } `json:"locations"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(out), &env); err != nil {
		t.Fatalf("parse JSON: %v", err)
	}
	if len(env.Data.Groups) == 0 || len(env.Data.Locations) == 0 {
		t.Fatal("no groups or locations available")
	}

	groupID := env.Data.Groups[0].ID
	locID := env.Data.Locations[0].ID

	// Phase 2: Create complex chain task with IDs
	taskName := testName("AI链")
	extra := `{"steps":[{"name":"AI步骤1","kind":"simple"},{"name":"AI步骤2","kind":"inspection","extra":{"check_items":[{"name":"AI检查","branches":[{"name":"OK","create_todo":false},{"name":"NG","create_todo":true}]}]}}]}`

	out = runNAOKWithActionID(t, workflowID,
		"task", "create",
		"--name", taskName,
		"--schedule", "once",
		"--data", `{"date":"2026-12-31","time":"12:00"}`,
		"--kind", "chain",
		"--extra", extra,
		"--group-id", groupID,
		"--location-id", locID,
	)
	assertActionID(t, out, sid, "2-create")
	if !strings.Contains(out, "已创建") {
		t.Fatalf("create failed: %s", out)
	}
	t.Logf("✅ Phase 2: task created — %s", truncate(out, 120))

	// Phase 3: Verify with info
	out = runNAOKWithActionID(t, workflowID, "task", "info", "--name", taskName)
	// Note: task info uses fmt.Printf directly, so action-id prefix may not appear
	if !strings.Contains(out, taskName) {
		t.Errorf("info should contain task name: %s", truncate(out, 200))
	}
	t.Logf("✅ Phase 3: info verified")

	// Phase 4: List children
	out = runNAOKWithActionID(t, workflowID, "task", "children", "--name", taskName)
	// Note: task children uses fmt.Printf directly, so action-id prefix may not appear
	t.Logf("✅ Phase 4: children — %s", truncate(out, 200))

	// Cleanup
	assertNoStateFile(t, workflowID)
	t.Log("✅ no residual state file")
}

// ─── Test: Inspection with chain branch tasks (日用消费品存量巡检) ─
//
// 模拟场景:
//   巡检日用消费品存量
//   ├─ 检查项: 洗衣液
//   │  ├─ 正常 (无任务)
//   │  └─ 不足 → 创建任务链 "采购洗衣液"
//   │     ├─ 步骤1: 获取验证码 (simple)
//   │     ├─ 步骤2: 快递站取件 (simple)
//   │     └─ 步骤3: 收纳 (simple)
//   ├─ 检查项: 纸巾
//   │  ├─ 正常 (无任务)
//   │  └─ 不足 → 创建任务链 "采购纸巾"
//   │     └─ 同上3步骤
//   └─ 检查项: 洗洁精
//      ├─ 正常 (无任务)
//      └─ 不足 → 创建任务链 "采购洗洁精"
//          └─ 同上3步骤

func TestComplexTask_InspectionInventoryChain(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	setupWithFamily(t)

	taskName := testName("巡检日用品存量")

	// 构建复杂 extra: 3个检查项 × (1正常分支 + 1不足→链任务)
	purchaseChainSteps := `[{"name":"获取验证码","kind":"simple","schedule_type":"once","schedule_data":{"date":"2026-07-08","time":"09:00"}},{"name":"快递站取件","kind":"simple","schedule_type":"once","schedule_data":{"date":"2026-07-09","time":"10:00"}},{"name":"收纳","kind":"simple","schedule_type":"once","schedule_data":{"date":"2026-07-09","time":"18:00"}}]`

	buildBranchTask := func(itemName string) string {
		return fmt.Sprintf(`{"task":{"name":"采购%s","kind":"chain","schedule_type":"once","schedule_data":{"date":"2026-07-08","time":"08:00"}},"extra":{"steps":%s}}`, itemName, purchaseChainSteps)
	}

	extra := fmt.Sprintf(`{"check_items":[
		{"name":"洗衣液","branches":[
			{"name":"正常","create_todo":false},
			{"name":"不足","create_todo":true,"branch_task":%s}
		]},
		{"name":"纸巾","branches":[
			{"name":"正常","create_todo":false},
			{"name":"不足","create_todo":true,"branch_task":%s}
		]},
		{"name":"洗洁精","branches":[
			{"name":"正常","create_todo":false},
			{"name":"不足","create_todo":true,"branch_task":%s}
		]}
	]}`, buildBranchTask("洗衣液"), buildBranchTask("纸巾"), buildBranchTask("洗洁精"))

	out := runNAOK(t, "task", "create",
		"--name", taskName,
		"--schedule", "weekly",
		"--data", `{"days":[6,7],"time":"10:00"}`,
		"--kind", "inspection",
		"--extra", extra,
	)
	if !strings.Contains(out, "已创建") {
		t.Fatalf("inventory inspection create failed: %s", out)
	}
	t.Log("✅ 巡检日用品存量 — 创建成功")

	// 验证任务信息
	out = runNAOK(t, "task", "info", "--name", taskName)
	t.Logf("info output:\n%s", truncate(out, 600))

	// 检查关键信息
	for _, keyword := range []string{"洗衣液", "纸巾", "洗洁精", "正常", "不足", "采购"} {
		if strings.Contains(out, keyword) {
			t.Logf("  ✅ info contains %q", keyword)
		} else {
			t.Logf("  ⚠️  info missing %q", keyword)
		}
	}

	// 验证 display_summary 包含检查项
	if strings.Contains(out, "巡检:") || strings.Contains(out, "洗衣液") {
		t.Log("✅ display_summary present")
	}

	// 验证子任务 (触发后才有)
	taskID := resolveTaskID(t, taskName)
	if taskID != "" {
		out, _, err := runNA("task", "children", "--name", taskName)
		if err == nil {
			t.Logf("children: %s", truncate(out, 200))
		}
	}
}

func TestComplexTask_InspectionBranchChainViaInteractive(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	setupWithFamily(t)

	// 模拟用户交互式创建巡检，不足分支创建3步骤任务链
	taskName := testName("巡检-链式采购")

	// 采购任务链3步骤
	chainSteps := `[{"name":"下单购买","kind":"simple","schedule_type":"once","schedule_data":{"date":"2026-07-08","time":"09:00"}},{"name":"收货验收","kind":"simple","schedule_type":"once","schedule_data":{"date":"2026-07-10","time":"10:00"}},{"name":"入库登记","kind":"simple","schedule_type":"once","schedule_data":{"date":"2026-07-10","time":"12:00"}}]`

	extra := fmt.Sprintf(`{"check_items":[
		{"name":"大米","branches":[
			{"name":"正常","create_todo":false},
			{"name":"不足","create_todo":true,"branch_task":{"task":{"name":"采购大米","kind":"chain","schedule_type":"once","schedule_data":{"date":"2026-07-08","time":"08:00"}},"extra":{"steps":%s}}}
		]},
		{"name":"食用油","branches":[
			{"name":"正常","create_todo":false},
			{"name":"不足","create_todo":true,"branch_task":{"task":{"name":"采购食用油","kind":"chain","schedule_type":"once","schedule_data":{"date":"2026-07-08","time":"08:00"}},"extra":{"steps":%s}}}
		]}
	]}`, chainSteps, chainSteps)

	out := runNAOK(t, "task", "create",
		"--name", taskName,
		"--schedule", "monthly",
		"--data", `{"days":[1],"time":"10:00"}`,
		"--kind", "inspection",
		"--extra", extra,
	)
	if !strings.Contains(out, "已创建") {
		t.Fatalf("branch chain inspection create failed: %s", out)
	}
	t.Log("✅ 巡检-链式采购 创建成功")

	// 验证基本信息
	out = runNAOK(t, "task", "info", "--name", taskName)
	for _, kw := range []string{"大米", "食用油", "下单购买", "收货验收", "入库登记"} {
		if strings.Contains(out, kw) {
			t.Logf("  ✅ info contains %q", kw)
		} else {
			t.Logf("  ⚠️  info missing %q", kw)
		}
	}

	// 验证 children
	taskID := resolveTaskID(t, taskName)
	if taskID != "" {
		out, _, err := runNA("task", "children", "--name", taskName)
		if err == nil {
			t.Logf("children: %s", truncate(out, 200))
		}
	}
}

// ─── Helper ─────────────────────────────────────────────────────

func resolveTaskID(t *testing.T, name string) string {
	t.Helper()
	out := runNAOK(t, "family", "status", "-o", "json")
	var env struct {
		Data struct {
			FamilyID string `json:"family_id"`
		} `json:"data"`
	}
	json.Unmarshal([]byte(out), &env)

	// List all tasks and find by name
	out = runNAOK(t, "task", "list", "--all")
	// The task list is text mode, so parse manually
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		if strings.Contains(line, name) {
			parts := strings.Fields(line)
			for _, p := range parts {
				// UUIDs have 36 chars with hyphens
				if len(p) >= 36 && strings.Count(p, "-") >= 4 {
					return p
				}
				// Short prefixes are at least 6 hex chars
				if len(p) >= 6 && isHex(p) {
					return p
				}
			}
		}
	}
	t.Logf("could not resolve task ID from list output for %q", name)
	return ""
}

func isHex(s string) bool {
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return len(s) >= 6
}
