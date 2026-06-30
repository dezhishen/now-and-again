package types

import (
	"time"
)

// ─── Todo ────────────────────────────────────────────────────────

type Todo struct {
	ID             string     `json:"id"`
	TaskID         string     `json:"task_id"`
	FamilyID       string     `json:"family_id"`
	LocationID     string     `json:"location_id,omitempty"`
	AssignedTo     string     `json:"assigned_to,omitempty"`
	Status         string     `json:"status"`                    // pending/done/skipped
	Remark         string     `json:"remark,omitempty"`          // user note on completion
	DisplaySummary string     `json:"display_summary,omitempty"` // kind-specific display text
	TaskName       string     `json:"task_name,omitempty"`       // redundant: survives task deletion
	TaskKind       string     `json:"task_kind,omitempty"`       // redundant: survives task deletion
	DueStart       time.Time  `json:"due_start"`
	DueDate        time.Time  `json:"due_date"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
	CompletedBy    string     `json:"completed_by,omitempty"`
	Task           *Task      `json:"task,omitempty"`
	User           *User      `json:"user,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type TodoWithExtra struct {
	Todo  *Todo `json:"todo"`
	Extra any   `json:"extra,omitempty"`
}

type CompleteTodoRequest struct {
	Todo  *Todo `json:"todo,omitempty"`
	Extra any   `json:"extra,omitempty"` // kind-specific payload
}
