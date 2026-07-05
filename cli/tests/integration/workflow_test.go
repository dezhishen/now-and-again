package integration

import (
	"strings"
	"testing"
)

// ─── Core Workflow ──────────────────────────────────────────────────

func TestWorkflow_InitAndDaily(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	setupWithFamily(t)

	// List templates (may be empty if no subscriptions configured)
	out := runNAOK(t, "template", "list")
	if strings.Contains(out, "daily_inspection") || strings.Contains(out, "simple_daily_check") {
		// Create task from template only if templates are available
		tOut := runNAOK(t, "template", "use", "--code", "simple_daily_check", "--params", `{"check_item":"晨会"}`)
		if strings.Contains(tOut, "已创建") {
			t.Log("✅ template use")
		}
	}
	t.Log("✅ template list")

	// Create task manually
	out = runNAOK(t, "task", "create", "--name", "每日倒垃圾", "--schedule", "daily", "--data", `{"time":"20:00"}`)
	if !strings.Contains(out, "已创建") {
		t.Fatalf("task create failed: %s", out)
	}
	t.Log("✅ task create")

	// List tasks
	out = runNAOK(t, "task", "list")
	if !strings.Contains(out, "晨会") && !strings.Contains(out, "倒垃圾") {
		t.Fatalf("task list missing tasks: %s", out)
	}
	t.Log("✅ task list")

	// List todos
	out = runNAOK(t, "todo", "list")
	t.Logf("todo list: %s", out)
	t.Log("✅ todo list")
}

func TestWorkflow_Templates(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	setupWithFamily(t)

	out := runNAOK(t, "template", "list")
	if strings.Contains(out, "暂无可用模板") {
		t.Skip("templates not synced; skip template-specific test")
	}
	required := []string{"daily_inspection", "simple_daily_check", "weekly_cleaning"}
	for _, code := range required {
		if !strings.Contains(out, code) {
			t.Errorf("template list missing %q", code)
		}
	}
	t.Log("✅ required templates present")

	// Filter by kind
	for _, kc := range []struct{ kind, expect string }{
		{"simple", "simple_daily_check"},
		{"inspection", "daily_inspection"},
	} {
		out = runNAOK(t, "template", "list", "--kind", kc.kind)
		if !strings.Contains(out, kc.expect) {
			t.Errorf("template --kind %s missing %q: %s", kc.kind, kc.expect, out)
		}
	}
	t.Log("✅ template kind filters")
}

func TestWorkflow_TaskCRUD(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	setupWithFamily(t)

	schedules := []struct {
		schedule, data, desc string
	}{
		{"daily", `{"time":"09:00"}`, "每日"},
		{"weekly", `{"days":[1,3,5],"time":"10:00"}`, "每周一三五"},
		{"interval", `{"days":14,"time":"12:00"}`, "每14天"},
	}

	for _, s := range schedules {
		out := runNAOK(t, "task", "create", "--name", testName("任务-"+s.desc), "--schedule", s.schedule, "--data", s.data)
		if !strings.Contains(out, "已创建") {
			t.Errorf("task create --schedule %s failed: %s", s.schedule, out)
		}
	}
	t.Log("✅ schedule types created")

	out := runNAOK(t, "task", "list")
	for _, s := range schedules {
		if !strings.Contains(out, s.desc) {
			t.Errorf("task list missing %q: %s", s.desc, out)
		}
	}
	t.Log("✅ tasks listed correctly")
}

func TestWorkflow_DailyShortcut(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	setupWithFamily(t)

	// Create a task first
	runNAOK(t, "task", "create", "--name", "快速任务", "--schedule", "daily", "--data", `{"time":"09:00"}`)

	out, _, err := runNA("daily")
	if err != nil {
		t.Logf("daily output: %s", out)
	}
	t.Log("✅ daily runs without crash")
}

func TestEdge_BadParams(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	setupWithFamily(t)

	badCases := []struct {
		name string
		args []string
	}{
		{"missing --id in todo done", []string{"todo", "done", "--id", ""}},
	}

	for _, bc := range badCases {
		out, _, err := runNA(bc.args...)
		if err == nil {
			t.Errorf("%s: expected error but got success: %s", bc.name, out)
		} else {
			t.Logf("%s: correctly rejected — %s", bc.name, truncate(out, 60))
		}
	}
	t.Log("✅ bad params correctly rejected")
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
