package types

import "time"

// ─── Task ────────────────────────────────────────────────────────

type Task struct {
	ID             string     `json:"id"`
	FamilyID       string     `json:"family_id"`
	GroupID        string     `json:"group_id,omitempty"`
	LocationID     string     `json:"location_id,omitempty"`
	ParentTaskID   string     `json:"parent_task_id,omitempty"`
	IsRoot         bool       `json:"is_root"`
	Name           string     `json:"name"`
	ScheduleType   string     `json:"schedule_type"`
	ScheduleData   any        `json:"schedule_data"`
	Enabled        bool       `json:"enabled"`
	Kind           string     `json:"kind"`
	DisplaySummary string     `json:"display_summary,omitempty"`
	LastTodoAt     *time.Time `json:"last_todo_at,omitempty"`
	CreatedBy      string     `json:"created_by"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type TaskWithExtra struct {
	Task  *Task `json:"task"`
	Extra any   `json:"extra,omitempty"`
}

type CreateTaskRequest struct {
	Task  Task `json:"task" binding:"required"`
	Extra any  `json:"extra,omitempty"`
}

type UpdateTaskRequest struct {
	Task  *Task `json:"task,omitempty"`
	Extra any   `json:"extra,omitempty"`
}
