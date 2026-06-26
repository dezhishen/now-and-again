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
			if task.ScheduleType == "once" {
				task.Enabled = false
				s.repo.UpdateTask(task)
			}
			return err
		},
	})
}

func (s *TaskService) onTaskTriggered(taskID, familyID string) error {
	return s.createTodo(taskID, familyID, false)
}

func (s *TaskService) createTodo(taskID, familyID string, force bool) error {
	now := time.Now()
	if !force {
		has, _ := s.repo.HasPendingTodoForTaskToday(taskID, now)
		if has {
			return nil
		}
	}
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
	kind := req.Kind
	if kind == "" {
		kind = "simple"
	}
	dataJSON, _ := json.Marshal(req.ScheduleData)
	t := &repository.TaskTemplateModel{
		FamilyID:     familyID.String(),
		GroupID:      req.GroupID,
		LocationID:   req.LocationID,
		Name:         req.Name,
		ScheduleType: req.ScheduleType,
		ScheduleData: string(dataJSON),
		Enabled:      true,
		Kind:         kind,
		Branches:     marshalJSON(req.Branches),
		CreatedBy:    userID,
	}
	if err := s.repo.CreateTask(t); err != nil {
		return nil, err
	}
	s.registerToScheduler(t)
	s.onTaskTriggered(t.ID, t.FamilyID)
	return taskModelToType(t), nil
}

func (s *TaskService) Trigger(ctx context.Context, taskID uuid.UUID) error {
	userID := ctx.Value("user_id").(string)
	task, err := s.repo.FindTaskByID(taskID.String())
	if err != nil {
		return err
	}
	s.repo.CreateUserLog(taskID.String(), userID, "manual", fmt.Sprintf("手动生成待办: %s", task.Name))
	return s.createTodo(task.ID, task.FamilyID, true)
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
	if req.Kind != nil {
		t.Kind = *req.Kind
	}
	if req.Branches != nil {
		t.Branches = marshalJSON(req.Branches)
	}
	if err := s.repo.UpdateTask(t); err != nil {
		return nil, err
	}
	s.registerToScheduler(t)
	return taskModelToType(t), nil
}

func (s *TaskService) Delete(ctx context.Context, taskID uuid.UUID) error {
	s.scheduler.RemoveJob(taskID.String())
	return s.repo.DeleteTask(taskID.String())
}

// ─── Logs ────────────────────────────────────────────────────────

