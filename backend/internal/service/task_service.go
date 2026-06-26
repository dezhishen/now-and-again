package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/dezhishen/now-and-again/backend/internal/logger"
	"github.com/dezhishen/now-and-again/backend/internal/repository"
	"github.com/dezhishen/now-and-again/backend/internal/scheduler"
	"github.com/dezhishen/now-and-again/shared/types"
)

type TaskService struct {
	repo      *repository.TaskRepo
	scheduler *scheduler.Scheduler
}

func NewTaskService(repo *repository.TaskRepo, sched *scheduler.Scheduler) *TaskService {
	svc := &TaskService{repo: repo, scheduler: sched}
	// Load existing enabled tasks into scheduler
	svc.loadExistingTasks()
	return svc
}

func (s *TaskService) loadExistingTasks() {
	tasks, err := s.repo.ListEnabledTasks()
	if err != nil {
		logger.Warnf("failed to load existing tasks: %v", err)
		return
	}
	for _, task := range tasks {
		s.registerToScheduler(&task)
	}
}

func (s *TaskService) registerToScheduler(task *repository.TaskTemplateModel) {
	if !task.Enabled {
		s.scheduler.RemoveJob(task.ID)
		return
	}

	s.scheduler.RegisterJob(&scheduler.JobBuilder{
		TaskID:       task.ID,
		ScheduleType: task.ScheduleType,
		ScheduleData: task.ScheduleData,
		Callback: func(taskID string, triggeredAt time.Time) error {
			err := s.onTaskTriggered(taskID, task.FamilyID)
			// Auto-disable one-time tasks after first trigger
			if task.ScheduleType == "once" {
				task.Enabled = false
				s.repo.UpdateTask(task)
			}
			return err
		},
	})
}

func (s *TaskService) onTaskTriggered(taskID, familyID string) error {
	now := time.Now()
	has, _ := s.repo.HasPendingTodoForTaskToday(taskID, now)
	if has {
		return nil
	}
	// Get task to copy location_id
	task, err := s.repo.FindTaskByID(taskID)
	if err != nil {
		return err
	}
	window := scheduleWindow(task.ScheduleType, task.ScheduleData)
	todo := &repository.TodoModel{
		TaskID:     taskID,
		FamilyID:   familyID,
		LocationID: task.LocationID,
		DueStart:   now,
		DueDate:    now.Add(window),
		Status:     "pending",
		TodoType:   task.TodoType(),
	}
	if err := s.repo.CreateTodo(todo); err != nil {
		return err
	}
	s.repo.UpdateLastTodoAt(taskID, now)
	return nil
}

// ─── Task Template ───────────────────────────────────────────────

func (s *TaskService) Create(ctx context.Context, familyID uuid.UUID, req *types.CreateTaskRequest) (*types.TaskTemplate, error) {
	userID := ctx.Value("user_id").(string)

	dataJSON, _ := json.Marshal(req.ScheduleData)
	t := &repository.TaskTemplateModel{
		FamilyID:         familyID.String(),
		GroupID:          req.GroupID,
		LocationID:       req.LocationID,
		Name:             req.Name,
		ScheduleType:     req.ScheduleType,
		ScheduleData:     string(dataJSON),
		Enabled:          true,
		IsInspection:     req.IsInspection,
		InspectionConfig: marshalJSON(req.InspectionConfig),
		CreatedBy:        userID,
	}
	if err := s.repo.CreateTask(t); err != nil {
		return nil, err
	}
	// Register in scheduler
	s.registerToScheduler(t)
	// Generate the first todo immediately
	s.onTaskTriggered(t.ID, t.FamilyID)
	return taskModelToType(t), nil
}

func (s *TaskService) List(ctx context.Context, familyID uuid.UUID) ([]types.TaskTemplate, error) {
	tasks, err := s.repo.ListTasksByFamily(familyID.String())
	if err != nil {
		return nil, err
	}
	result := make([]types.TaskTemplate, len(tasks))
	for i, t := range tasks {
		result[i] = *taskModelToType(&t)
	}
	return result, nil
}

