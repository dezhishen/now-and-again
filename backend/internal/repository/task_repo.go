package repository

import (
	"encoding/json"
	"time"

	"github.com/dezhishen/now-and-again/backend/pkg/timeutil"
	"github.com/dezhishen/now-and-again/backend/pkg/types"
	"gorm.io/gorm"
)

// ─── Transaction ─────────────────────────────────────────────────

// DB returns the underlying *gorm.DB.
func (r *TaskRepo) DB() *gorm.DB { return r.db }

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

func (r *TaskRepo) FindTaskByParentId(parentID string) (*TaskModel, error) {
	var t TaskModel
	err := r.db.Preload("Group").Where("parent_task_id = ?", parentID).First(&t).Error
	return &t, err
}

func (r *TaskRepo) ListTasksByFamily(familyID string) ([]TaskModel, error) {
	var tasks []TaskModel
	err := r.db.Preload("Group").Where("family_id = ? AND is_root = ?", familyID, true).Order("created_at ASC").Find(&tasks).Error
	return tasks, err
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

// ─── Todo ────────────────────────────────────────────────────────

func (r *TaskRepo) CreateTodo(todo *TodoModel) error {
	return r.db.Create(todo).Error
}

func (r *TaskRepo) FindTodoByID(id string) (*TodoModel, error) {
	var t TodoModel
	err := r.db.Preload("Task").Preload("User").Where("id = ?", id).First(&t).Error
	return &t, err
}

// FindTodoFull delegates to FindTodoByID (identical query; kept for semantic clarity).
func (r *TaskRepo) FindTodoFull(id string) (*TodoModel, error) {
	return r.FindTodoByID(id)
}

func (r *TaskRepo) ListTodosByFamily(familyID string, status string) ([]TodoModel, error) {
	var todos []TodoModel
	q := r.db.Preload("Task.Group").Preload("User").Where("family_id = ?", familyID)
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

// CompleteTodo marks a todo as done/skipped. Only updates if currently 'pending'.
// Returns true if a row was actually updated (idempotent — duplicate calls are safe).
func (r *TaskRepo) CompleteTodo(id, userID, status, remark string) (bool, error) {
	now := timeutil.Now()
	result := r.db.Model(&TodoModel{}).
		Where("id = ? AND status = ?", id, "pending").
		Updates(map[string]interface{}{
			"status":       status,
			"remark":       remark,
			"completed_at": now,
			"completed_by": userID,
		})
	return result.RowsAffected > 0, result.Error
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

// childTaskIDs returns a subquery selecting IDs of tasks whose parent is taskID.
func (r *TaskRepo) childTaskIDs(taskID string) *gorm.DB {
	return r.db.Model(&TaskModel{}).Select("id").Where("parent_task_id = ?", taskID)
}

// ListLogs returns logs for a task AND its child tasks, with Task names preloaded.
func (r *TaskRepo) ListLogs(taskID string, limit, offset int) ([]TaskLogModel, error) {
	var logs []TaskLogModel
	err := r.db.
		Preload("Task").
		Where("task_id = ? OR task_id IN (?)", taskID, r.childTaskIDs(taskID)).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&logs).Error
	return logs, err
}

// ListUserLogs returns user-generated logs for a task AND its child tasks.
func (r *TaskRepo) ListUserLogs(taskID string, limit, offset int) ([]TaskLogModel, error) {
	var logs []TaskLogModel
	err := r.db.
		Preload("Task").
		Where("(task_id = ? OR task_id IN (?)) AND log_type = ?", taskID, r.childTaskIDs(taskID), "user").
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&logs).Error
	return logs, err
}

// ListLogsByFamily returns all task logs for a family within a date range,
// using GORM model-driven JOIN through TaskLogModel.Task relation.
func (r *TaskRepo) ListLogsByFamily(familyID string, since, until string) ([]TaskLogModel, error) {
	var logs []TaskLogModel
	err := r.db.
		Joins("Task").
		Where("Task.family_id = ?", familyID).
		Where("task_logs.created_at >= ? AND task_logs.created_at < ?", since, until).
		Order("task_logs.created_at ASC").
		Find(&logs).Error
	return logs, err
}

// ListLogsByFamilyAndTask returns task logs for a specific task within a date range.
func (r *TaskRepo) ListLogsByFamilyAndTask(familyID, taskID, since, until string) ([]TaskLogModel, error) {
	var logs []TaskLogModel
	err := r.db.
		Joins("Task").
		Where("Task.family_id = ? AND task_logs.task_id = ?", familyID, taskID).
		Where("task_logs.created_at >= ? AND task_logs.created_at < ?", since, until).
		Order("task_logs.created_at ASC").
		Find(&logs).Error
	return logs, err
}

// DeleteChildren removes child tasks (is_root=false) for a given parent.
func (r *TaskRepo) DeleteChildren(parentTaskID string) error {
	return r.db.Where("parent_task_id = ? AND is_root = ?", parentTaskID, false).Delete(&TaskModel{}).Error
}

// TaskModelToType converts a TaskModel to the public types.Task DTO.
func TaskModelToType(t *TaskModel) *types.Task {
	var data any
	json.Unmarshal([]byte(t.ScheduleData), &data)
	return &types.Task{
		ID: t.ID, FamilyID: t.FamilyID, GroupID: t.GroupID,
		ParentTaskID: t.ParentTaskID, IsRoot: t.IsRoot,
		LocationID: t.LocationID,
		Name:       t.Name, ScheduleType: t.ScheduleType, ScheduleData: data,
		Enabled: t.Enabled, Kind: t.Kind, DisplaySummary: t.DisplaySummary,
		LastTodoAt: t.LastTodoAt,
		CreatedBy:  t.CreatedBy, CreatedAt: t.CreatedAt, UpdatedAt: t.UpdatedAt,
	}
}
