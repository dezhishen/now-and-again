package types

import "time"

// ─── Task Log ────────────────────────────────────────────────────

type TaskLog struct {
	ID         string    `json:"id"`
	TaskID     string    `json:"task_id"`
	TaskName   string    `json:"task_name,omitempty"`
	TodoID     string    `json:"todo_id,omitempty"`
	Status     string    `json:"status"`
	Message    string    `json:"message,omitempty"`
	LogType    string    `json:"log_type"`
	OperatorID string    `json:"operator_id,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}
