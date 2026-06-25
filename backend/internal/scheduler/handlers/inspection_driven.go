package handlers

import (
	"time"

	"github.com/dezhishen/now-and-again/backend/internal/repository"
	"github.com/dezhishen/now-and-again/backend/internal/scheduler"
)

// ─── InspectionDriven: created from inspection, high priority ─────

type InspectionDrivenHandler struct{}

func (InspectionDrivenHandler) Code() string            { return "inspection_driven" }
func (InspectionDrivenHandler) Name() string            { return "巡检驱动" }
func (InspectionDrivenHandler) Icon() string            { return "🔍" }
func (InspectionDrivenHandler) Category() string        { return "now" }
func (InspectionDrivenHandler) DefaultPriority() string { return "high" }

func (InspectionDrivenHandler) OnCreate(task *repository.TaskModel) error {
	return nil
}

func (InspectionDrivenHandler) OnComplete(task *repository.TaskModel) *repository.TaskModel {
	return nil // inspection tasks: no auto-reset
}

func (InspectionDrivenHandler) Reminders() []time.Duration {
	return []time.Duration{2 * time.Hour}
}

func init() { scheduler.Register(InspectionDrivenHandler{}) }
