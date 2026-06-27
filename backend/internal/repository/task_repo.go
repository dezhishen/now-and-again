package repository

import (
	"time"

	"github.com/dezhishen/now-and-again/backend/pkg/timeutil"
	"gorm.io/gorm"
)

// ─── Transaction ─────────────────────────────────────────────────

// Tx runs fn in a transaction, creating a new TaskRepo with the tx DB.
// All operations inside fn use the same transaction.
func (r *TaskRepo) Tx(fn func(tx *TaskRepo) error) error {
	return r.db.Transaction(func(txDB *gorm.DB) error {
		return fn(&TaskRepo{db: txDB})
	})
}

// ─── Task Template ───────────────────────────────────────────────

func (r *TaskRepo) CreateTask(t *TaskModel) error {
	return r.db.Create(t).Error
}

func (r *TaskRepo) FindTaskByID(id string) (*TaskModel, error) {
	var t TaskModel
	err := r.db.Preload("Group").Where("id = ?", id).First(&t).Error
	return &t, err
}

func (r *TaskRepo) ListTasksByFamily(familyID string) ([]TaskModel, error) {
	var tasks []TaskModel
	err := r.db.Preload("Group").Where("family_id = ? AND is_root = ?", familyID, true).Order("created_at ASC").Find(&tasks).Error
	return tasks, err
}

// FindTaskFull loads a task with all sub-tables for detail view / plugin actions.
func (r *TaskRepo) FindTaskFull(id string) (*TaskModel, error) {
	var t TaskModel
	err := r.db.Preload("Group").Preload("CheckItems.Branches.BranchTask").Preload("Children").Where("id = ?", id).First(&t).Error
	return &t, err
}

func (r *TaskRepo) ListEnabledTasks() ([]TaskModel, error) {
	var tasks []TaskModel
	err := r.db.Where("enabled = ?", true).Find(&tasks).Error
	return tasks, err
}

func (r *TaskRepo) UpdateTask(t *TaskModel) error {
	return r.db.Save(t).Error
}

// DisableTask sets enabled=false on a single task without touching other columns.
func (r *TaskRepo) DisableTask(id string) error {
	return r.db.Model(&TaskModel{}).Where("id = ?", id).Update("enabled", false).Error
}

func (r *TaskRepo) DeleteTask(id string) error {
	return r.db.Where("id = ?", id).Delete(&TaskModel{}).Error
}

func (r *TaskRepo) UpdateLastTodoAt(taskID string, t time.Time) error {
	return r.db.Model(&TaskModel{}).Where("id = ?", taskID).Update("last_todo_at", t).Error
}

func (r *TaskRepo) UpdateDisplaySummary(taskID, summary string) error {
	return r.db.Model(&TaskModel{}).Where("id = ?", taskID).Update("display_summary", summary).Error
}

// ─── Todo ────────────────────────────────────────────────────────

func (r *TaskRepo) CreateTodo(todo *TodoModel) error {
	return r.db.Create(todo).Error
}

func (r *TaskRepo) FindTodoByID(id string) (*TodoModel, error) {
	var t TodoModel
	err := r.db.Preload("Task").Preload("User").Where("id = ?", id).First(&t).Error
	return &t, err
}

// FindTodoFull loads a todo with full task detail (check items, children).
func (r *TaskRepo) FindTodoFull(id string) (*TodoModel, error) {
	var t TodoModel
	err := r.db.Preload("Task.CheckItems.Branches").Preload("Task.Children").Preload("User").Where("id = ?", id).First(&t).Error
	return &t, err
}

func (r *TaskRepo) ListTodosByFamily(familyID string, status string) ([]TodoModel, error) {
	var todos []TodoModel
	q := r.db.Preload("Task").Preload("User").Where("family_id = ?", familyID)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	err := q.Order("due_date ASC").Find(&todos).Error
	return todos, err
}

func (r *TaskRepo) ListTodosByUser(userID string, status string) ([]TodoModel, error) {
	var todos []TodoModel
	q := r.db.Preload("Task").Preload("User").Where("assigned_to = ?", userID)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	err := q.Order("due_date ASC").Find(&todos).Error
	return todos, err
}

func (r *TaskRepo) CompleteTodo(id, userID, status, remark string) error {
	now := timeutil.Now()
	return r.db.Model(&TodoModel{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":       status,
		"remark":       remark,
		"completed_at": now,
		"completed_by": userID,
	}).Error
}

func (r *TaskRepo) CompleteInspection(id, userID, inspectionResult string) error {
	now := timeutil.Now()
	return r.db.Model(&TodoModel{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":            "done",
		"inspection_result": inspectionResult,
		"completed_at":      now,
		"completed_by":      userID,
	}).Error
}

