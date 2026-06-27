package types

import "time"

// ─── Task Template ───────────────────────────────────────────────

type TaskTemplate struct {
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
	DisplaySummary string     `json:"display_summary,omitempty"` // plugin-populated for list view
	LastTodoAt     *time.Time `json:"last_todo_at,omitempty"`
	CreatedBy      string     `json:"created_by"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type CheckItemDTO struct {
	ID        string               `json:"id,omitempty"`
	Name      string               `json:"name"`
	SortOrder int                  `json:"sort_order,omitempty"`
	Branches  []CheckItemBranchDTO `json:"branches"`
}

type CheckItemBranchDTO struct {
	ID           string `json:"id,omitempty"`
	Name         string `json:"name"`
	CreateTodo   bool   `json:"create_todo"`
	TodoName     string `json:"todo_name,omitempty"`
	BranchTaskID string `json:"branch_task_id,omitempty"`
	SortOrder    int    `json:"sort_order,omitempty"`
}

type CreateTaskRequest struct {
	Name         string         `json:"name" binding:"required"`
	ScheduleType string         `json:"schedule_type" binding:"required"`
	ScheduleData any            `json:"schedule_data" binding:"required"`
	GroupID      string         `json:"group_id,omitempty"`
	LocationID   string         `json:"location_id,omitempty"`
	Kind         string         `json:"kind,omitempty"`        // simple | inspection
	CheckItems   []CheckItemDTO `json:"check_items,omitempty"` // only for kind=inspection
}

type UpdateTaskRequest struct {
	Name         *string        `json:"name,omitempty"`
	ScheduleType *string        `json:"schedule_type,omitempty"`
	ScheduleData any            `json:"schedule_data,omitempty"`
	Enabled      *bool          `json:"enabled,omitempty"`
	GroupID      *string        `json:"group_id,omitempty"`
	LocationID   *string        `json:"location_id,omitempty"`
	Kind         *string        `json:"kind,omitempty"`
	CheckItems   []CheckItemDTO `json:"check_items,omitempty"`
}

// ─── Todo ────────────────────────────────────────────────────────

type Todo struct {
	ID          string        `json:"id"`
	TaskID      string        `json:"task_id"`
	FamilyID    string        `json:"family_id"`
	LocationID  string        `json:"location_id,omitempty"`
	AssignedTo  string        `json:"assigned_to,omitempty"`
	Status      string        `json:"status"`                // pending/done/skipped
	BranchName  string        `json:"branch_name,omitempty"` // selected branch
	Remark      string        `json:"remark,omitempty"`      // user note on completion
	DueStart    time.Time     `json:"due_start"`
	DueDate     time.Time     `json:"due_date"`
	CompletedAt *time.Time    `json:"completed_at,omitempty"`
	CompletedBy string        `json:"completed_by,omitempty"`
	Task        *TaskTemplate `json:"task,omitempty"`
	User        *User         `json:"user,omitempty"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

type CompleteTodoRequest struct {
	Status     string                `json:"status" binding:"required"` // "done" or "skipped"
	BranchName string                `json:"branch_name,omitempty"`
	Remark     string                `json:"remark,omitempty"`     // user note
	Selections []InspectionSelection `json:"selections,omitempty"` // for inspection kind
}

type InspectionSelection struct {
	Item   string `json:"item"`   // check item name
	Branch string `json:"branch"` // selected branch name
}

// ─── Task Log ────────────────────────────────────────────────────

type TaskLog struct {
	ID         string    `json:"id"`
	TaskID     string    `json:"task_id"`
	TodoID     string    `json:"todo_id,omitempty"`
	Status     string    `json:"status"`
	Message    string    `json:"message,omitempty"`
	LogType    string    `json:"log_type"`
	OperatorID string    `json:"operator_id,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}
