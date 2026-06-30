package taskkind

import (
	"github.com/dezhishen/now-and-again/backend/pkg/model"
	"gorm.io/gorm"
)

// ─── Operations ──────────────────────────────────────────────────
// Read and display-summary methods may operate on any task.
// Write methods go through the full lifecycle (triggering kind handlers).
type TaskStorage interface {
	FindTaskByID(taskID string) (*model.TaskModel, error)
	FindTaskByParentId(parentID string) (*model.TaskModel, error)
	// CreateNoRootTask creates a non-root task AND triggers its SaveExtra handler.
	// The caller does NOT need to look up the kind handler or call SaveExtra manually.
	CreateNoRootTask(task *model.TaskModel, extra any) error
	// UpdateNoRootTask updates task fields and always triggers the kind's
	// UpdateExtra handler. nil extra means "clear all extra data".
	// For field-only updates without extra lifecycle, use UpdateTaskFields.
	UpdateNoRootTask(task *model.TaskModel, extra any) error
	// UpdateTaskFields updates task fields WITHOUT triggering any kind handler.
	// Use for simple field changes (e.g., setting RootTaskID, display_summary).
	UpdateTaskFields(task *model.TaskModel) error
	// DeleteNonRootTask deletes a non-root task AND all its descendants,
	// triggering DeleteExtra for each task in the subtree.
	DeleteNonRootTask(taskID string) error
	// CreateTodo creates a pending todo for the given task.
	// displaySummary is an optional kind-specific context string shown on the todo card.
	// Returns the created todo so callers can attach logs.
	CreateTodo(taskID string, displaySummary string) (*model.TodoModel, error)
	DB() *gorm.DB
}

// ─── Handler Interface ───────────────────────────────────────────

// Handler defines task-kind-specific behavior. All methods are mandatory.
// Use taskStorage.DB() to obtain the *gorm.DB for kind-specific DB operations.
type Handler interface {
	Kind() string

	// Extra data lifecycle — called by taskService for plugin-specific data persistence.
	SaveExtra(taskStorage TaskStorage, task *model.TaskModel, extra any) error
	UpdateExtra(taskStorage TaskStorage, task *model.TaskModel, extra any) error
	DeleteExtra(taskStorage TaskStorage, task *model.TaskModel) error
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
