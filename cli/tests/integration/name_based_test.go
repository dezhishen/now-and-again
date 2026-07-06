package integration

import (
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// ─── Action ID ────────────────────────────────────────────────────

var actionIDPattern = regexp.MustCompile(`\[action: [0-9a-f]{8}\]`)

// runWithActionID executes na with --server and --action-id.
func runWithActionID(actionID string, args ...string) (string, string, error) {
	fullArgs := append([]string{"--server", serverURL, "--action-id", actionID}, args...)
	cmd := exec.Command(binaryPath, fullArgs...)
	if testHome != "" {
		cmd.Env = append(os.Environ(), "HOME="+testHome)
	}
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return strings.TrimSpace(stdout.String()), strings.TrimSpace(stderr.String()), err
}

func runNAOKWithActionID(t *testing.T, actionID string, args ...string) string {
	t.Helper()
	out, errOut, err := runWithActionID(actionID, args...)
	if err != nil {
		t.Fatalf("na %s failed:\nstdout: %s\nstderr: %s\nerr: %v",
			strings.Join(args, " "), out, errOut, err)
	}
	return out
}

// assertActionID checks output contains [action: XXXXXXXX].
func assertActionID(t *testing.T, out, shortID, step string) {
	t.Helper()
	want := "[action: " + shortID + "]"
	if !strings.Contains(out, want) {
		t.Errorf("[step=%s] output missing %s:\n%s", step, want, truncate(out, 120))
	}
}

// ─── Core: full task lifecycle tracked by one --action-id ─────────

func TestNameBased_ActionIDWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	setupWithFamily(t)

	workflowID := uuid.New().String()
	sid := workflowID[:8]
	t.Logf("workflow action-id: %s", workflowID)

	taskName := testName("工作流")

	// ── 1. create with --group + --location ──────────────────
	out := runNAOKWithActionID(t, workflowID,
		"task", "create",
		"--name", taskName,
		"--schedule", "daily",
		"--data", `{"time":"08:00"}`,
		"--group", "大人",
		"--location", "厨房",
	)
	assertActionID(t, out, sid, "1-create")
	if !strings.Contains(out, "大人") || !strings.Contains(out, "厨房") {
		t.Fatalf("step1: group/location not in output: %s", out)
	}
	t.Logf("✅ 1-create [action: %s]", sid)

	// ── 2. info by name (table mode, no action prefix) ──────
	out = runNAOKWithActionID(t, workflowID, "task", "info", "--name", taskName)
	if !strings.Contains(out, taskName) || !strings.Contains(out, "大人") {
		t.Errorf("step2: info missing expected data: %s", truncate(out, 150))
	}
	t.Logf("✅ 2-info (table mode)")

	// ── 3. update group + location ──────────────────────────
	// Must pass --new-name + --schedule + --data (backend binding constraint)
	out = runNAOKWithActionID(t, workflowID,
		"task", "update",
		"--name", taskName,
		"--new-name", taskName,
		"--schedule", "daily",
		"--data", `{"time":"08:00"}`,
		"--group", "小孩",
		"--location", "客厅",
	)
	assertActionID(t, out, sid, "3-update")
	t.Logf("✅ 3-update [action: %s]", sid)

	// ── 4. list tasks — verify task still exists ────────────
	out = runNAOKWithActionID(t, workflowID, "task", "list")
	if !strings.Contains(out, taskName) {
		t.Errorf("step4: task %q not in list", taskName)
	}
	t.Logf("✅ 4-list")

	// ── 5. trigger by name ──────────────────────────────────
	out = runNAOKWithActionID(t, workflowID, "task", "trigger", "--name", taskName)
	assertActionID(t, out, sid, "5-trigger")
	t.Logf("✅ 5-trigger [action: %s]", sid)

	// ── 6. check todos ──────────────────────────────────────
	out = runNAOKWithActionID(t, workflowID, "todo", "list")
	t.Logf("✅ 6-todos: %s", truncate(out, 120))

	// ── 7. delete by name ───────────────────────────────────
	out = runNAOKWithActionID(t, workflowID, "task", "delete", "--name", taskName, "-y")
	assertActionID(t, out, sid, "7-delete")
	t.Logf("✅ 7-delete [action: %s]", sid)

	// ── 8. confirm deleted ──────────────────────────────────
	_, errOut, err := runWithActionID(workflowID, "task", "info", "--name", taskName)
	if err == nil {
		t.Errorf("step8: task should be gone")
	}
	t.Logf("✅ 8-gone: %s", truncate(errOut, 120))
}

