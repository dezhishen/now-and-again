package repository

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// ─── TaskRepo ─────────────────────────────────────────────────────

func (r *TaskRepo) Create(task *TaskModel) error {
	return r.db.Create(task).Error
}

func (r *TaskRepo) FindByID(id string) (*TaskModel, error) {
	var t TaskModel
	err := r.db.Preload("TaskType").Preload("Assignees.User").Where("id = ?", id).First(&t).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &t, err
}

func (r *TaskRepo) List(familyID string, status *string, assigneeID *string, offset, limit int) ([]TaskModel, int64, error) {
	query := r.db.Model(&TaskModel{}).Where("family_id = ?", familyID)
	if status != nil && *status != "" {
		query = query.Where("status = ?", *status)
	}
	if assigneeID != nil && *assigneeID != "" {
		query = query.Where("id IN (SELECT task_id FROM task_assignees WHERE user_id = ?)", *assigneeID)
	}

	var total int64
	query.Count(&total)

	var tasks []TaskModel
	err := query.Preload("TaskType").Preload("Assignees.User").
		Order("created_at DESC").Offset(offset).Limit(limit).Find(&tasks).Error
	return tasks, total, err
}

func (r *TaskRepo) Update(task *TaskModel) error {
	return r.db.Save(task).Error
}

// ─── TaskAssignee ─────────────────────────────────────────────────

func (r *TaskRepo) SetAssignees(taskID string, userIDs []string) error {
	// Remove existing
	r.db.Where("task_id = ?", taskID).Delete(&TaskAssigneeModel{})
	// Add new
	for _, uid := range userIDs {
		a := &TaskAssigneeModel{
			TaskID:     taskID,
			UserID:     uid,
			AssignedAt: time.Now(),
		}
		if err := r.db.Create(a).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *TaskRepo) ListAssignees(taskID string) ([]TaskAssigneeModel, error) {
	var assignees []TaskAssigneeModel
	err := r.db.Where("task_id = ?", taskID).Preload("User").Find(&assignees).Error
	return assignees, err
}

// ─── TaskDependency ───────────────────────────────────────────────

func (r *TaskRepo) CreateDependency(dep *TaskDependencyModel) error {
	return r.db.Create(dep).Error
}

func (r *TaskRepo) FindDependency(id string) (*TaskDependencyModel, error) {
	var d TaskDependencyModel
	err := r.db.Where("id = ?", id).First(&d).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &d, err
}

func (r *TaskRepo) RemoveDependency(id string) error {
	return r.db.Delete(&TaskDependencyModel{}, "id = ?", id).Error
}

// ─── TaskType ─────────────────────────────────────────────────────

func (r *TaskRepo) ListTaskTypes() ([]TaskTypeModel, error) {
	var types []TaskTypeModel
	err := r.db.Where("is_active = true").Find(&types).Error
	return types, err
}
