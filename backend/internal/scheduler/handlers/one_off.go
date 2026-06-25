package handlers

import (
	"time"

	"github.com/dezhishen/now-and-again/backend/internal/repository"
	"github.com/dezhishen/now-and-again/backend/internal/scheduler"
)

// ─── OneOff: one-time task, auto-archive after completion ─────────

type OneOffHandler struct{}

func (OneOffHandler) Code() string            { return "one_off" }
func (OneOffHandler) Name() string            { return "一次性" }
func (OneOffHandler) Icon() string            { return "📌" }
func (OneOffHandler) Category() string        { return "now" }
func (OneOffHandler) DefaultPriority() string { return "medium" }

func (OneOffHandler) OnCreate(task *repository.TaskModel) error {
	return nil // no special setup
}

func (OneOffHandler) OnComplete(task *repository.TaskModel) *repository.TaskModel {
	return nil // one-off: no auto-reset, just archive
}

func (OneOffHandler) Reminders() []time.Duration {
	return []time.Duration{1 * time.Hour, 24 * time.Hour}
}

func init() { scheduler.Register(OneOffHandler{}) }
