package handlers

import (
	"time"

	"github.com/dezhishen/now-and-again/backend/internal/repository"
	"github.com/dezhishen/now-and-again/backend/internal/scheduler"
)

// ─── RecurringDaily: auto-reset every day ─────────────────────────

type RecurringDailyHandler struct{}

func (RecurringDailyHandler) Code() string            { return "recurring_daily" }
func (RecurringDailyHandler) Name() string            { return "每日循环" }
func (RecurringDailyHandler) Icon() string            { return "🔄" }
func (RecurringDailyHandler) Category() string        { return "again" }
func (RecurringDailyHandler) DefaultPriority() string { return "medium" }

func (RecurringDailyHandler) OnCreate(task *repository.TaskModel) error {
	return nil
}

func (h RecurringDailyHandler) OnComplete(task *repository.TaskModel) *repository.TaskModel {
	// Clone the task with fresh ID and reset due date
	nextDue := time.Now().Add(24 * time.Hour)
	return &repository.TaskModel{
		FamilyID:    task.FamilyID,
		TaskTypeID:  task.TaskTypeID,
		Title:       task.Title,
		Description: task.Description,
		Status:      "todo",
		Priority:    task.Priority,
		DueDate:     &nextDue,
		CreatedBy:   task.CreatedBy,
	}
}

func (RecurringDailyHandler) Reminders() []time.Duration {
	return []time.Duration{1 * time.Hour}
}

func init() { scheduler.Register(RecurringDailyHandler{}) }