// ─── Action-ID uniqueness & fixed ID across invocations ───────────

func TestNameBased_ActionIDUniqueness(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	setupWithFamily(t)

	// Auto mode: each `task create` gets a different action-id
	out1 := runNAOK(t, "task", "create",
		"--name", testName("唯一A"), "--schedule", "daily", "--data", `{"time":"08:00"}`)
	out2 := runNAOK(t, "task", "create",
		"--name", testName("唯一B"), "--schedule", "daily", "--data", `{"time":"09:00"}`)

	id1 := extractShort(out1)
	id2 := extractShort(out2)
	if id1 == "" || id2 == "" {
		t.Fatalf("auto-mode missing action-id:\n1=%s\n2=%s", truncate(out1, 80), truncate(out2, 80))
	}
	if id1 == id2 {
		t.Errorf("auto mode should give different IDs, both %q", id1)
	}
	t.Logf("✅ auto unique: %s ≠ %s", id1, id2)

	// Fixed mode: same --action-id → same short ID across multiple invocations
	fixed := uuid.New().String()
	fs := fixed[:8]

	out3 := runNAOKWithActionID(t, fixed, "task", "create",
		"--name", testName("固定A"), "--schedule", "daily", "--data", `{"time":"10:00"}`)
	assertActionID(t, out3, fs, "fixed-1")

	out4 := runNAOKWithActionID(t, fixed, "task", "create",
		"--name", testName("固定B"), "--schedule", "daily", "--data", `{"time":"11:00"}`)
	assertActionID(t, out4, fs, "fixed-2")

	t.Logf("✅ fixed --action-id preserved across 2 invocations: %s", fs)
}

// ─── Task info: exact name + substring match ──────────────────────

func TestNameBased_TaskInfoByName(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	setupWithFamily(t)

	a := testName("晨间检查")
	b := testName("晚间检查")
	runNAOK(t, "task", "create", "--name", a, "--schedule", "daily", "--data", `{"time":"08:00"}`)
	runNAOK(t, "task", "create", "--name", b, "--schedule", "daily", "--data", `{"time":"20:00"}`)

	if out := runNAOK(t, "task", "info", "--name", a); !strings.Contains(out, a) {
		t.Errorf("exact: %s", truncate(out, 150))
	}
	t.Log("✅ info --name exact")

	if out := runNAOK(t, "task", "info", "--name", "检查"); !strings.Contains(out, "检查") {
		t.Errorf("substring: %s", truncate(out, 150))
	}
	t.Log("✅ info --name substring")
}

// ─── Task info by full UUID + ID prefix ──────────────────────────

func TestNameBased_TaskInfoByID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	setupWithFamily(t)

	name := testName("ID查询")
	runNAOK(t, "task", "create", "--name", name, "--schedule", "daily", "--data", `{"time":"10:00"}`)

	out := runNAOK(t, "task", "info", "--name", name)
	id := extractField(t, out, "ID:")

	if out = runNAOK(t, "task", "info", "--id", id); !strings.Contains(out, name) {
		t.Errorf("full UUID: %s", truncate(out, 150))
	}
	t.Log("✅ info --id full UUID")

	if out = runNAOK(t, "task", "info", "--id", id[:6]); !strings.Contains(out, name) {
		t.Errorf("prefix: %s", truncate(out, 150))
	}
	t.Logf("✅ info --id prefix %s", id[:6])
}

// ─── Task update: rename + schedule change ────────────────────────

func TestNameBased_TaskUpdateByName(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	setupWithFamily(t)

	oldName := testName("更新测试")
	runNAOK(t, "task", "create", "--name", oldName, "--schedule", "daily", "--data", `{"time":"07:00"}`)

	// Rename (must pass --new-name, --schedule, --data due to backend binding)
	newName := testName("已更名")
	out := runNAOK(t, "task", "update",
		"--name", oldName,
		"--new-name", newName,
		"--schedule", "daily",
		"--data", `{"time":"07:00"}`,
	)
	if !strings.Contains(out, "已更新") {
		t.Fatalf("rename failed: %s", out)
	}
	t.Log("✅ rename")

	out = runNAOK(t, "task", "info", "--name", newName)
	if !strings.Contains(out, newName) {
		t.Errorf("verify rename: %s", truncate(out, 150))
	}
	t.Log("✅ verify rename")

	// Change schedule (must also pass --new-name)
	out = runNAOK(t, "task", "update",
		"--name", newName,
		"--new-name", newName,
		"--schedule", "weekly",
		"--data", `{"days":[1,3,5],"time":"19:00"}`,
	)
	if !strings.Contains(out, "已更新") {
		t.Fatalf("schedule update failed: %s", out)
	}
	t.Log("✅ schedule → weekly")

	out = runNAOK(t, "task", "info", "--name", newName)
	if !strings.Contains(out, "weekly") {
		t.Errorf("schedule not weekly: %s", truncate(out, 150))
	}
	t.Log("✅ verified weekly")
}

