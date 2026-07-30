package state

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSave_And_Load(t *testing.T) {
	s := &ActionState{
		ActionID: "test-save-load",
		Step:     "resolve_group",
		Command:  "task create",
		Args:     map[string]string{"name": "测试任务"},
		Candidates: []EntityOption{
			{ID: "g1", Name: "大人"},
			{ID: "g2", Name: "小孩"},
		},
		Resolved: map[string]string{"group": "g1"},
	}

	if err := Save(s); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := Load("test-save-load")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if loaded.ActionID != "test-save-load" {
		t.Errorf("expected ActionID 'test-save-load', got %q", loaded.ActionID)
	}
	if loaded.Step != "resolve_group" {
		t.Errorf("expected Step 'resolve_group', got %q", loaded.Step)
	}
	if loaded.Command != "task create" {
		t.Errorf("expected Command 'task create', got %q", loaded.Command)
	}
	if len(loaded.Candidates) != 2 {
		t.Fatalf("expected 2 candidates, got %d", len(loaded.Candidates))
	}
	if loaded.Candidates[0].ID != "g1" || loaded.Candidates[0].Name != "大人" {
		t.Errorf("unexpected candidate[0]: %+v", loaded.Candidates[0])
	}
	if loaded.Candidates[1].ID != "g2" || loaded.Candidates[1].Name != "小孩" {
		t.Errorf("unexpected candidate[1]: %+v", loaded.Candidates[1])
	}
	if loaded.Resolved["group"] != "g1" {
		t.Errorf("expected resolved group='g1', got %q", loaded.Resolved["group"])
	}

	// Cleanup
	Delete("test-save-load")
}

func TestLoad_NotFound(t *testing.T) {
	s, err := Load("nonexistent-action-id")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if s != nil {
		t.Error("expected nil for nonexistent state file")
	}
}

func TestDelete_RemovesFile(t *testing.T) {
	s := &ActionState{
		ActionID: "test-delete",
		Step:     "test",
	}
	if err := Save(s); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// Verify file exists
	path := statePath("test-delete")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("expected state file to exist after Save")
	}

	Delete("test-delete")

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("expected state file to be removed after Delete")
	}
}

func TestStatePath_Format(t *testing.T) {
	path := statePath("my-action-id")
	dir := os.TempDir()
	expectedDir := filepath.Clean(dir)
	gotDir := filepath.Dir(path)

	if !strings.HasPrefix(gotDir, expectedDir) {
		t.Errorf("expected path to be in temp dir %q, got dir %q", expectedDir, gotDir)
	}
	if !strings.HasSuffix(path, "na-action-my-action-id.json") {
		t.Errorf("expected filename 'na-action-my-action-id.json', got %q", path)
	}
}

func TestSave_OverwritesExisting(t *testing.T) {
	s1 := &ActionState{
		ActionID: "test-overwrite",
		Step:     "step1",
	}
	if err := Save(s1); err != nil {
		t.Fatalf("first Save: %v", err)
	}

	s2 := &ActionState{
		ActionID: "test-overwrite",
		Step:     "step2",
		Args:     map[string]string{"key": "new-value"},
	}
	if err := Save(s2); err != nil {
		t.Fatalf("second Save: %v", err)
	}

	loaded, err := Load("test-overwrite")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.Step != "step2" {
		t.Errorf("expected Step 'step2' after overwrite, got %q", loaded.Step)
	}
	if loaded.Args["key"] != "new-value" {
		t.Errorf("expected Args['key']='new-value', got %q", loaded.Args["key"])
	}

	Delete("test-overwrite")
}
