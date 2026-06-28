package task

// ─── Check Items (inspection task kind) ─────────────────────────

type CheckItemDTO struct {
	ID        string               `json:"id,omitempty"`
	Name      string               `json:"name"`
	SortOrder int                  `json:"sort_order,omitempty"`
	Branches  []CheckItemBranchDTO `json:"branches"`
}

type CheckItemBranchDTO struct {
	ID           string         `json:"id,omitempty"`
	Name         string         `json:"name"`
	CreateTodo   bool           `json:"create_todo"`
	BranchTaskID string         `json:"branch_task_id,omitempty"`
	SortOrder    int            `json:"sort_order,omitempty"`
	BranchTask   *TaskWithExtra `json:"branch_task,omitempty"`
}
