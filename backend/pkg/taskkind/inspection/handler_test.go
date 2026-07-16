package inspection

import (
	"testing"

	"github.com/dezhishen/now-and-again/backend/pkg/model"
	"github.com/dezhishen/now-and-again/backend/pkg/taskkind"
	"github.com/dezhishen/now-and-again/backend/pkg/types"
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

func TestSaveExtra_NoAnomalyBranch(t *testing.T) {
	h := &handler{}
	task := &model.TaskModel{BaseModel: model.BaseModel{ID: "test-id"}, Name: "巡检测试"}

	// items without any anomaly branch (all create_todo = false) should fail
	err := h.SaveExtra(noopTaskStorage{}, task, map[string]any{
		"check_items": []any{
			map[string]any{
				"name": "区域A",
				"branches": []any{
					map[string]any{"name": "正常", "create_todo": false},
				},
			},
		},
	})
	if err == nil {
		t.Error("SaveExtra without anomaly branch: expected error, got nil")
	}
}

func TestHasAnomalyBranch(t *testing.T) {
	tests := []struct {
		name     string
		items    []types.CheckItemDTO
		expected bool
	}{
		{
			name:     "empty items",
			items:    []types.CheckItemDTO{},
			expected: false,
		},
		{
			name: "no anomaly branch",
			items: []types.CheckItemDTO{
				{
					Name: "a",
					Branches: []types.CheckItemBranchDTO{
						{Name: "正常", CreateTodo: false},
					},
				},
			},
			expected: false,
		},
		{
			name: "has anomaly branch",
			items: []types.CheckItemDTO{
				{
					Name: "a",
					Branches: []types.CheckItemBranchDTO{
						{Name: "正常", CreateTodo: false},
						{Name: "异常", CreateTodo: true},
					},
				},
			},
			expected: true,
		},
		{
			name: "multiple items, second has anomaly",
			items: []types.CheckItemDTO{
				{Name: "a", Branches: []types.CheckItemBranchDTO{{Name: "正常", CreateTodo: false}}},
				{Name: "b", Branches: []types.CheckItemBranchDTO{{Name: "异常", CreateTodo: true}}},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasAnomalyBranch(tt.items)
			if got != tt.expected {
				t.Errorf("hasAnomalyBranch() = %v, want %v", got, tt.expected)
			}
		})
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
