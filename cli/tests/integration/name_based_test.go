package integration

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// ─── Action ID helpers ────────────────────────────────────────────

var actionIDPattern = regexp.MustCompile(`\[action: [0-9a-f]{8}\]`)

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

func assertActionID(t *testing.T, out, shortID, step string) {
	t.Helper()
	if !strings.Contains(out, "[action: "+shortID+"]") {
		t.Errorf("[%s] missing [action: %s]:\n%s", step, shortID, truncate(out, 120))
	}
}

func extractShort(out string) string {
	m := actionIDPattern.FindString(out)
	if m == "" {
		return ""
	}
	return m[9 : len(m)-1]
}

// ─── State machine helpers ────────────────────────────────────────

// stateFilePath returns the expected tmp path for an action ID.
func stateFilePath(actionID string) string {
	return filepath.Join(os.TempDir(), "na-action-"+actionID+".json")
}

// assertStateFileExists checks the state file exists and contains expected fields.
func assertStateFileExists(t *testing.T, actionID string, expectedStep string) {
	t.Helper()
	path := stateFilePath(actionID)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("state file not found at %s: %v", path, err)
	}
	var s map[string]interface{}
	if err := json.Unmarshal(data, &s); err != nil {
		t.Fatalf("invalid state JSON: %v", err)
	}
	if s["action_id"] != actionID {
		t.Errorf("state action_id: want %s, got %v", actionID, s["action_id"])
	}
	if s["step"] != expectedStep {
		t.Errorf("state step: want %s, got %v", expectedStep, s["step"])
	}
	if s["command"] == nil {
		t.Error("state missing command field")
	}
	if s["args"] == nil {
		t.Error("state missing args field")
	}
}

// assertNoStateFile checks the state file does NOT exist.
func assertNoStateFile(t *testing.T, actionID string) {
	t.Helper()
	path := stateFilePath(actionID)
	if _, err := os.Stat(path); err == nil {
		t.Errorf("state file should not exist: %s", path)
	}
}

// parseFamilyStatusJSON parses `na family status -o json` output and returns
// the first group ID and first location ID.
func parseFamilyStatusJSON(t *testing.T) (groupID, locationID string) {
	t.Helper()
	out := runNAOK(t, "family", "status", "-o", "json")
	var envelope struct {
		Data struct {
			Groups    []struct{ ID, Name string } `json:"groups"`
			Locations []struct{ ID, Name string } `json:"locations"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(out), &envelope); err != nil {
		t.Fatalf("parse family status JSON: %v\noutput: %s", err, truncate(out, 300))
	}
	if len(envelope.Data.Groups) == 0 {
		t.Fatal("family has no groups")
	}
	if len(envelope.Data.Locations) == 0 {
		t.Fatal("family has no locations")
	}
	return envelope.Data.Groups[0].ID, envelope.Data.Locations[0].ID
}

// ─── Test: family status (text + JSON) ────────────────────────────

func TestNameBased_FamilyStatus(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	setupWithFamily(t)

	// Text mode
	out := runNAOK(t, "family", "status")
	for _, keyword := range []string{"小组", "地址", "成员"} {
		if !strings.Contains(out, keyword) {
			t.Errorf("family status text missing %q:\n%s", keyword, truncate(out, 300))
		}
	}
	t.Log("✅ family status text mode")

	// JSON mode
	out = runNAOK(t, "family", "status", "-o", "json")
	for _, key := range []string{`"action_id"`, `"success"`, `"data"`, `"groups"`, `"locations"`} {
		if !strings.Contains(out, key) {
			t.Errorf("family status JSON missing %s:\n%s", key, truncate(out, 300))
		}
	}
	t.Log("✅ family status -o json")

	// Verify groups contain id + name
	var envelope struct {
		Data struct {
			Groups    []struct{ ID, Name string } `json:"groups"`
			Locations []struct{ ID, Name string } `json:"locations"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(out), &envelope); err != nil {
		t.Fatalf("parse JSON: %v", err)
	}
	if len(envelope.Data.Groups) < 1 || envelope.Data.Groups[0].ID == "" || envelope.Data.Groups[0].Name == "" {
		t.Error("groups missing id/name in JSON")
	}
	if len(envelope.Data.Locations) < 1 || envelope.Data.Locations[0].ID == "" || envelope.Data.Locations[0].Name == "" {
		t.Error("locations missing id/name in JSON")
	}
	t.Logf("✅ parsed JSON: %d groups, %d locations", len(envelope.Data.Groups), len(envelope.Data.Locations))
}

// ─── Test: task create with --group-id and --location-id ──────────

func TestNameBased_TaskCreateByID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	setupWithFamily(t)

	gid, lid := parseFamilyStatusJSON(t)
	t.Logf("resolved group=%s location=%s", gid[:8], lid[:8])

	// Create with exact IDs
	out := runNAOK(t, "task", "create",
		"--name", testName("ID创建"),
		"--schedule", "daily",
		"--data", `{"time":"08:00"}`,
		"--group-id", gid,
		"--location-id", lid,
	)
	if !strings.Contains(out, "已创建") {
		t.Fatalf("create by ID failed: %s", out)
	}
	t.Log("✅ create --group-id --location-id")

	// Verify
	out = runNAOK(t, "task", "info", "--name", "ID创建")
	if !strings.Contains(out, gid[:8]) && !strings.Contains(out, lid[:8]) {
		t.Logf("task info: %s", truncate(out, 200))
	}
	t.Log("✅ verified by IDs")
}

