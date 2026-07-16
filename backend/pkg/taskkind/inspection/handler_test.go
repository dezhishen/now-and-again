package inspection

import (
	"testing"

	"github.com/dezhishen/now-and-again/backend/pkg/model"
	"github.com/dezhishen/now-and-again/backend/pkg/taskkind"
	"gorm.io/gorm"
)

type noopTaskStorage struct{}

func (noopTaskStorage) FindTaskByID(taskID string) (*model.TaskModel, error)          { return nil, nil }
func (noopTaskStorage) FindTaskByParentId(parentID string) (*model.TaskModel, error)  { return nil, nil }
func (noopTaskStorage) CreateNoRootTask(task *model.TaskModel, extra any) error        { return nil }
func (noopTaskStorage) UpdateNoRootTask(task *model.TaskModel, extra any) error        { return nil }
func (noopTaskStorage) UpdateTaskFields(task *model.TaskModel) error                   { return nil }
func (noopTaskStorage) DeleteNonRootTask(taskID string) error                          { return nil }
func (noopTaskStorage) CreateTodo(taskID string, displaySummary string) (*model.TodoModel, error) {
	return nil, nil
}
func (noopTaskStorage) LookupHandler(kind string) taskkind.Handler { return nil }
func (noopTaskStorage) DB() *gorm.DB                               { return nil }

func TestSaveExtra_EmptyCheckItems(t *testing.T) {
	h := &handler{}
	task := &model.TaskModel{BaseModel: model.BaseModel{ID: "test-id"}}

	// nil extra should fail
	err := h.SaveExtra(noopTaskStorage{}, task, nil)
	if err == nil {
		t.Error("SaveExtra with nil extra: expected error, got nil")
	}

	// empty check_items should fail
	err = h.SaveExtra(noopTaskStorage{}, task, map[string]any{"check_items": []any{}})
	if err == nil {
		t.Error("SaveExtra with empty check_items: expected error, got nil")
	}
}

func TestParseCheckItems(t *testing.T) {
	// nil extra → empty slice, no error
	items, err := parseCheckItems(nil)
	if err != nil {
		t.Errorf("parseCheckItems(nil): unexpected error: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("parseCheckItems(nil): expected empty, got %d items", len(items))
	}

	// empty check_items → empty slice, no error
	items, err = parseCheckItems(map[string]any{"check_items": []any{}})
	if err != nil {
		t.Errorf("parseCheckItems(empty): unexpected error: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("parseCheckItems(empty): expected empty, got %d items", len(items))
	}

	// valid check_items → parsed items
	items, err = parseCheckItems(map[string]any{
		"check_items": []any{
			map[string]any{"name": "a", "branches": []any{}},
		},
	})
	if err != nil {
		t.Errorf("parseCheckItems(valid): unexpected error: %v", err)
	}
	if len(items) != 1 || items[0].Name != "a" {
		t.Errorf("parseCheckItems(valid): expected 1 item with name 'a', got %+v", items)
	}
}
