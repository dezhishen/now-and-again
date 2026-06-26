package taskkind

import (
	"github.com/dezhishen/now-and-again/backend/internal/repository"
	"github.com/dezhishen/now-and-again/backend/pkg/scheduler"
	"github.com/gin-gonic/gin"
)

// ─── Operations ──────────────────────────────────────────────────

// Ops bundles the dependencies handlers need from the service layer.
// Handlers use this to read/write tasks, todos, logs, and the scheduler.
type Ops struct {
	Repo      *repository.TaskRepo
	Scheduler *scheduler.Scheduler
}

// ─── Handler Interface ───────────────────────────────────────────

// Handler defines task-kind-specific behavior.
type Handler interface {
	Kind() string
	OnComplete(ops *Ops, todo *repository.TodoModel, branchName, userID string) error
}

// Inspector is an optional interface for handlers that support multi-item inspection.
type Inspector interface {
	Handler
	OnInspect(ops *Ops, todo *repository.TodoModel, selections []Selection, userID string) error
}

// RouteRegistrar is an optional interface for task kinds that expose additional
// API endpoints beyond the generic task CRUD. Routes are registered under
// a kind-specific path prefix, e.g. /api/tasks/:task_id/{kind}/...
//
// The router parameter is a *gin.RouterGroup scoped to the task kind, so
// handlers can call router.POST("/submit", ...) without repeating the prefix.
type RouteRegistrar interface {
	Handler
	RegisterRoutes(router *gin.RouterGroup, ops *Ops)
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