func (s *TaskService) Update(ctx context.Context, taskID uuid.UUID, req *types.UpdateTaskRequest) (*types.TaskTemplate, error) {
	t, err := s.repo.FindTaskByID(taskID.String())
	if err != nil {
		return nil, fmt.Errorf("task not found")
	}
	if req.Name != nil {
		t.Name = *req.Name
	}
	if req.ScheduleType != nil {
		t.ScheduleType = *req.ScheduleType
	}
	if req.ScheduleData != nil {
		dataJSON, _ := json.Marshal(req.ScheduleData)
		t.ScheduleData = string(dataJSON)
	}
	if req.Enabled != nil {
		t.Enabled = *req.Enabled
	}
	if req.GroupID != nil {
		t.GroupID = *req.GroupID
	}
	if req.LocationID != nil {
		t.LocationID = *req.LocationID
	}
	if req.IsInspection != nil {
		t.IsInspection = *req.IsInspection
	}
	if req.InspectionConfig != nil {
		t.InspectionConfig = marshalJSON(req.InspectionConfig)
	}
	if err := s.repo.UpdateTask(t); err != nil {
		return nil, err
	}
	// Keep scheduler in sync
	s.registerToScheduler(t)
	return taskModelToType(t), nil
}

func (s *TaskService) Delete(ctx context.Context, taskID uuid.UUID) error {
	s.scheduler.RemoveJob(taskID.String())
	return s.repo.DeleteTask(taskID.String())
}

// ─── Logs ────────────────────────────────────────────────────────

func (s *TaskService) ListLogs(ctx context.Context, taskID uuid.UUID, limit int) ([]types.TaskLog, error) {
	if limit <= 0 {
		limit = 50
	}
	logs, err := s.repo.ListLogs(taskID.String(), limit)
	if err != nil {
		return nil, err
	}
	result := make([]types.TaskLog, len(logs))
	for i, l := range logs {
		result[i] = types.TaskLog{
			ID: l.ID, TaskID: l.TaskID, Status: l.Status,
			Message: l.Message, CreatedAt: l.CreatedAt,
		}
	}
	return result, nil
}

// ─── Todo ────────────────────────────────────────────────────────

func (s *TaskService) ListTodos(ctx context.Context, familyID uuid.UUID, groupID, status string) ([]types.Todo, error) {
	todos, err := s.repo.ListTodosByFamily(familyID.String(), status)
	if err != nil {
		return nil, err
	}
	result := make([]types.Todo, 0, len(todos))
	for _, t := range todos {
		// Filter by group if the task has a group_id
		if groupID != "" && t.Task.GroupID != "" && t.Task.GroupID != groupID {
			continue
		}
		result = append(result, *todoModelToType(&t))
	}
	return result, nil
}

func (s *TaskService) CompleteTodo(ctx context.Context, todoID uuid.UUID, req *types.CompleteTodoRequest) (*types.Todo, error) {
	userID := ctx.Value("user_id").(string)

	// First load the todo to check its type
	todo, err := s.repo.FindTodoByID(todoID.String())
	if err != nil {
		return nil, err
	}

	// For inspection todos, use special completion flow
	if todo.TodoType == "inspection" && req.InspectionResult != "" {
		if err := s.repo.CompleteInspection(todoID.String(), userID, req.InspectionResult); err != nil {
			return nil, err
		}
		// If abnormal, create a follow-up todo based on inspection config
		if req.InspectionResult == "abnormal" && todo.Task.InspectionConfig != "" {
			s.handleAbnormalInspection(todo, userID)
		}
	} else {
		if err := s.repo.CompleteTodo(todoID.String(), userID, req.Status); err != nil {
			return nil, err
		}
	}

	t, err := s.repo.FindTodoByID(todoID.String())
	if err != nil {
		return nil, err
	}
	return todoModelToType(t), nil
}

