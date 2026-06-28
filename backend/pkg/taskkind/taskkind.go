package taskkind

import (
	"github.com/dezhishen/now-and-again/backend/pkg/model"
	"gorm.io/gorm"
)

// ─── Operations ──────────────────────────────────────────────────
// Read and display-summary methods may operate on any task.
// Write methods (Create/Update/Delete) are suffixed NoRoot because they
// must only be used for plugin-owned sub-tasks, never the root task.
type TaskStorage interface {
	FindTaskByID(taskID string) (*model.TaskModel, error)
	FindTaskByParentId(parentID string) (*model.TaskModel, error)
	CreateNoRootTask(task *model.TaskModel) error
	UpdateNoRootTask(task *model.TaskModel) error
	DeleteNoRootTask(taskID string) error
	DB() *gorm.DB
}

// ─── Handler Interface ───────────────────────────────────────────

// Handler defines task-kind-specific behavior. All methods are mandatory.
// Use taskStorage.DB() to obtain the *gorm.DB for kind-specific DB operations.
type Handler interface {
	Kind() string

	// Lifecycle — called by taskService for every task.
	OnCreate(taskStorage TaskStorage, task *model.TaskModel, extra any) error
	OnUpdate(taskStorage TaskStorage, task *model.TaskModel, extra any) error
	OnDelete(taskStorage TaskStorage, task *model.TaskModel) error
	// OnComplete is called when a todo of this kind is completed.
	// extra carries the kind-specific payload from CompleteTodoRequest.
	OnComplete(taskStorage TaskStorage, todo *model.TodoModel, extra any) error
	// GetExtra returns kind-specific data for the task detail page.
	// e.g. for inspection: check_items + children
	GetExtra(taskStorage TaskStorage, task *model.TaskModel) (any, error)
}

// Selection represents one checked item in an inspection.
type Selection struct {
	ItemID     string `json:"item_id"`
	BranchID   string `json:"branch_id"`
	ItemName   string `json:"item_name"`
	BranchName string `json:"branch_name"`
}

type TaskManager struct {
	registry map[string]Handler
}

func NewTaskManager() *TaskManager {
	return &TaskManager{registry: make(map[string]Handler)}
}

func (tm *TaskManager) Register(h Handler) { tm.registry[h.Kind()] = h }

func (tm *TaskManager) Get(kind string) Handler { return tm.registry[kind] }

func (tm *TaskManager) All() []Handler {
	result := make([]Handler, 0, len(tm.registry))
	for _, h := range tm.registry {
		result = append(result, h)
	}
	return result
}

var defaultTaskManager = NewTaskManager()

// ─── Registry ────────────────────────────────────────────────────

func Register(h Handler) {
	defaultTaskManager.Register(h)
}

func GetManager() *TaskManager {
	return defaultTaskManager
}