func (s *TaskService) ListLogs(ctx context.Context, taskID uuid.UUID, limit int, userOnly bool) ([]types.TaskLog, error) {
	if limit <= 0 {
		limit = 50
	}
	var logs []repository.TaskLogModel
	var err error
	if userOnly {
		logs, err = s.repo.ListUserLogs(taskID.String(), limit)
	} else {
		logs, err = s.repo.ListLogs(taskID.String(), limit)
	}
	if err != nil {
		return nil, err
	}
	result := make([]types.TaskLog, len(logs))
	for i, l := range logs {
		result[i] = types.TaskLog{
			ID: l.ID, TaskID: l.TaskID, Status: l.Status,
			Message: l.Message, LogType: l.LogType,
			OperatorID: l.OperatorID, CreatedAt: l.CreatedAt,
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
		if groupID != "" && t.Task.GroupID != "" && t.Task.GroupID != groupID {
			continue
		}
		result = append(result, *todoModelToType(&t))
	}
	return result, nil
}

func (s *TaskService) CompleteTodo(ctx context.Context, todoID uuid.UUID, req *types.CompleteTodoRequest) (*types.Todo, error) {
	userID := ctx.Value("user_id").(string)
	todo, err := s.repo.FindTodoByID(todoID.String())
	if err != nil {
		return nil, err
	}

	// Branch completion flow
	if todo.Task.Kind == "branched" && req.BranchName != "" {
		if err := s.repo.CompleteTodo(todoID.String(), userID, "done"); err != nil {
			return nil, err
		}
		s.repo.CreateUserLog(todo.TaskID, userID, "branch:"+req.BranchName,
			fmt.Sprintf("巡检完成 [%s]: %s", req.BranchName, todo.Task.Name))
		s.handleBranchFollowUp(todo, req.BranchName, userID)
	} else {
		if err := s.repo.CompleteTodo(todoID.String(), userID, req.Status); err != nil {
			return nil, err
		}
		s.repo.CreateUserLog(todo.TaskID, userID, req.Status,
			fmt.Sprintf("完成待办: %s", todo.Task.Name))
	}

	// Auto-disable one-time tasks
	if todo.Task.ScheduleType == "once" && req.Status == "done" {
		s.repo.UpdateTask(&repository.TaskTemplateModel{
			BaseModel: repository.BaseModel{ID: todo.TaskID},
			Enabled:   false,
		})
		s.scheduler.RemoveJob(todo.TaskID)
	}

	t, err := s.repo.FindTodoByID(todoID.String())
	if err != nil {
		return nil, err
	}
	return todoModelToType(t), nil
}

// handleBranchFollowUp creates an independent one-time task for the selected branch.
func (s *TaskService) handleBranchFollowUp(todo *repository.TodoModel, branchName string, userID string) {
	type branch struct {
		Name      string `json:"name"`
		CreateTodo bool   `json:"create_todo"`
		TodoName  string `json:"todo_name"`
		GroupID   string `json:"group_id"`
	}
	var branches []branch
	if err := json.Unmarshal([]byte(todo.Task.Branches), &branches); err != nil {
		logger.Warnf("failed to parse branches: %v", err)
		return
	}
	for _, b := range branches {
		if b.Name == branchName && b.CreateTodo {
			name := b.TodoName
			if name == "" {
				name = todo.Task.Name + " - " + branchName
			}
			name = strings.Replace(name, "{name}", todo.Task.Name, -1)

			// Create independent one-time task
			followTask := &repository.TaskTemplateModel{
				FamilyID:     todo.FamilyID,
				GroupID:      b.GroupID,
				LocationID:   todo.LocationID,
				Name:         name,
				ScheduleType: "once",
				ScheduleData: `{"time":"00:00"}`,
				Enabled:      true,
				Kind:         "simple",
				CreatedBy:    userID,
			}
			if err := s.repo.CreateTask(followTask); err != nil {
				logger.Errorf("failed to create follow-up task: %v", err)
				return
			}

			// Log on both sides
			s.repo.CreateUserLog(todo.TaskID, userID, "follow_up",
				fmt.Sprintf("分支「%s」→ 创建跟进任务「%s」(%s)", branchName, name, followTask.ID))
			s.repo.CreateUserLog(followTask.ID, userID, "created",
				fmt.Sprintf("从巡检「%s」分支「%s」创建", todo.Task.Name, branchName))

			// Create todo for the follow-up task
			s.createTodo(followTask.ID, todo.FamilyID, true)
			return
		}
	}
}

// ─── Helpers ─────────────────────────────────────────────────────

func scheduleWindow(scheduleType, scheduleData string) time.Duration {
	switch scheduleType {
	case "daily":
		return 24 * time.Hour
	case "weekly":
		return 7 * 24 * time.Hour
	case "monthly":
		return 30 * 24 * time.Hour
	case "once":
		return 24 * time.Hour
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
	var branches any
	if t.Branches != "" {
		json.Unmarshal([]byte(t.Branches), &branches)
	}
	return &types.TaskTemplate{
		ID: t.ID, FamilyID: t.FamilyID, GroupID: t.GroupID,
		LocationID: t.LocationID,
		Name: t.Name, ScheduleType: t.ScheduleType, ScheduleData: data,
		Enabled: t.Enabled, Kind: t.Kind, Branches: branches,
		LastTodoAt: t.LastTodoAt,
		CreatedBy: t.CreatedBy, CreatedAt: t.CreatedAt, UpdatedAt: t.UpdatedAt,
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
		AssignedTo: t.AssignedTo, Status: t.Status, BranchName: t.BranchName,
		DueStart:    t.DueStart,
		DueDate:     t.DueDate,
		CompletedAt: t.CompletedAt, CompletedBy: t.CompletedBy,
		Task: task, User: user,
		CreatedAt: t.CreatedAt, UpdatedAt: t.UpdatedAt,
	}
}
