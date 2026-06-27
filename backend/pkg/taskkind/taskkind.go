package taskkind

import (
	"github.com/dezhishen/now-and-again/backend/internal/repository"
	"github.com/dezhishen/now-and-again/backend/pkg/scheduler"
)

// ─── Operations ──────────────────────────────────────────────────

// Ops bundles the dependencies handlers need from the service layer.
// Handlers use this to read/write tasks, todos, logs, and the scheduler.
type Ops struct {
	Repo      *repository.TaskRepo
	Scheduler *scheduler.Scheduler
}

// ─── Handler Interface ───────────────────────────────────────────

// Handler defines task-kind-specific behavior. All methods are mandatory.
type Handler interface {
	Kind() string

	// Lifecycle — called by taskService for every task
	OnCreate(ops *Ops, task *repository.TaskModel, extra any) error
	OnUpdate(ops *Ops, task *repository.TaskModel, extra any) error
	OnDelete(ops *Ops, task *repository.TaskModel) error

	// OnComplete is called when a todo of this kind is completed.
	// extra carries the kind-specific payload from CompleteTodoRequest.
	OnComplete(ops *Ops, todo *repository.TodoModel, extra any, branchName, userID string) error

	// GetExtra returns kind-specific data for the task detail page.
	// e.g. for inspection: check_items + children
	GetExtra(ops *Ops, task *repository.TaskModel) (any, error)
}

// Selection represents one checked item in an inspection.
type Selection struct {
	Item   string
	Branch string
}

// ─── Registry ────────────────────────────────────────────────────

var registry = map[string]Handler{}

func Register(h Handler) { registry[h.Kind()] = h }

func Get(kind string) Handler { return registry[kind] }

// All returns all registered handlers.
func All() []Handler {
	result := make([]Handler, 0, len(registry))
	for _, h := range registry {
		result = append(result, h)
	}
	return result
}
