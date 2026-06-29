package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/dezhishen/now-and-again/backend/internal/logger"
	"github.com/dezhishen/now-and-again/backend/internal/repository"
	"github.com/dezhishen/now-and-again/backend/pkg/scheduler"
	"github.com/dezhishen/now-and-again/backend/pkg/taskkind"
	"github.com/dezhishen/now-and-again/backend/pkg/timeutil"
	"github.com/dezhishen/now-and-again/backend/pkg/types"
	"gorm.io/gorm"

	// Blank imports trigger init() registration of task kind handlers.
	_ "github.com/dezhishen/now-and-again/backend/pkg/taskkind/inspection"
	_ "github.com/dezhishen/now-and-again/backend/pkg/taskkind/simple"
)

type TaskService struct {
	*taskOrchestrator
}

type _taskStorage struct {
	repo        *repository.TaskRepo
	taskManager *taskkind.TaskManager
}

// taskOrchestrator bundles the shared dependencies for TaskService and TodoService.
type taskOrchestrator struct {
	repo        *repository.TaskRepo
	taskManager *taskkind.TaskManager
	taskStorage taskkind.TaskStorage
}

func newTaskOrchestrator(repo *repository.TaskRepo) *taskOrchestrator {
	o := &taskOrchestrator{
		repo:        repo,
		taskManager: taskkind.GetManager(),
	}
	o.taskStorage = &_taskStorage{repo: repo, taskManager: o.taskManager}
	return o
}

func (s *_taskStorage) FindTaskByID(taskID string) (*repository.TaskModel, error) {
	return s.repo.FindTaskByID(taskID)
}

func (s *_taskStorage) FindTaskByParentId(parentID string) (*repository.TaskModel, error) {
	return s.repo.FindTaskByParentId(parentID)
}

func (s *_taskStorage) UpdateNoRootTask(task *repository.TaskModel) error {
	return s.repo.UpdateTask(task)
}

// CreateNoRootTask creates a non-root task and triggers its kind's SaveExtra handler.
func (s *_taskStorage) CreateNoRootTask(task *repository.TaskModel, extra any) error {
	task.IsRoot = false
	if err := s.repo.CreateTask(task); err != nil {
		return err
	}
	if h := s.taskManager.Get(task.Kind); h != nil {
		return h.SaveExtra(s, task, extra)
	}
	return nil
}

// DeleteNonRootTask deletes a non-root task and all descendants, triggering DeleteExtra for each.
func (s *_taskStorage) DeleteNonRootTask(taskID string) error {
	task, err := s.repo.FindTaskByID(taskID)
	if err != nil || task == nil {
		return nil
	}
	// Recursively delete children first
	children, _ := s.repo.FindChildren(task.ID)
	for _, child := range children {
		if err := s.DeleteNonRootTask(child.ID); err != nil {
			return err
		}
	}
	// Plugin cleanup + task record
	if h := s.taskManager.Get(task.Kind); h != nil {
		if err := h.DeleteExtra(s, task); err != nil {
			return err
		}
	}
	return s.repo.DeleteTask(taskID)
}

func (s *_taskStorage) DB() *gorm.DB {
	return s.repo.DB()
}

func NewTaskService(repo *repository.TaskRepo) *TaskService {
	svc := &TaskService{taskOrchestrator: newTaskOrchestrator(repo)}
	svc._init()
	return svc
}

func (s *TaskService) _init() {
	tasks, err := s.repo.ListEnabledTasks()
	if err != nil {
		logger.Warnf("failed to load existing tasks: %v", err)
		return
	}
	for _, task := range tasks {
		s.registerToScheduler(&task)
	}
}

func (s *TaskService) ScheduleTasks(tasks []repository.TaskModel) {
	if len(tasks) == 0 {
		logger.Info("No tasks to schedule")
	}
	for _, t := range tasks {
		s.registerToScheduler(&t)
	}
}

// ─── TaskContract delegates ──────────────────────────────────────
// These implement shared/contracts.TaskContract so the service plugs into
// the compile-time-checked contract system used by both server and CLI.

func (s *TaskService) CreateTask(ctx context.Context, familyID uuid.UUID, req *types.CreateTaskRequest) (*types.Task, error) {
	return s.Create(ctx, familyID, req)
}
func (s *TaskService) UpdateTask(ctx context.Context, taskID uuid.UUID, req *types.UpdateTaskRequest) (*types.Task, error) {
	return s.Update(ctx, taskID, req)
}
func (s *TaskService) SetTaskEnabled(ctx context.Context, taskID uuid.UUID, enabled bool) (*types.Task, error) {
	return s.SetEnabled(ctx, taskID, enabled)
}
func (s *TaskService) DeleteTask(ctx context.Context, taskID uuid.UUID) error {
	return s.Delete(ctx, taskID)
}
func (s *TaskService) ListTasks(ctx context.Context, familyID uuid.UUID) ([]types.Task, error) {
	return s.List(ctx, familyID)
}
func (s *TaskService) TriggerTask(ctx context.Context, taskID uuid.UUID) error {
	return s.Trigger(ctx, taskID)
}