// handleAbnormalInspection creates a follow-up todo when inspection is abnormal
func (s *TaskService) handleAbnormalInspection(todo *repository.TodoModel, userID string) {
	var config struct {
		AbnormalAction     string `json:"abnormal_action"`
		AbnormalTaskName   string `json:"abnormal_task_name"`
		AbnormalGroupID    string `json:"abnormal_group_id"`
		AbnormalLocationID string `json:"abnormal_location_id"`
	}
	if err := json.Unmarshal([]byte(todo.Task.InspectionConfig), &config); err != nil {
		logger.Warnf("failed to parse inspection config: %v", err)
		return
	}
	if config.AbnormalAction != "create_todo" {
		return
	}
	// Generate follow-up name: replace {name} with the original task name
	name := config.AbnormalTaskName
	if name == "" {
		name = "修复: " + todo.Task.Name
	}
	// Simple template replacement
	name = strings.Replace(name, "{name}", todo.Task.Name, -1)

	followUp := &repository.TodoModel{
		TaskID:     todo.TaskID, // link to same template
		FamilyID:   todo.FamilyID,
		LocationID: firstNonEmpty(config.AbnormalLocationID, todo.LocationID),
		AssignedTo: config.AbnormalGroupID,
		Status:     "pending",
		TodoType:   "task",
		DueStart:   time.Now(),
		DueDate:    time.Now().Add(24 * time.Hour),
	}
	if err := s.repo.CreateTodo(followUp); err != nil {
		logger.Errorf("failed to create abnormal follow-up todo: %v", err)
	}
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

// ─── Helpers ─────────────────────────────────────────────────────

// scheduleWindow returns the duration until the next trigger for a schedule type.
func scheduleWindow(scheduleType, scheduleData string) time.Duration {
	switch scheduleType {
	case "daily":
		return 24 * time.Hour
	case "weekly":
		return 7 * 24 * time.Hour
	case "monthly":
		return 30 * 24 * time.Hour // approximate
	case "once":
		return 24 * time.Hour // one-time tasks have a 24h window
	case "interval":
		var data struct {
			Days int    `json:"days"`
			Time string `json:"time"`
		}
		if err := json.Unmarshal([]byte(scheduleData), &data); err == nil && data.Days > 0 {
			return time.Duration(data.Days) * 24 * time.Hour
		}
		return 24 * time.Hour
	default:
		return 24 * time.Hour
	}
}

func marshalJSON(v any) string {
	if v == nil {
		return ""
	}
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}

func taskModelToType(t *repository.TaskTemplateModel) *types.TaskTemplate {
	var data any
	json.Unmarshal([]byte(t.ScheduleData), &data)
	var inspCfg any
	if t.InspectionConfig != "" {
		json.Unmarshal([]byte(t.InspectionConfig), &inspCfg)
	}
	return &types.TaskTemplate{
		ID: t.ID, FamilyID: t.FamilyID, GroupID: t.GroupID,
		LocationID: t.LocationID,
		Name:       t.Name, ScheduleType: t.ScheduleType, ScheduleData: data,
		Enabled: t.Enabled, IsInspection: t.IsInspection,
		InspectionConfig: inspCfg,
		LastTodoAt:       t.LastTodoAt,
		CreatedBy:        t.CreatedBy, CreatedAt: t.CreatedAt, UpdatedAt: t.UpdatedAt,
	}
}

func todoModelToType(t *repository.TodoModel) *types.Todo {
	var task *types.TaskTemplate
	if t.Task.ID != "" {
		task = taskModelToType(&t.Task)
	}
	var user *types.User
	if t.User.ID != "" {
		user = userModelToUser(&t.User)
	}
	return &types.Todo{
		ID: t.ID, TaskID: t.TaskID, FamilyID: t.FamilyID,
		LocationID: t.LocationID,
		AssignedTo: t.AssignedTo, Status: t.Status, TodoType: t.TodoType,
		InspectionResult: t.InspectionResult,
		DueStart:         t.DueStart,
		DueDate:          t.DueDate,
		CompletedAt:      t.CompletedAt, CompletedBy: t.CompletedBy,
		Task: task, User: user,
		CreatedAt: t.CreatedAt, UpdatedAt: t.UpdatedAt,
	}
}
