package inspection

import (
	"github.com/dezhishen/now-and-again/backend/pkg/model"
)

// ─── Inspection-specific GORM models ────────────────────────────

// InspectionResultModel stores each inspection submission for audit trail.
type InspectionResultModel struct {
	model.BaseModel
	TaskID     string `gorm:"index;type:char(36);not null"`
	TodoID     string `gorm:"index;type:char(36);not null"`
	FamilyID   string `gorm:"index;type:char(36);not null"`
	ItemName   string `gorm:"size:128;not null"`
	BranchName string `gorm:"size:128;not null"`
	Remark     string `gorm:"size:512"`
	CreatedBy  string `gorm:"index;type:char(36);not null"`
}

func (InspectionResultModel) TableName() string { return "inspection_results" }

// CheckItemModel is a single check item under an inspection task.
type CheckItemModel struct {
	model.BaseModel
	TaskID    string                 `gorm:"index;type:char(36);not null"`
	Name      string                 `gorm:"size:128;not null"`
	SortOrder int                    `gorm:"not null;default:0"`
	Task      model.TaskModel        `gorm:"foreignKey:TaskID"`
	Branches  []CheckItemBranchModel `gorm:"foreignKey:CheckItemID"`
}

func (CheckItemModel) TableName() string { return "check_items" }

// CheckItemBranchModel is a branch/option under a check item.
type CheckItemBranchModel struct {
	model.BaseModel
	CheckItemID  string           `gorm:"index;type:char(36);not null"`
	Name         string           `gorm:"size:128;not null"`
	CreateTodo   bool             `gorm:"not null;default:false"`
	BranchTaskID string           `gorm:"index;type:char(36)"`
	SortOrder    int              `gorm:"not null;default:0"`
	CheckItem    CheckItemModel   `gorm:"foreignKey:CheckItemID"`
	BranchTask   *model.TaskModel `gorm:"foreignKey:BranchTaskID"`
}

func (CheckItemBranchModel) TableName() string { return "check_item_branches" }

func init() {
	model.RegisterModel(&InspectionResultModel{})
	model.RegisterModel(&CheckItemModel{})
	model.RegisterModel(&CheckItemBranchModel{})
}