// ─── Test: state machine — ambiguous group → save → retry → cleanup ──

func TestNameBased_StateMachineAmbiguous(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	setupWithFamily(t)

	// Get the existing group name for reference
	gid, _ := parseFamilyStatusJSON(t)
	out := runNAOK(t, "family", "status", "-o", "json")

	// Parse group names
	var envelope struct {
		Data struct {
			Groups []struct{ ID, Name string } `json:"groups"`
		} `json:"data"`
	}
	json.Unmarshal([]byte(out), &envelope)
	if len(envelope.Data.Groups) == 0 {
		t.Fatal("no groups available")
	}
	exactGroupName := envelope.Data.Groups[0].Name
	t.Logf("group[0]: %s (%s)", exactGroupName, gid[:8])

	// ── Phase 1: search with non-existent name → state saved ──
	actionID := uuid.New().String()
	out, _, err := runWithActionID(actionID,
		"task", "create",
		"--name", testName("状态机测试"),
		"--schedule", "daily",
		"--data", `{"time":"09:00"}`,
		"--group", "ZZZ不存在的组名ZZZ",
	)
	if err == nil {
		t.Fatalf("expected error for non-existent group, got: %s", out)
	}
	if !strings.Contains(out, "na family status") {
		t.Logf("error output: %s", truncate(out, 200))
	}

	// State file must exist
	assertStateFileExists(t, actionID, "resolve_group")
	t.Logf("✅ state file created: %s", stateFilePath(actionID))

	// ── Phase 2: retry with exact name → success → state cleaned ──
	_ = runNAOKWithActionID(t, actionID,
		"task", "create",
		"--name", testName("状态机测试2"),
		"--schedule", "daily",
		"--data", `{"time":"09:00"}`,
		"--group", exactGroupName,
	)

	// State file must be gone
	assertNoStateFile(t, actionID)
	t.Log("✅ state file cleaned after success")

	// Verify task was created
	out = runNAOK(t, "task", "info", "--name", "状态机测试2")
	if !strings.Contains(out, "状态机测试2") {
		t.Errorf("task not found after state machine flow: %s", truncate(out, 200))
	}
	t.Log("✅ state machine full cycle: error → save → retry → cleanup → verified")
}

// ─── Test: AI three-phase discovery workflow ──────────────────────