func (s *TaskService) registerToScheduler(task *repository.TaskModel) {
	if !task.Enabled {
		scheduler.Unschedule(task.ID)
		return
	}
	if err := scheduler.Schedule(scheduler.TaskInfo{
		TaskID:       task.ID,
		ScheduleType: task.ScheduleType,
		ScheduleData: task.ScheduleData,
		OnFire: func(taskID string, triggeredAt time.Time) error {
			return s.onTaskTriggered(taskID, task.FamilyID)
		},
	}); err != nil {
		logger.Warnf("scheduler: register task %s (%s) failed: %v", task.ID, task.ScheduleType, err)
	}
}

func (s *TaskService) onTaskTriggered(taskID, familyID string) error {
	// Double-check: don't create todo for disabled or deleted tasks.
	task, err := s.repo.FindTaskByID(taskID)
	if err != nil || task == nil || !task.Enabled {
		return nil
	}
	if err := s.createTodo(taskID, familyID, false); err != nil {
		return err
	}
	// One-shot tasks are archived after their single fire.
	if task.ScheduleType == "once" {
		_ = s.repo.ArchiveTask(taskID)
	}
	return nil
}

func (s *TaskService) createTodo(taskID, familyID string, force bool) error {
	return s.repo.Tx(func(tx *repository.TaskRepo) error {
		return s.createTodoWithTx(tx, taskID, familyID, force)
	})
}

func (s *TaskService) createTodoWithTx(tx *repository.TaskRepo, taskID, familyID string, force bool) error {
	now := timeutil.Now()

	task, err := tx.FindTaskByID(taskID)
	if err != nil {
		return err
	}
	window := scheduleWindow(task.ScheduleType, task.ScheduleData)

	if !force {
		has, _ := tx.HasPendingTodoForTaskToday(taskID, now)
		if has {
			return nil
		}
	}

	todo := &repository.TodoModel{
		TaskID:     taskID,
		FamilyID:   familyID,
		LocationID: task.LocationID,
		DueStart:   now,
		DueDate:    now.Add(window),
		Status:     "pending",
	}
	if err := tx.CreateTodo(todo); err != nil {
		return err
	}
	// Log system-generated todo creation so it appears in task logs.
	tx.CreateLog(taskID, "generated", fmt.Sprintf("系统自动生成待办: %s", task.Name))
	return tx.UpdateLastTodoAt(taskID, now)
}

// ─── Task Template ───────────────────────────────────────────────

func (s *TaskService) GetTask(ctx context.Context, taskID uuid.UUID) (*types.Task, error) {
	t, err := s.repo.FindTaskByID(taskID.String())
	if err != nil {
		return nil, err
	}
	return taskModelToType(t), nil
}

func (s *TaskService) GetTaskWithExtra(ctx context.Context, taskID uuid.UUID) (*types.TaskWithExtra, error) {
	t, err := s.repo.FindTaskByID(taskID.String())
	if err != nil {
		return nil, err
	}
	result := &types.TaskWithExtra{Task: taskModelToType(t)}
	if h := s.taskManager.Get(t.Kind); h != nil {
		result.Extra, _ = h.GetExtra(s.taskStorage, t)
	}
	return result, nil
}

