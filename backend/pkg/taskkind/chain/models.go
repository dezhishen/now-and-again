package chain

import (
	"database/sql"

	"github.com/dezhishen/now-and-again/backend/pkg/model"
)

// ChainStepModel stores one step in a chain task's definition.
// TaskID is the ROOT task that owns this chain.
// ChildTaskID is the actual TaskModel record created for this step.
type ChainStepModel struct {
	model.BaseModel
	TaskID      string         `gorm:"index;type:char(36);not null"                              json:"task_id"`
	SortOrder   int            `gorm:"not null"                                                   json:"sort_order"`
	Name        string         `gorm:"size:128;not null"                                          json:"name"`
	Kind        string         `gorm:"size:16;not null;default:simple"                            json:"kind"`
	GroupID     sql.NullString `gorm:"type:char(36)"                                              json:"-"`
	LocationID  sql.NullString `gorm:"type:char(36)"                                              json:"-"`
	ChildTaskID string         `gorm:"index;type:char(36)"                                        json:"child_task_id"`
}

func (ChainStepModel) TableName() string { return "chain_steps" }

func init() {
	model.RegisterModel(&ChainStepModel{})
}
