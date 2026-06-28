package inspection

import "gorm.io/gorm"

// ─── CheckItemRepo ─────────────────────────────────────────────

type CheckItemRepo struct{ db *gorm.DB }

func NewCheckItemRepo(db *gorm.DB) *CheckItemRepo { return &CheckItemRepo{db} }

func (r *CheckItemRepo) FindCheckItemsByTask(taskID string) ([]CheckItemModel, error) {
	var items []CheckItemModel
	err := r.db.Preload("Branches", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort_order ASC")
	}).Where("task_id = ?", taskID).Order("sort_order ASC").Find(&items).Error
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (r *CheckItemRepo) CreateCheckItem(ci *CheckItemModel) error {
	return r.db.Create(ci).Error
}

func (r *CheckItemRepo) DeleteCheckItemsByTask(taskID string) error {
	return r.db.Where("task_id = ?", taskID).Delete(&CheckItemModel{}).Error
}

func (r *CheckItemRepo) CreateInspectionResult(result *InspectionResultModel) error {
	return r.db.Create(result).Error
}

// ─── CheckItemBranchRepo ───────────────────────────────────────

type CheckItemBranchRepo struct{ db *gorm.DB }

func NewCheckItemBranchRepo(db *gorm.DB) *CheckItemBranchRepo { return &CheckItemBranchRepo{db} }

func (r *CheckItemBranchRepo) DeleteCheckItemBranchesByTask(taskID string) error {
	return r.db.
		Where("check_item_id IN (?)", r.db.Model(&CheckItemModel{}).Select("id").Where("task_id = ?", taskID)).
		Delete(&CheckItemBranchModel{}).Error
}

func (r *CheckItemBranchRepo) CreateCheckItemBranch(cb *CheckItemBranchModel) error {
	return r.db.Create(cb).Error
}