func (s *TaskService) Create(ctx context.Context, familyID uuid.UUID, req *types.CreateTaskRequest) (*types.Task, error) {
	userID := ctx.Value("user_id").(string)
	kind := req.Task.Kind
	if kind == "" {
		kind = "simple"
	}
	dataJSON, _ := json.Marshal(req.Task.ScheduleData)
	t := &repository.TaskModel{
		FamilyID:       familyID.String(),
		GroupID:        req.Task.GroupID,
		LocationID:     req.Task.LocationID,
		IsRoot:         true,
		Name:           req.Task.Name,
		ScheduleType:   req.Task.ScheduleType,
		ScheduleData:   string(dataJSON),
		Enabled:        true,
		Kind:           kind,
		DisplaySummary: req.Task.DisplaySummary,
		CreatedBy:      userID,
	}
	if err := s.repo.Tx(func(tx *repository.TaskRepo) error {
		if err := tx.CreateTask(t); err != nil {
			return err
		}
		// Set RootTaskID after insert (root task → self)
		if err := tx.SetRootTaskID(t.ID, t.ID); err != nil {
			return err
		}
		if h := s.taskManager.Get(kind); h != nil {
			if err := h.SaveExtra(&_taskStorage{repo: tx, taskManager: s.taskManager}, t, req.Extra); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	s.registerToScheduler(t)
	t, _ = s.repo.FindTaskByID(t.ID)
	return taskModelToType(t), nil
}

func (s *TaskService) Trigger(ctx context.Context, taskID uuid.UUID) error {
	userID := ctx.Value("user_id").(string)
	task, err := s.repo.FindTaskByID(taskID.String())
	if err != nil {
		return err
	}
	if !task.Enabled {
		return fmt.Errorf("禁用的任务无法生成待办，请先启用")
	}

	// Avoid duplicate: don't create another todo if one is already pending today.
	if has, _ := s.repo.HasPendingTodoForTaskToday(taskID.String(), timeutil.Now()); has {
		return fmt.Errorf("今天已有待办，无需重复生成")
	}

	s.repo.CreateUserLog(taskID.String(), "", userID, "manual", fmt.Sprintf("手动生成待办: %s", task.Name))
	return s.createTodo(task.ID, task.FamilyID, true)
}

func (s *TaskService) List(ctx context.Context, familyID uuid.UUID) ([]types.Task, error) {
	tasks, err := s.repo.ListTasksByFamily(familyID.String())
	if err != nil {
		return nil, err
	}
	result := make([]types.Task, len(tasks))
	for i, t := range tasks {
		result[i] = *taskModelToType(&t)
	}
	return result, nil
}

func (s *TaskService) Update(ctx context.Context, taskID uuid.UUID, req *types.UpdateTaskRequest) (*types.Task, error) {
	t, err := s.repo.FindTaskByID(taskID.String())
	if err != nil {
		return nil, fmt.Errorf("task not found")
	}
	if !t.Enabled {
		return nil, fmt.Errorf("禁用的任务无法修改，请先启用")
	}
	if req.Task != nil {
		f := req.Task
		if f.Name != "" {
			t.Name = f.Name
		}
		if f.ScheduleType != "" {
			t.ScheduleType = f.ScheduleType
		}
		if f.ScheduleData != nil {
			dataJSON, _ := json.Marshal(f.ScheduleData)
			t.ScheduleData = string(dataJSON)
		}
		if f.GroupID != "" {
			t.GroupID = f.GroupID
		}
		if f.LocationID != "" {
			t.LocationID = f.LocationID
		}
		if f.Kind != "" {
			t.Kind = f.Kind
		}
		if f.DisplaySummary != "" {
			t.DisplaySummary = f.DisplaySummary
		}
	}
	if err := s.repo.Tx(func(tx *repository.TaskRepo) error {
		if err := tx.UpdateTask(t); err != nil {
			return err
		}
		// Only invoke plugin lifecycle when extra data is actually provided.
		// Toggling enabled or changing schedule fields should not delete/recreate plugin data.
		if req.Extra != nil {
			if h := s.taskManager.Get(t.Kind); h != nil {
				if err := h.UpdateExtra(&_taskStorage{repo: tx, taskManager: s.taskManager}, t, req.Extra); err != nil {
					return err
				}
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	s.registerToScheduler(t)
	t, _ = s.repo.FindTaskByID(taskID.String())
	return taskModelToType(t), nil
}

// SetEnabled enables or disables a task without touching plugin-owned data.
// It only updates the enabled flag and registers/unregisters with the scheduler.
func (s *TaskService) SetEnabled(ctx context.Context, taskID uuid.UUID, enabled bool) (*types.Task, error) {
	t, err := s.repo.FindTaskByID(taskID.String())
	if err != nil {
		return nil, fmt.Errorf("task not found")
	}
	t.Enabled = enabled
	if err := s.repo.UpdateTask(t); err != nil {
		return nil, err
	}
	s.registerToScheduler(t)
	t, _ = s.repo.FindTaskByID(taskID.String())
	return taskModelToType(t), nil
}

func (s *TaskService) Delete(ctx context.Context, taskID uuid.UUID) error {
	task, _ := s.repo.FindTaskByID(taskID.String())
	if task != nil {
		if err := s.repo.Tx(func(tx *repository.TaskRepo) error {
			if h := s.taskManager.Get(task.Kind); h != nil {
				if err := h.DeleteExtra(&_taskStorage{repo: tx, taskManager: s.taskManager}, task); err != nil {
					return err
				}
			}
			return tx.DeleteTask(task.ID)
		}); err != nil {
			return err
		}
	}
	scheduler.Unschedule(taskID.String())
	return nil
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

func taskModelToType(t *repository.TaskModel) *types.Task {
	return types.TaskFromModel(t)
}
