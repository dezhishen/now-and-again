package types

import (
	"encoding/json"

	"github.com/google/uuid"
)

// ─── TaskLog ──────────────────────────────────────────────────────

// TaskLogAction enumerates all auditable actions.
type TaskLogAction string

const (
	ActionCreated    TaskLogAction = "created"
	ActionStarted    TaskLogAction = "started"
	ActionCompleted  TaskLogAction = "completed"
	ActionReopened   TaskLogAction = "reopened"
	ActionReassigned TaskLogAction = "reassigned"
	ActionReset      TaskLogAction = "reset"
	ActionArchived   TaskLogAction = "archived"
	ActionCommented  TaskLogAction = "comment"
)

type TaskLog struct {
	ID        uuid.UUID       `json:"id"`
	TaskID    uuid.UUID       `json:"task_id"`
	UserID    uuid.UUID       `json:"user_id"`
	Action    TaskLogAction   `json:"action"`
	Detail    json.RawMessage `json:"detail"`
	CreatedAt string          `json:"created_at"`
	// Expanded
	User *User `json:"user,omitempty"`
}

// LogDetail carries the structured payload of a log entry.
type LogDetail struct {
	PreviousStatus   *TaskStatus `json:"previous_status,omitempty"`
	NewStatus        *TaskStatus `json:"new_status,omitempty"`
	PreviousAssignee *uuid.UUID  `json:"previous_assignee,omitempty"`
	NewAssignee      *uuid.UUID  `json:"new_assignee,omitempty"`
	Reason           string      `json:"reason,omitempty"`
	Comment          string      `json:"comment,omitempty"`
	ResetReason      string      `json:"reset_reason,omitempty"`
	NextDueDate      string      `json:"next_due_date,omitempty"`
}

type AddCommentRequest struct {
	Comment string `json:"comment" binding:"required,min=1"`
}
