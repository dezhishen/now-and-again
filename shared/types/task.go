package types

import "time"

// ─── Task Template ───────────────────────────────────────────────

type TaskTemplate struct {
	ID               string     `json:"id"`
	FamilyID         string     `json:"family_id"`
	GroupID          string     `json:"group_id,omitempty"`
	LocationID       string     `json:"location_id,omitempty"`
	Name             string     `json:"name"`
	ScheduleType     string     `json:"schedule_type"`
	ScheduleData     any        `json:"schedule_data"` // JSON: varies by type
	Enabled          bool       `json:"enabled"`
	IsInspection     bool       `json:"is_inspection"`
	InspectionConfig any        `json:"inspection_config,omitempty"` // null for regular tasks
	LastTodoAt       *time.Time `json:"last_todo_at,omitempty"`
	CreatedBy        string     `json:"created_by"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type CreateTaskRequest struct {
	Name             string `json:"name" binding:"required"`
	ScheduleType     string `json:"schedule_type" binding:"required"`
	ScheduleData     any    `json:"schedule_data" binding:"required"`
	GroupID          string `json:"group_id,omitempty"`
	LocationID       string `json:"location_id,omitempty"`
	IsInspection     bool   `json:"is_inspection,omitempty"`
	InspectionConfig any    `json:"inspection_config,omitempty"`
}

type UpdateTaskRequest struct {
	Name             *string `json:"name,omitempty"`
	ScheduleType     *string `json:"schedule_type,omitempty"`
	ScheduleData     any     `json:"schedule_data,omitempty"`
	Enabled          *bool   `json:"enabled,omitempty"`
	GroupID          *string `json:"group_id,omitempty"`
	LocationID       *string `json:"location_id,omitempty"`
	IsInspection     *bool   `json:"is_inspection,omitempty"`
	InspectionConfig any     `json:"inspection_config,omitempty"`
}

// ─── Todo ────────────────────────────────────────────────────────

type Todo struct {
	ID               string        `json:"id"`
	TaskID           string        `json:"task_id"`
	FamilyID         string        `json:"family_id"`
	LocationID       string        `json:"location_id,omitempty"`
	AssignedTo       string        `json:"assigned_to,omitempty"`
	Status           string        `json:"status"`                      // pending/done/skipped
	TodoType         string        `json:"todo_type"`                   // task/inspection
	InspectionResult string        `json:"inspection_result,omitempty"` // normal/abnormal
	DueStart         time.Time     `json:"due_start"`
	DueDate          time.Time     `json:"due_date"`
	CompletedAt      *time.Time    `json:"completed_at,omitempty"`
	CompletedBy      string        `json:"completed_by,omitempty"`
	Task             *TaskTemplate `json:"task,omitempty"`
	User             *User         `json:"user,omitempty"`
	CreatedAt        time.Time     `json:"created_at"`
	UpdatedAt        time.Time     `json:"updated_at"`
}

type CompleteTodoRequest struct {
	Status           string `json:"status" binding:"required"`   // "done" or "skipped"
	InspectionResult string `json:"inspection_result,omitempty"` // "normal" or "abnormal" (only for inspection type)
}

// ─── Task Log ────────────────────────────────────────────────────

type TaskLog struct {
	ID        string    `json:"id"`
	TaskID    string    `json:"task_id"`
	Status    string    `json:"status"` // registered/triggered/success/error
	Message   string    `json:"message,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}
