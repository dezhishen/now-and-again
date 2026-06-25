package types

import "github.com/google/uuid"

// ─── TaskChain ────────────────────────────────────────────────────

type TaskChain struct {
	ID          uuid.UUID `json:"id"`
	FamilyID    uuid.UUID `json:"family_id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Icon        string    `json:"icon,omitempty"`
	IsActive    bool      `json:"is_active"`
	CreatedBy   uuid.UUID `json:"created_by"`
	Timestamps
	// Expanded
	Steps []TaskChainStep `json:"steps,omitempty"`
}

type CreateChainRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=128"`
	Description string `json:"description,omitempty"`
	Icon        string `json:"icon,omitempty"`
}

// ─── TaskChainStep ────────────────────────────────────────────────

type TaskChainStep struct {
	ID                   uuid.UUID  `json:"id"`
	ChainID              uuid.UUID  `json:"chain_id"`
	SortOrder            int        `json:"sort_order"`
	Title                string     `json:"title"`
	Description          string     `json:"description,omitempty"`
	TaskCode           uuid.UUID  `json:"task_code"`
	AssignedRole         AssignRole `json:"assigned_role"`
	AssignedSubGroupID   *uuid.UUID `json:"assigned_sub_group_id,omitempty"`
	DelayAfterPrevious   string     `json:"delay_after_previous"` // "0h", "1d"
	IsOptional           bool       `json:"is_optional"`
	Priority             Priority   `json:"priority"`
	CreatedAt            string     `json:"created_at"`
	// Expanded
}

type AddStepRequest struct {
	Title              string     `json:"title" binding:"required,min=1,max=255"`
	Description        string     `json:"description,omitempty"`
	TaskCode         uuid.UUID  `json:"task_code" binding:"required"`
	AssignedRole       AssignRole `json:"assigned_role" binding:"required"`
	AssignedSubGroupID *uuid.UUID `json:"assigned_sub_group_id,omitempty"`
	DelayAfterPrevious string     `json:"delay_after_previous,omitempty"`
	IsOptional         bool       `json:"is_optional,omitempty"`
	Priority           Priority   `json:"priority,omitempty"`
}

type ReorderStepsRequest struct {
	StepIDs []uuid.UUID `json:"step_ids" binding:"required"`
}

// StartChainResponse is returned when a chain is instantiated.
type StartChainResponse struct {
	ChainRunID uuid.UUID `json:"chain_run_id"`
	Tasks      []Task    `json:"tasks"`
}
