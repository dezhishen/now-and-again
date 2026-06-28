package types

import (
	"time"

	"github.com/dezhishen/now-and-again/backend/pkg/types/task"
)

// ─── Todo ────────────────────────────────────────────────────────

type Todo struct {
	ID          string     `json:"id"`
	TaskID      string     `json:"task_id"`
	FamilyID    string     `json:"family_id"`
	LocationID  string     `json:"location_id,omitempty"`
	AssignedTo  string     `json:"assigned_to,omitempty"`
	Status      string     `json:"status"`           // pending/done/skipped
	Remark      string     `json:"remark,omitempty"` // user note on completion
	DueStart    time.Time  `json:"due_start"`
	DueDate     time.Time  `json:"due_date"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	CompletedBy string     `json:"completed_by,omitempty"`
	Task        *task.Task `json:"task,omitempty"`
	User        *User      `json:"user,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type TodoWithExtra struct {
	Todo  *Todo `json:"todo"`
	Extra any   `json:"extra,omitempty"`
}

type CompleteTodoRequest struct {
	Todo  *Todo `json:"todo,omitempty"`
	Extra any   `json:"extra,omitempty"` // kind-specific payload
}
