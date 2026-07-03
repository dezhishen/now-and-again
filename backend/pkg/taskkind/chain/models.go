package chain

import (
	"database/sql"

	"github.com/dezhishen/now-and-again/backend/pkg/model"
)

// ChainStepModel stores one step in a chain task's definition.
// TaskID is the ROOT task that owns this chain.
// ChildTaskID is the actual TaskModel record created for this step (NilUUID until created).
type ChainStepModel struct {
	model.BaseModel
	TaskID      string         `gorm:"index;type:char(36);not null"` // root task
	SortOrder   int            `gorm:"not null"`
	Name        string         `gorm:"size:128;not null"`
	Kind        string         `gorm:"size:16;not null;default:simple"`
	GroupID     sql.NullString `gorm:"type:char(36)"`
	LocationID  sql.NullString `gorm:"type:char(36)"`
	ChildTaskID string         `gorm:"index;type:char(36)"` // the actual task record, set after creation
}

func (ChainStepModel) TableName() string { return "chain_steps" }

func init() {
	model.RegisterModel(&ChainStepModel{})
}
