// Package scheduler provides the task scheduling engine.
//
// Scheduling types are registered via init() by implementing the
// ScheduleHandler interface. Each handler defines both metadata
// (for UI display via GET /api/task-types) and behavior
// (lifecycle hooks: OnCreate / OnComplete / Reminders).
//
// To add a new scheduling type:
//  1. Create a file in backend/internal/scheduler/handlers/
//  2. Implement the ScheduleHandler interface
//  3. Call Register() in init()
package scheduler

import (
	"time"

	"github.com/dezhishen/now-and-again/backend/internal/repository"
)

// ─── ScheduleHandler Interface ────────────────────────────────────

// ScheduleHandler defines the metadata and lifecycle behavior of one
// task scheduling type. A handler controls HOW a task is scheduled
// and what happens when it's created or completed.
type ScheduleHandler interface {
	Code() string
	Name() string
	Icon() string
	Category() string        // "now" | "again"
	DefaultPriority() string // "low" | "medium" | "high" | "urgent"

	// OnCreate is called when a task of this type is created.
	OnCreate(task *repository.TaskModel) error
	// OnComplete is called when a task of this type is marked done.
	// Returns a new TaskModel if the task should auto-reset, nil otherwise.
	OnComplete(task *repository.TaskModel) *repository.TaskModel
	// Reminders returns durations before due date to send reminders.
	Reminders() []time.Duration
}

// ─── Registry ────────────────────────────────────────────────────

var registry = map[string]ScheduleHandler{}

func Register(h ScheduleHandler)      { registry[h.Code()] = h }
func Get(code string) ScheduleHandler { return registry[code] }
func All() []ScheduleHandler {
	list := make([]ScheduleHandler, 0, len(registry))
	for _, h := range registry {
		list = append(list, h)
	}
	return list
}

// ─── HandlerDef (for API) ────────────────────────────────────────

type HandlerDef struct {
	Code            string `json:"code"`
	Name            string `json:"name"`
	Icon            string `json:"icon"`
	Category        string `json:"category"`
	DefaultPriority string `json:"default_priority"`
}

func ListHandlerDefs() []HandlerDef {
	list := make([]HandlerDef, 0, len(registry))
	for _, h := range registry {
		list = append(list, HandlerDef{
			Code: h.Code(), Name: h.Name(), Icon: h.Icon(),
			Category: h.Category(), DefaultPriority: h.DefaultPriority(),
		})
	}
	return list
}
