package types

import (
	"time"

	"github.com/google/uuid"
)

// ─── Inspection ───────────────────────────────────────────────────

type Inspection struct {
	ID          uuid.UUID  `json:"id"`
	FamilyID    uuid.UUID  `json:"family_id"`
	Title       string     `json:"title"`
	Description string     `json:"description,omitempty"`
	Status      TaskStatus `json:"status"` // in_progress | done
	CreatedBy   uuid.UUID  `json:"created_by"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Timestamps
	// Expanded
	Items []InspectionItem `json:"items,omitempty"`
	User  *User            `json:"user,omitempty"`
}

type CreateInspectionRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description,omitempty"`
}

// ─── InspectionItem ───────────────────────────────────────────────

type InspectionItem struct {
	ID              uuid.UUID        `json:"id"`
	InspectionID    uuid.UUID        `json:"inspection_id"`
	CheckPoint      string           `json:"check_point"`
	Result          InspectionResult `json:"result"`
	Note            string           `json:"note,omitempty"`
	GeneratedTaskID *uuid.UUID       `json:"generated_task_id,omitempty"`
	CheckedAt       *time.Time       `json:"checked_at,omitempty"`
	// Expanded
	GeneratedTask *Task `json:"generated_task,omitempty"`
}

type AddInspectionItemRequest struct {
	CheckPoint string `json:"check_point" binding:"required"`
}

type UpdateInspectionItemRequest struct {
	Result          *InspectionResult `json:"result,omitempty"`
	Note            *string           `json:"note,omitempty"`
	CreateTask      *bool             `json:"create_task,omitempty"`      // if true, auto-generate a task
	TaskTitle       *string           `json:"task_title,omitempty"`       // title for generated task
	TaskDescription *string           `json:"task_description,omitempty"` // description for generated task
	TaskPriority    *Priority         `json:"task_priority,omitempty"`
	AssigneeIDs     []uuid.UUID       `json:"assignee_ids,omitempty"`
}