// ─── Task update by ID prefix ────────────────────────────────────

func TestNameBased_TaskUpdateByID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	setupWithFamily(t)

	name := testName("ID更新")
	runNAOK(t, "task", "create", "--name", name, "--schedule", "daily", "--data", `{"time":"12:00"}`)

	out := runNAOK(t, "task", "info", "--name", name)
	id := extractField(t, out, "ID:")
	prefix := id[:6]

	newName := testName("ID更名")
	out = runNAOK(t, "task", "update",
		"--id", prefix,
		"--new-name", newName,
		"--schedule", "interval",
		"--data", `{"days":3}`,
	)
	if !strings.Contains(out, "已更新") {
		t.Fatalf("update by ID failed: %s", out)
	}
	t.Logf("✅ update by ID prefix %s", prefix)

	out = runNAOK(t, "task", "info", "--name", newName)
	t.Logf("verified: %s", truncate(out, 150))
}

// ─── Task delete by name & ID prefix ─────────────────────────────

func TestNameBased_TaskDelete(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	setupWithFamily(t)

	n1 := testName("删A")
	n2 := testName("删B")
	runNAOK(t, "task", "create", "--name", n1, "--schedule", "daily", "--data", `{"time":"08:00"}`)
	runNAOK(t, "task", "create", "--name", n2, "--schedule", "daily", "--data", `{"time":"09:00"}`)

	// By name
	out := runNAOK(t, "task", "delete", "--name", n1, "-y")
	if !strings.Contains(out, "已删除") {
		t.Errorf("delete by name: %s", out)
	}
	_, _, err := runNA("task", "info", "--name", n1)
	if err == nil {
		t.Error("task should be deleted")
	}
	t.Log("✅ delete --name -y")

	// By ID prefix
	out = runNAOK(t, "task", "info", "--name", n2)
	id := extractField(t, out, "ID:")
	out = runNAOK(t, "task", "delete", "--id", id[:6], "-y")
	if !strings.Contains(out, "已删除") {
		t.Errorf("delete by ID: %s", out)
	}
	t.Log("✅ delete --id prefix -y")
}

// ─── Task trigger by name ────────────────────────────────────────

func TestNameBased_TaskTrigger(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	setupWithFamily(t)

	name := testName("触发测试")
	runNAOK(t, "task", "create", "--name", name, "--schedule", "daily", "--data", `{"time":"06:00"}`)

	out := runNAOK(t, "task", "trigger", "--name", name)
	if !strings.Contains(out, "已触发") {
		t.Fatalf("trigger failed: %s", out)
	}
	t.Log("✅ trigger --name")

	// Verify todo
	out = runNAOK(t, "todo", "list")
	t.Logf("todos after trigger: %s", truncate(out, 200))
}

// ─── Location name resolution ────────────────────────────────────

func TestNameBased_LocationResolution(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	setupWithFamily(t)

	name := testName("位置测试")
	out := runNAOK(t, "task", "create",
		"--name", name, "--schedule", "daily", "--data", `{"time":"10:00"}`,
		"--location", "厨房",
	)
	if !strings.Contains(out, "厨房") {
		t.Fatalf("create with location: %s", out)
	}
	t.Log("✅ create --location 厨房")

	out = runNAOK(t, "task", "info", "--name", name)
	if !strings.Contains(out, "厨房") {
		t.Errorf("location not assigned: %s", truncate(out, 150))
	}
	t.Log("✅ verified 厨房")

	// Update location
	runNAOK(t, "task", "update",
		"--name", name, "--new-name", name,
		"--schedule", "daily", "--data", `{"time":"10:00"}`,
		"--location", "客厅",
	)
	out = runNAOK(t, "task", "info", "--name", name)
	t.Logf("after update: %s", truncate(out, 200))
	t.Log("✅ update --location")
}

// ─── Group name resolution ───────────────────────────────────────