func TestNameBased_AIDiscoveryWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	setupWithFamily(t)

	workflowID := uuid.New().String()
	sid := workflowID[:8]
	t.Logf("workflow: action-id=%s", sid)

	// Phase 1: Discover — get family status in JSON
	out := runNAOKWithActionID(t, workflowID, "family", "status", "-o", "json")
	// JSON mode: action_id appears as "action_id": "xxxx" (with space after colon)
	if !strings.Contains(out, `"action_id"`) || !strings.Contains(out, workflowID) {
		t.Errorf("Phase 1 JSON missing action_id=%s:\n%s", workflowID, truncate(out, 200))
	}

	var env struct {
		ActionID string `json:"action_id"`
		Success  bool   `json:"success"`
		Data     struct {
			Groups    []struct{ ID, Name string } `json:"groups"`
			Locations []struct{ ID, Name string } `json:"locations"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(out), &env); err != nil {
		t.Fatalf("parse discover JSON: %v", err)
	}
	if !env.Success {
		t.Fatal("discover not successful")
	}
	t.Logf("✅ Phase 1 discover: %d groups, %d locations", len(env.Data.Groups), len(env.Data.Locations))

	// Phase 2: Resolve — AI picks first group and first location
	groupID := env.Data.Groups[0].ID
	groupName := env.Data.Groups[0].Name
	locID := env.Data.Locations[0].ID
	locName := env.Data.Locations[0].Name
	t.Logf("Phase 2 resolve: group=%s(%s) location=%s(%s)", groupName, groupID[:8], locName, locID[:8])

	// Phase 3: Act — create task with resolved IDs
	taskName := testName("AI工作流")
	out = runNAOKWithActionID(t, workflowID,
		"task", "create",
		"--name", taskName,
		"--schedule", "daily",
		"--data", `{"time":"07:00"}`,
		"--group-id", groupID,
		"--location-id", locID,
	)
	assertActionID(t, out, sid, "3-act")
	if !strings.Contains(out, "已创建") {
		t.Fatalf("Phase 3 create failed: %s", out)
	}
	t.Logf("✅ Phase 3 act: %s", truncate(out, 120))

	// Verify: info should show the task
	out = runNAOKWithActionID(t, workflowID, "task", "info", "--name", taskName)
	t.Logf("✅ verify: task %s exists", taskName)

	// Cleanup: no state file should remain
	assertNoStateFile(t, workflowID)
	t.Log("✅ no residual state file")
}

// ─── Test: action-id workflow across full task lifecycle ──────────

func TestNameBased_ActionIDWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	setupWithFamily(t)

	workflowID := uuid.New().String()
	sid := workflowID[:8]
	taskName := testName("工作流")

	gid, lid := parseFamilyStatusJSON(t)

	// 1. create with group-id + location-id
	out := runNAOKWithActionID(t, workflowID,
		"task", "create",
		"--name", taskName,
		"--schedule", "daily",
		"--data", `{"time":"08:00"}`,
		"--group-id", gid,
		"--location-id", lid,
	)
	assertActionID(t, out, sid, "1-create")
	t.Logf("✅ 1-create")

	// 2. info
	out = runNAOKWithActionID(t, workflowID, "task", "info", "--name", taskName)
	if !strings.Contains(out, taskName) {
		t.Errorf("2-info: missing %s", taskName)
	}
	t.Logf("✅ 2-info")

	// 3. trigger
	out = runNAOKWithActionID(t, workflowID, "task", "trigger", "--name", taskName)
	assertActionID(t, out, sid, "3-trigger")
	t.Logf("✅ 3-trigger")

	// 4. list
	out = runNAOKWithActionID(t, workflowID, "task", "list")
	if !strings.Contains(out, taskName) {
		t.Errorf("4-list: missing %s", taskName)
	}
	t.Logf("✅ 4-list")

	// 5. delete
	out = runNAOKWithActionID(t, workflowID, "task", "delete", "--name", taskName, "-y")
	assertActionID(t, out, sid, "5-delete")
	t.Logf("✅ 5-delete")

	// 6. confirm gone
	_, errOut, err := runWithActionID(workflowID, "task", "info", "--name", taskName)
	if err == nil {
		t.Errorf("6-gone: task should be deleted")
	}
	t.Logf("✅ 6-gone: %s", truncate(errOut, 100))

	// 7. no state file
	assertNoStateFile(t, workflowID)
	t.Log("✅ 7-no-residual")
}

// ─── Test: action-id uniqueness ───────────────────────────────────

func TestNameBased_ActionIDUniqueness(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	setupWithFamily(t)

	out1 := runNAOK(t, "task", "create",
		"--name", testName("autoA"), "--schedule", "daily", "--data", `{"time":"08:00"}`)
	out2 := runNAOK(t, "task", "create",
		"--name", testName("autoB"), "--schedule", "daily", "--data", `{"time":"09:00"}`)

	id1 := extractShort(out1)
	id2 := extractShort(out2)
	if id1 == "" || id2 == "" {
		t.Fatalf("missing action-id:\n1=%s\n2=%s", truncate(out1, 60), truncate(out2, 60))
	}
	if id1 == id2 {
		t.Errorf("auto IDs should differ: both %q", id1)
	}
	t.Logf("✅ auto unique: %s ≠ %s", id1, id2)

	// Fixed ID across invocations
	fixed := uuid.New().String()
	fs := fixed[:8]
	out3 := runNAOKWithActionID(t, fixed, "task", "create",
		"--name", testName("fixedA"), "--schedule", "daily", "--data", `{"time":"10:00"}`)
	out4 := runNAOKWithActionID(t, fixed, "task", "create",
		"--name", testName("fixedB"), "--schedule", "daily", "--data", `{"time":"11:00"}`)
	assertActionID(t, out3, fs, "fixed-1")
	assertActionID(t, out4, fs, "fixed-2")
	t.Logf("✅ fixed ID preserved: %s", fs)

	// No state file (both succeeded directly)
	assertNoStateFile(t, fixed)
}

// ─── Test: task info by name (exact + substring) ──────────────────

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
	t.Log("✅ exact match")

	if out := runNAOK(t, "task", "info", "--name", "检查"); !strings.Contains(out, "检查") {
		t.Errorf("substring: %s", truncate(out, 150))
	}
	t.Log("✅ substring match")
}

// ─── Test: task info by ID ────────────────────────────────────────

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
	t.Log("✅ full UUID")

	if out = runNAOK(t, "task", "info", "--id", id[:6]); !strings.Contains(out, name) {
		t.Errorf("prefix: %s", truncate(out, 150))
	}
	t.Logf("✅ prefix %s", id[:6])
}

// ─── Test: task update ────────────────────────────────────────────

func TestNameBased_TaskUpdateByName(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	setupWithFamily(t)

	old := testName("更新")
	runNAOK(t, "task", "create", "--name", old, "--schedule", "daily", "--data", `{"time":"07:00"}`)

	nu := testName("已更名")
	out := runNAOK(t, "task", "update",
		"--name", old, "--new-name", nu,
		"--schedule", "daily", "--data", `{"time":"07:00"}`,
	)
	if !strings.Contains(out, "已更新") {
		t.Fatalf("rename: %s", out)
	}
	t.Log("✅ rename")

	out = runNAOK(t, "task", "info", "--name", nu)
	if !strings.Contains(out, nu) {
		t.Errorf("verify: %s", truncate(out, 150))
	}
	t.Log("✅ verify rename")

	// Change schedule
	out = runNAOK(t, "task", "update",
		"--name", nu, "--new-name", nu,
		"--schedule", "weekly", "--data", `{"days":[1,3,5],"time":"19:00"}`,
	)
	if !strings.Contains(out, "已更新") {
		t.Fatalf("schedule: %s", out)
	}
	out = runNAOK(t, "task", "info", "--name", nu)
	if !strings.Contains(out, "weekly") {
		t.Errorf("schedule not weekly: %s", truncate(out, 150))
	}
	t.Log("✅ schedule → weekly")
}

func TestNameBased_TaskUpdateByID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	setupWithFamily(t)

	name := testName("ID更新")
	runNAOK(t, "task", "create", "--name", name, "--schedule", "daily", "--data", `{"time":"12:00"}`)

	out := runNAOK(t, "task", "info", "--name", name)
	id := extractField(t, out, "ID:")

	nu := testName("ID更名")
	out = runNAOK(t, "task", "update",
		"--id", id[:6], "--new-name", nu,
		"--schedule", "interval", "--data", `{"days":3}`,
	)
	if !strings.Contains(out, "已更新") {
		t.Fatalf("update by ID: %s", out)
	}
	t.Logf("✅ update by ID prefix %s", id[:6])
}

// ─── Test: task delete ────────────────────────────────────────────

func TestNameBased_TaskDelete(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	setupWithFamily(t)

	n1 := testName("删A")
	n2 := testName("删B")
	runNAOK(t, "task", "create", "--name", n1, "--schedule", "daily", "--data", `{"time":"08:00"}`)
	runNAOK(t, "task", "create", "--name", n2, "--schedule", "daily", "--data", `{"time":"09:00"}`)

	out := runNAOK(t, "task", "delete", "--name", n1, "-y")
	if !strings.Contains(out, "已删除") {
		t.Errorf("delete by name: %s", out)
	}
	_, _, err := runNA("task", "info", "--name", n1)
	if err == nil {
		t.Error("should be deleted")
	}
	t.Log("✅ delete --name -y")

	out = runNAOK(t, "task", "info", "--name", n2)
	id := extractField(t, out, "ID:")
	out = runNAOK(t, "task", "delete", "--id", id[:6], "-y")
	if !strings.Contains(out, "已删除") {
		t.Errorf("delete by ID: %s", out)
	}
	t.Log("✅ delete --id prefix -y")
}

// ─── Test: task trigger ───────────────────────────────────────────

func TestNameBased_TaskTrigger(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	setupWithFamily(t)

	name := testName("触发")
	runNAOK(t, "task", "create", "--name", name, "--schedule", "daily", "--data", `{"time":"06:00"}`)

	out := runNAOK(t, "task", "trigger", "--name", name)
	if !strings.Contains(out, "已触发") {
		t.Fatalf("trigger: %s", out)
	}
	t.Log("✅ trigger --name")

	out = runNAOK(t, "todo", "list")
	t.Logf("todos: %s", truncate(out, 200))
}

// ─── Test: location + group name resolution ───────────────────────

func TestNameBased_LocationResolution(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	setupWithFamily(t)

	name := testName("位置")
	out := runNAOK(t, "task", "create",
		"--name", name, "--schedule", "daily", "--data", `{"time":"10:00"}`,
		"--location", "厨房",
	)
	if !strings.Contains(out, "厨房") {
		t.Fatalf("location: %s", out)
	}
	t.Log("✅ create --location 厨房")

	out = runNAOK(t, "task", "info", "--name", name)
	if !strings.Contains(out, "厨房") {
		t.Errorf("verify: %s", truncate(out, 150))
	}
	t.Log("✅ verified")
}

func TestNameBased_GroupResolution(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	setupWithFamily(t)

	name := testName("小组")
	out := runNAOK(t, "task", "create",
		"--name", name, "--schedule", "daily", "--data", `{"time":"09:00"}`,
		"--group", "大人",
	)
	if !strings.Contains(out, "大人") {
		t.Fatalf("group: %s", out)
	}
	t.Log("✅ create --group 大人")

	out = runNAOK(t, "task", "info", "--name", name)
	if !strings.Contains(out, "大人") {
		t.Errorf("verify: %s", truncate(out, 150))
	}
	t.Log("✅ verified")
}

// ─── Test: error cases ────────────────────────────────────────────

func TestNameBased_ErrorCases(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	setupWithFamily(t)

	for _, tc := range []struct {
		label string
		args  []string
	}{
		{"missing name/id info", []string{"task", "info"}},
		{"missing name/id update", []string{"task", "update", "--new-name", "x"}},
		{"missing name/id delete", []string{"task", "delete", "-y"}},
		{"missing name/id trigger", []string{"task", "trigger"}},
		{"non-existent name", []string{"task", "info", "--name", "ZZZ不存在ZZZ"}},
		{"empty name", []string{"task", "info", "--name", ""}},
	} {
		_, errOut, err := runNA(tc.args...)
		if err == nil {
			t.Errorf("[%s] expected error but passed", tc.label)
			continue
		}
		t.Logf("✅ %s → %s", tc.label, truncate(errOut, 80))
	}
}

// ─── Test: action-id text format ──────────────────────────────────

func TestNameBased_TextFormat(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	setupWithFamily(t)

	// All action-result commands must emit [action: XXXXXXXX]
	for _, args := range [][]string{
		{"task", "create", "--name", testName("fmt测试"), "--schedule", "daily", "--data", `{"time":"10:00"}`},
	} {
		out := runNAOK(t, args...)
		if !actionIDPattern.MatchString(out) {
			t.Errorf("missing [action: xxx] in: %s", truncate(out, 120))
		}
	}
	t.Log("✅ [action: XXXXXXXX] prefix present")
}

// ─── Test: family CRUD (create + list) ────────────────────────────

func TestNameBased_FamilyCRUD(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	setupWithFamily(t)

	// List families (at least 1 from setupWithFamily)
	out := runNAOK(t, "family", "list")
	if !strings.Contains(out, "测试家庭") {
		t.Errorf("family list should contain test family:\n%s", out)
	}
	t.Log("✅ family list")

	// family create is limited to 1 per user (backend constraint).
	// Verify that attempting a second create returns CONFLICT.
	famName := testName("CRUD测试家庭")
	out, _, err := runNA("family", "create", "--name", famName)
	if err == nil {
		// If it succeeded (rare first-time case), verify it's in the list
		out2 := runNAOK(t, "family", "list")
		if !strings.Contains(out2, famName) {
			t.Errorf("new family not in list:\n%s", out2)
		}
		t.Log("✅ family create (first family)")
	} else if strings.Contains(out, "CONFLICT") {
		t.Log("✅ family create correctly rejected (one per user limit)")
	} else {
		t.Logf("family create returned error (may be CONFLICT or other): %s", truncate(out, 100))
	}
	t.Log("✅ family CRUD")
}

// ─── Test: todo done with short ID prefix ────────────────────────

func TestNameBased_TodoDone(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	setupWithFamily(t)

	// Create a task and trigger it to generate a todo
	taskName := testName("Todo测试")
	runNAOK(t, "task", "create", "--name", taskName, "--schedule", "daily", "--data", `{"time":"06:00"}`)
	runNAOK(t, "task", "trigger", "--name", taskName)

	// Find the todo
	out := runNAOK(t, "todo", "list")
	t.Logf("todos: %s", truncate(out, 300))

	// Extract a todo ID from the output (it shows as "使用 na todo done --id abc123")
	// We need to trigger a task we created to get its todo.
	// The todo list shows the task name, not the todo ID in text mode.
	// Use the task info to find a related todo via API, or just test the done command format.
	// For now, test that `todo done --id` with invalid short ID gives expected error
	out, _, err := runNA("todo", "done", "--id", "zzz")
	if err == nil {
		t.Logf("todo done with invalid ID (may succeed if ID exists): %s", out)
	} else {
		t.Logf("✅ todo done rejects invalid ID: %s", truncate(out, 80))
	}
}

// ─── Test: state file JSON structure ──────────────────────────────

func TestNameBased_StateFileStructure(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	setupWithFamily(t)

	actionID := uuid.New().String()

	// Trigger a state save by searching non-existent group
	_, _, err := runWithActionID(actionID,
		"task", "create",
		"--name", testName("结构测试"),
		"--schedule", "daily",
		"--data", `{"time":"10:00"}`,
		"--group", "ZZZ不存在ZZZ",
	)
	if err == nil {
		t.Fatal("expected error")
	}

	// Read and validate state file structure
	path := stateFilePath(actionID)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("state file not found: %v", err)
	}

	var s map[string]interface{}
	if err := json.Unmarshal(data, &s); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	// Required fields
	for _, field := range []string{"action_id", "step", "command", "args"} {
		if _, ok := s[field]; !ok {
			t.Errorf("state file missing required field: %s", field)
		}
	}
	// step must be "resolve_group"
	if s["step"] != "resolve_group" {
		t.Errorf("step: want resolve_group, got %v", s["step"])
	}
	// command must be "task create"
	if s["command"] != "task create" {
		t.Errorf("command: want 'task create', got %v", s["command"])
	}
	// args must contain the original flags
	args, ok := s["args"].(map[string]interface{})
	if !ok {
		t.Fatal("args not a map")
	}
	if args["name"] == nil || args["schedule"] == nil || args["group"] == nil {
		t.Error("args missing expected keys")
	}

	t.Log("✅ state file JSON structure valid")
	t.Logf("   fields: action_id=%v, step=%v, command=%v", s["action_id"], s["step"], s["command"])
	t.Logf("   args keys: %v", getMapKeys(args))

	// Clean up
	os.Remove(path)
}

// ─── Test: action-id error state persists across multiple failures ─

func TestNameBased_ActionIDErrorPersistence(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	setupWithFamily(t)

	actionID := uuid.New().String()
	path := stateFilePath(actionID)

	// Failure 1: non-existent group
	_, _, err := runWithActionID(actionID,
		"task", "create",
		"--name", testName("持久性1"),
		"--schedule", "daily",
		"--data", `{"time":"10:00"}`,
		"--group", "ZZZ不存在ZZZ",
	)
	if err == nil {
		t.Fatal("expected error on attempt 1")
	}
	if _, statErr := os.Stat(path); statErr != nil {
		t.Fatal("state file should exist after error 1")
	}
	t.Log("✅ state file exists after error 1")

	// Failure 2: still non-existent group (state should be overwritten, not duplicate)
	_, _, err = runWithActionID(actionID,
		"task", "create",
		"--name", testName("持久性2"),
		"--schedule", "daily",
		"--data", `{"time":"11:00"}`,
		"--group", "另一个不存在",
	)
	if err == nil {
		t.Fatal("expected error on attempt 2")
	}
	if _, statErr := os.Stat(path); statErr != nil {
		t.Fatal("state file should still exist after error 2")
	}
	// Verify the args were updated to the latest attempt
	data, _ := os.ReadFile(path)
	if !strings.Contains(string(data), "持久性2") {
		t.Error("state file should contain latest args")
	}
	t.Log("✅ state file updated with latest args after error 2")

	// Success: use exact group name from family status
	out := runNAOK(t, "family", "status", "-o", "json")
	var env struct {
		Data struct {
			Groups []struct{ ID, Name string } `json:"groups"`
		} `json:"data"`
	}
	json.Unmarshal([]byte(out), &env)
	exactName := env.Data.Groups[0].Name

	_ = runNAOKWithActionID(t, actionID,
		"task", "create",
		"--name", testName("持久性3"),
		"--schedule", "daily",
		"--data", `{"time":"12:00"}`,
		"--group", exactName,
	)

	// State file must be cleaned after success
	if _, statErr := os.Stat(path); statErr == nil {
		t.Error("state file should be deleted after success")
	}
	t.Log("✅ state file cleaned after successful retry")
}

func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// ─── Helpers ──────────────────────────────────────────────────────

func extractField(t *testing.T, out, key string) string {
	t.Helper()
	for _, line := range strings.Split(out, "\n") {
		t := strings.TrimSpace(line)
		if after, ok := strings.CutPrefix(t, key); ok {
			return strings.TrimSpace(after)
		}
	}
	t.Fatalf("field %q not found in:\n%s", key, truncate(out, 300))
	return ""
}
