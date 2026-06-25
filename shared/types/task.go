package types

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ─── ScheduleType ─────────────────────────────────────────────────────

type ScheduleType struct {
	ID              uuid.UUID       `json:"id"`
	Code            string          `json:"code"`
	Name            string          `json:"name"`
	Category        TaskCategory    `json:"category"`
	BehaviorConfig  json.RawMessage `json:"behavior_config"`
	DefaultPriority Priority        `json:"default_priority"`
	Icon            string          `json:"icon,omitempty"`
	IsActive        bool            `json:"is_active"`
	Timestamps
}

// BehaviorConfig is the parsed form of ScheduleType.BehaviorConfig.
type BehaviorConfig struct {
	Recurrence      *RecurrenceConfig `json:"recurrence,omitempty"`
	ResetStrategy   string            `json:"reset_strategy,omitempty"` // "reset_on_complete" | "manual"
	Reminder        *ReminderConfig   `json:"reminder,omitempty"`
	Source          string            `json:"source,omitempty"` // "inspection" | "manual"
	RequiresConfirm bool              `json:"requires_confirm,omitempty"`
	AutoArchiveDays int               `json:"auto_archive_days,omitempty"`
}

type RecurrenceConfig struct {
	Type             string   `json:"type"`             // "interval" | "cron" | "custom"
	DefaultInterval  string   `json:"default_interval"` // "7d", "14d", "1mon"
	AllowedIntervals []string `json:"allowed_intervals"`
}

type ReminderConfig struct {
	BeforeDue []string `json:"before_due,omitempty"` // ["1h", "1d"]
	OnOverdue bool     `json:"on_overdue"`
}

// ─── Task ─────────────────────────────────────────────────────────

type Task struct {
	ID               uuid.UUID       `json:"id"`
	FamilyID         uuid.UUID       `json:"family_id"`
	SubGroupID       *uuid.UUID      `json:"sub_group_id,omitempty"`
	TaskCode string       `json:"task_code"`
	ChainID          *uuid.UUID      `json:"chain_id,omitempty"` // NULL if not part of a chain
	Title            string          `json:"title"`
	Description      string          `json:"description,omitempty"`
	Status           TaskStatus      `json:"status"`
	Priority         Priority        `json:"priority"`
	DueDate          *time.Time      `json:"due_date,omitempty"`
	RecurrenceConfig json.RawMessage `json:"recurrence_config,omitempty"`
	TypeSpecificData json.RawMessage `json:"type_specific_data,omitempty"`
	CreatedBy        uuid.UUID       `json:"created_by"`
	CompletedAt      *time.Time      `json:"completed_at,omitempty"`
	Timestamps
	// Expanded
	TaskType  *ScheduleType        `json:"task_type,omitempty"`
	Assignees []TaskAssignee   `json:"assignees,omitempty"`
	BlockedBy []TaskDependency `json:"blocked_by,omitempty"`
	Blocks    []TaskDependency `json:"blocks,omitempty"`
}

type CreateTaskRequest struct {
	FamilyID         uuid.UUID       `json:"family_id" binding:"required"`
	SubGroupID       *uuid.UUID      `json:"sub_group_id,omitempty"`
	TaskCode string       `json:"task_code" binding:"required"`
	Title            string          `json:"title" binding:"required,min=1,max=255"`
	Description      string          `json:"description,omitempty"`
	Priority         Priority        `json:"priority,omitempty"`
	DueDate          *time.Time      `json:"due_date,omitempty"`
	RecurrenceConfig json.RawMessage `json:"recurrence_config,omitempty"`
	TypeSpecificData json.RawMessage `json:"type_specific_data,omitempty"`
	AssigneeIDs      []uuid.UUID     `json:"assignee_ids,omitempty"`
}

type UpdateTaskRequest struct {
	Title            *string         `json:"title,omitempty"`
	Description      *string         `json:"description,omitempty"`
	Status           *TaskStatus     `json:"status,omitempty"`
	Priority         *Priority       `json:"priority,omitempty"`
	DueDate          *time.Time      `json:"due_date,omitempty"`
	RecurrenceConfig json.RawMessage `json:"recurrence_config,omitempty"`
	TypeSpecificData json.RawMessage `json:"type_specific_data,omitempty"`
}

// ─── TaskAssignee ─────────────────────────────────────────────────

type TaskAssignee struct {
	ID         uuid.UUID `json:"id"`
	TaskID     uuid.UUID `json:"task_id"`
	UserID     uuid.UUID `json:"user_id"`
	AssignedAt string    `json:"assigned_at"`
	// Expanded
	User *User `json:"user,omitempty"`
}

type SetAssigneesRequest struct {
	UserIDs []uuid.UUID `json:"user_ids" binding:"required,min=1"`
}

// ─── TaskDependency ───────────────────────────────────────────────

type TaskDependency struct {
	ID             uuid.UUID      `json:"id"`
	BlockedTaskID  uuid.UUID      `json:"blocked_task_id"`
	BlockerTaskID  uuid.UUID      `json:"blocker_task_id"`
	DependencyType DependencyType `json:"dependency_type"`
	CreatedAt      string         `json:"created_at"`
}

type CreateDependencyRequest struct {
	BlockedTaskID  uuid.UUID      `json:"blocked_task_id" binding:"required"`
	BlockerTaskID  uuid.UUID      `json:"blocker_task_id" binding:"required"`
	DependencyType DependencyType `json:"dependency_type" binding:"required"`
}
