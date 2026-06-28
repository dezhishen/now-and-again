// This file contains the repository which is used by the inspection task kind to manage check items and their branches.
package repository

import "gorm.io/gorm"

type CheckItemRepo struct{ db *gorm.DB }

func NewCheckItemRepo(db *gorm.DB) *CheckItemRepo { return &CheckItemRepo{db} }

// ======== CheckItemRepo ========

func (repo *CheckItemRepo) FindCheckItemsByTask(taskID string) ([]CheckItemModel, error) {
	var checkItems []CheckItemModel
	err := repo.db.Preload("Branches", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort_order ASC")
	}).Where("task_id = ?", taskID).Order("sort_order ASC").Find(&checkItems).Error
	if err != nil {
		return nil, err
	}
	return checkItems, nil
}

func (repo *CheckItemRepo) CreateCheckItem(ci *CheckItemModel) error {
	return repo.db.Create(ci).Error
}

func (repo *CheckItemRepo) DeleteCheckItemsByTask(taskID string) error {
	return repo.db.Where("task_id = ?", taskID).Delete(&CheckItemModel{}).Error
}

func (repo *CheckItemRepo) CreateInspectionResult(r *InspectionResultModel) error {
	return repo.db.Create(r).Error
}

// ======== CheckItemBranchRepo ========
type CheckItemBranchRepo struct{ db *gorm.DB }

func NewCheckItemBranchRepo(db *gorm.DB) *CheckItemBranchRepo { return &CheckItemBranchRepo{db} }

func (repo *CheckItemBranchRepo) FindCheckItemBranchesByCheckItem(checkItemID string) ([]CheckItemBranchModel, error) {
	var branches []CheckItemBranchModel
	err := repo.db.Where("check_item_id = ?", checkItemID).Order("sort_order ASC").Find(&branches).Error
	return branches, err
}

func (checkItemBranchRepo *CheckItemBranchRepo) DeleteCheckItemBranchesByTask(taskID string) error {
	return checkItemBranchRepo.db.
		Where("check_item_id IN (?)", checkItemBranchRepo.db.Model(&CheckItemModel{}).Select("id").Where("task_id = ?", taskID)).
		Delete(&CheckItemBranchModel{}).Error
}

// CreateCheckItemBranch creates a new check item branch.
func (checkItemBranchRepo *CheckItemBranchRepo) CreateCheckItemBranch(cb *CheckItemBranchModel) error {
	return checkItemBranchRepo.db.Create(cb).Error
}

// BatchCreateCheckItemBranches creates multiple check item branches in a single transaction.
func (checkItemBranchRepo *CheckItemBranchRepo) BatchCreateCheckItemBranches(branches []CheckItemBranchModel) error {
	return checkItemBranchRepo.db.Create(&branches).Error
}
