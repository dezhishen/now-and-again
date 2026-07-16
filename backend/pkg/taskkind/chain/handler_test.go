package chain

import (
	"testing"

	"github.com/dezhishen/now-and-again/backend/pkg/model"
	"github.com/dezhishen/now-and-again/backend/pkg/taskkind"
	"gorm.io/gorm"
)

type noopTaskStorage struct{}

func (noopTaskStorage) FindTaskByID(taskID string) (*model.TaskModel, error)         { return nil, nil }
func (noopTaskStorage) FindTaskByParentId(parentID string) (*model.TaskModel, error) { return nil, nil }
func (noopTaskStorage) CreateNoRootTask(task *model.TaskModel, extra any) error      { return nil }
func (noopTaskStorage) UpdateNoRootTask(task *model.TaskModel, extra any) error      { return nil }
func (noopTaskStorage) UpdateTaskFields(task *model.TaskModel) error                 { return nil }
func (noopTaskStorage) DeleteNonRootTask(taskID string) error                        { return nil }
func (noopTaskStorage) CreateTodo(taskID string, displaySummary string) (*model.TodoModel, error) {
	return nil, nil
}
func (noopTaskStorage) LookupHandler(kind string) taskkind.Handler { return nil }
func (noopTaskStorage) DB() *gorm.DB                               { return nil }

func TestSaveExtra_EmptySteps(t *testing.T) {
	h := &handler{}
	task := &model.TaskModel{BaseModel: model.BaseModel{ID: "test-id"}}

	// nil extra should fail
	err := h.SaveExtra(noopTaskStorage{}, task, nil)
	if err == nil {
		t.Error("SaveExtra with nil extra: expected error, got nil")
	}

	// empty steps should fail
	err = h.SaveExtra(noopTaskStorage{}, task, map[string]any{"steps": []any{}})
	if err == nil {
		t.Error("SaveExtra with empty steps: expected error, got nil")
	}
}

func TestParseSteps(t *testing.T) {
	// nil extra → empty slice, no error
	steps, err := parseSteps(nil)
	if err != nil {
		t.Errorf("parseSteps(nil): unexpected error: %v", err)
	}
	if len(steps) != 0 {
		t.Errorf("parseSteps(nil): expected empty, got %d steps", len(steps))
	}

	// empty steps → empty slice, no error
	steps, err = parseSteps(map[string]any{"steps": []any{}})
	if err != nil {
		t.Errorf("parseSteps(empty): unexpected error: %v", err)
	}
	if len(steps) != 0 {
		t.Errorf("parseSteps(empty): expected empty, got %d steps", len(steps))
	}

	// valid steps → parsed steps
	steps, err = parseSteps(map[string]any{
		"steps": []any{
			map[string]any{"name": "s1", "kind": "simple"},
		},
	})
	if err != nil {
		t.Errorf("parseSteps(valid): unexpected error: %v", err)
	}
	if len(steps) != 1 || steps[0].Name != "s1" {
		t.Errorf("parseSteps(valid): expected 1 step with name 's1', got %+v", steps)
	}
}
