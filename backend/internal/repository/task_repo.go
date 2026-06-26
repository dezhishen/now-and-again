package repository

import "time"

// ─── Task Template ───────────────────────────────────────────────

func (r *TaskRepo) CreateTask(t *TaskTemplateModel) error {
	return r.db.Create(t).Error
}

func (r *TaskRepo) FindTaskByID(id string) (*TaskTemplateModel, error) {
	var t TaskTemplateModel
	err := r.db.Where("id = ?", id).First(&t).Error
	return &t, err
}

func (r *TaskRepo) ListTasksByFamily(familyID string) ([]TaskTemplateModel, error) {
	var tasks []TaskTemplateModel
	err := r.db.Where("family_id = ?", familyID).Order("created_at ASC").Find(&tasks).Error
	return tasks, err
}

func (r *TaskRepo) ListEnabledTasks() ([]TaskTemplateModel, error) {
	var tasks []TaskTemplateModel
	err := r.db.Where("enabled = ?", true).Find(&tasks).Error
	return tasks, err
}

func (r *TaskRepo) UpdateTask(t *TaskTemplateModel) error {
	return r.db.Save(t).Error
}

func (r *TaskRepo) DeleteTask(id string) error {
	return r.db.Where("id = ?", id).Delete(&TaskTemplateModel{}).Error
}

func (r *TaskRepo) UpdateLastTodoAt(taskID string, t time.Time) error {
	return r.db.Model(&TaskTemplateModel{}).Where("id = ?", taskID).Update("last_todo_at", t).Error
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

func (r *TaskRepo) CompleteTodo(id, userID, status string) error {
	now := time.Now()
	return r.db.Model(&TodoModel{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":       status,
		"completed_at": now,
		"completed_by": userID,
	}).Error
}

func (r *TaskRepo) CompleteInspection(id, userID, inspectionResult string) error {
	now := time.Now()
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
		Where("task_id = ? AND created_at >= ? AND created_at < ?", taskID, startOfDay, endOfDay).
		Count(&count).Error
	return count > 0, err
}

// ─── Task Log ────────────────────────────────────────────────────

func (r *TaskRepo) CreateLog(taskID, status, message string) error {
	return r.db.Create(&TaskLogModel{TaskID: taskID, Status: status, Message: message, LogType: "system"}).Error
}

func (r *TaskRepo) CreateUserLog(taskID, userID, action, message string) error {
	return r.db.Create(&TaskLogModel{TaskID: taskID, Status: action, Message: message, LogType: "user", OperatorID: userID}).Error
}

func (r *TaskRepo) ListLogs(taskID string, limit int) ([]TaskLogModel, error) {
	var logs []TaskLogModel
	err := r.db.Where("task_id = ?", taskID).Order("created_at DESC").Limit(limit).Find(&logs).Error
	return logs, err
}

func (r *TaskRepo) ListUserLogs(taskID string, limit int) ([]TaskLogModel, error) {
	var logs []TaskLogModel
	err := r.db.Where("task_id = ? AND log_type = ?", taskID, "user").Order("created_at DESC").Limit(limit).Find(&logs).Error
	return logs, err
}