func (r *TaskRepo) DeleteTodo(id string) error {
	return r.db.Where("id = ?", id).Delete(&TodoModel{}).Error
}

func (r *TaskRepo) HasPendingTodoForTaskToday(taskID string, today time.Time) (bool, error) {
	var count int64
	startOfDay := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)
	err := r.db.Model(&TodoModel{}).
		Where("task_id = ? AND status = 'pending' AND created_at >= ? AND created_at < ?", taskID, startOfDay, endOfDay).
		Count(&count).Error
	return count > 0, err
}

// ─── Task Log ────────────────────────────────────────────────────

func (r *TaskRepo) CreateLog(taskID, status, message string) error {
	return r.db.Create(&TaskLogModel{TaskID: taskID, Status: status, Message: message, LogType: "system"}).Error
}

func (r *TaskRepo) CreateUserLog(taskID, todoID, userID, action, message string) error {
	return r.db.Create(&TaskLogModel{
		TaskID: taskID, TodoID: todoID,
		Status: action, Message: message, LogType: "user", OperatorID: userID,
	}).Error
}

func (r *TaskRepo) ListLogs(taskID string, limit, offset int) ([]TaskLogModel, error) {
	var logs []TaskLogModel
	err := r.db.Where("task_id = ?", taskID).Order("created_at DESC").Limit(limit).Offset(offset).Find(&logs).Error
	return logs, err
}

func (r *TaskRepo) ListUserLogs(taskID string, limit, offset int) ([]TaskLogModel, error) {
	var logs []TaskLogModel
	err := r.db.Where("task_id = ? AND log_type = ?", taskID, "user").Order("created_at DESC").Limit(limit).Offset(offset).Find(&logs).Error
	return logs, err
}

// ListLogsByFamily returns all task logs for a family within a date range,
// joined through task_templates.
func (r *TaskRepo) ListLogsByFamily(familyID string, since, until string) ([]TaskLogModel, error) {
	var logs []TaskLogModel
	err := r.db.
		Table("task_logs").
		Joins("JOIN task_templates ON task_templates.id = task_logs.task_id").
		Where("task_templates.family_id = ?", familyID).
		Where("task_logs.created_at >= ? AND task_logs.created_at < ?", since, until).
		Order("task_logs.created_at ASC").
		Find(&logs).Error
	return logs, err
}

// ListLogsByFamilyAndTask returns task logs for a specific task within a date range.
func (r *TaskRepo) ListLogsByFamilyAndTask(familyID, taskID, since, until string) ([]TaskLogModel, error) {
	var logs []TaskLogModel
	err := r.db.
		Table("task_logs").
		Joins("JOIN task_templates ON task_templates.id = task_logs.task_id").
		Where("task_templates.family_id = ? AND task_logs.task_id = ?", familyID, taskID).
		Where("task_logs.created_at >= ? AND task_logs.created_at < ?", since, until).
		Order("task_logs.created_at ASC").
		Find(&logs).Error
	return logs, err
}

// ─── Inspection Result ───────────────────────────────────────────

func (r *TaskRepo) CreateInspectionResult(ir *InspectionResultModel) error {
	return r.db.Create(ir).Error
}

// ─── Check Items ─────────────────────────────────────────────────

func (r *TaskRepo) CreateCheckItem(item *CheckItemModel) error {
	return r.db.Create(item).Error
}

func (r *TaskRepo) CreateCheckItemBranch(branch *CheckItemBranchModel) error {
	return r.db.Create(branch).Error
}

func (r *TaskRepo) DeleteCheckItemsByTask(taskID string) error {
	// Delete branches first, then items
	r.db.Where("check_item_id IN (SELECT id FROM check_items WHERE task_id = ?)", taskID).Delete(&CheckItemBranchModel{})
	return r.db.Where("task_id = ?", taskID).Delete(&CheckItemModel{}).Error
}

// DeleteChildren removes child tasks (is_root=false) for a given parent.
func (r *TaskRepo) DeleteChildren(parentTaskID string) error {
	return r.db.Where("parent_task_id = ? AND is_root = ?", parentTaskID, false).Delete(&TaskModel{}).Error
}

// FindBranchTask looks up the branch task for a given inspection selection.
func (r *TaskRepo) FindBranchTask(taskID, itemName, branchName string) (*TaskModel, error) {
	var t TaskModel
	err := r.db.
		Table("tasks").
		Joins("JOIN check_item_branches ON check_item_branches.branch_task_id = tasks.id").
		Joins("JOIN check_items ON check_items.id = check_item_branches.check_item_id").
		Where("check_items.task_id = ? AND check_items.name = ? AND check_item_branches.name = ?", taskID, itemName, branchName).
		First(&t).Error
	return &t, err
}