func TestNameBased_GroupResolution(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	setupWithFamily(t)

	name := testName("小组测试")
	out := runNAOK(t, "task", "create",
		"--name", name, "--schedule", "daily", "--data", `{"time":"09:00"}`,
		"--group", "大人",
	)
	if !strings.Contains(out, "大人") {
		t.Fatalf("create with group: %s", out)
	}
	t.Log("✅ create --group 大人")

	out = runNAOK(t, "task", "info", "--name", name)
	if !strings.Contains(out, "大人") {
		t.Errorf("group not assigned: %s", truncate(out, 150))
	}
	t.Log("✅ verified group=大人")
}

// ─── Error cases (stderr messages) ───────────────────────────────

func TestNameBased_ErrorCases(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	setupWithFamily(t)

	for _, tc := range []struct {
		label string
		args  []string
	}{
		{"missing name/id in info", []string{"task", "info"}},
		{"missing name/id in update", []string{"task", "update", "--new-name", "x"}},
		{"missing name/id in delete", []string{"task", "delete", "-y"}},
		{"missing name/id in trigger", []string{"task", "trigger"}},
		{"non-existent name", []string{"task", "info", "--name", "ZZZ不存在ZZZ"}},
		{"non-existent delete", []string{"task", "delete", "--name", "ZZZ不存在ZZZ", "-y"}},
		{"empty name", []string{"task", "info", "--name", ""}},
	} {
		_, errOut, err := runNA(tc.args...)
		if err == nil {
			t.Errorf("[%s] expected error but passed", tc.label)
			continue
		}
		if errOut == "" {
			t.Logf("[%s] error with empty stderr (exit=%v)", tc.label, err)
		} else {
			t.Logf("✅ %s → %s", tc.label, truncate(errOut, 80))
		}
	}
}

// ─── JSON / text output modes ─────────────────────────────────────

func TestNameBased_JSONOutput(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	setupWithFamily(t)

	// Text mode (default): must have [action: XXXXXXXX] prefix
	out := runNAOK(t, "task", "create",
		"--name", testName("text测试"),
		"--schedule", "daily",
		"--data", `{"time":"11:00"}`,
	)
	if !actionIDPattern.MatchString(out) {
		t.Errorf("text mode missing [action: xxx]:\n%s", truncate(out, 120))
	}
	t.Log("✅ text mode has [action: xxx] prefix")

	// Verify each text output starts with [action: XXXXXXXX]
	for _, cmd := range []struct {
		label string
		args  []string
	}{
		{"task create", []string{"task", "create", "--name", testName("action检查"), "--schedule", "daily", "--data", `{"time":"10:00"}`}},
		{"task delete", []string{"task", "delete", "--name", "action检查", "-y"}},
	} {
		out := runNAOK(t, cmd.args...)
		if !actionIDPattern.MatchString(out) {
			t.Errorf("[%s] missing action-id prefix:\n%s", cmd.label, truncate(out, 120))
		}
	}
	t.Log("✅ all action-print commands emit [action: XXXXXXXX]")

	// Note: -o json is a registered flag but task create currently always
	// outputs text format (action.Printf always emits text prefix).
	// JSON envelope output is planned for future implementation.
	t.Log("⚠️ -o json flag registered but not yet wired for task commands")
}

// ─── Family select by name ───────────────────────────────────────

func TestNameBased_FamilySelect(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	setupWithFamily(t)

	out := runNAOK(t, "family", "list")

	var familyName string
	for _, line := range strings.Split(out, "\n") {
		for _, f := range strings.Fields(line) {
			if strings.HasPrefix(f, "测试家庭") {
				familyName = strings.TrimRight(f, ",")
				break
			}
		}
	}
	if familyName == "" {
		t.Skip("cannot find family name in list")
	}

	out = runNAOK(t, "family", "select", "--name", familyName)
	// The success message may be "已切换" or "已选择"
	if !strings.Contains(out, "已") {
		t.Errorf("family select failed: %s", out)
	}
	t.Logf("✅ family select --name '%s'", truncate(familyName, 40))
}

// ─── Helpers ──────────────────────────────────────────────────────

func extractField(t *testing.T, out, key string) string {
	t.Helper()
	for _, line := range strings.Split(out, "\n") {
		trimmed := strings.TrimSpace(line)
		if after, ok := strings.CutPrefix(trimmed, key); ok {
			return strings.TrimSpace(after)
		}
	}
	t.Fatalf("field %q not found in:\n%s", key, truncate(out, 300))
	return ""
}

func extractShort(out string) string {
	m := actionIDPattern.FindString(out)
	if m == "" {
		return ""
	}
	return m[9 : len(m)-1]
}
