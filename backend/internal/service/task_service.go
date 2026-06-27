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

	// Blank imports trigger init() registration of task kind handlers.
	_ "github.com/dezhishen/now-and-again/backend/pkg/taskkind/inspection"
	_ "github.com/dezhishen/now-and-again/backend/pkg/taskkind/simple"

	"github.com/dezhishen/now-and-again/backend/pkg/types"
)

type TaskService struct {
	repo      *repository.TaskRepo
	scheduler *scheduler.Scheduler
	Ops       *taskkind.Ops
}

func NewTaskService(repo *repository.TaskRepo, sched *scheduler.Scheduler) *TaskService {
	svc := &TaskService{
		repo:      repo,
		scheduler: sched,
		Ops:       &taskkind.Ops{Repo: repo, Scheduler: sched},
	}
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

// ─── TaskContract delegates ──────────────────────────────────────
// These implement shared/contracts.TaskContract so the service plugs into
// the compile-time-checked contract system used by both server and CLI.

func (s *TaskService) CreateTask(ctx context.Context, familyID uuid.UUID, req *types.CreateTaskRequest) (*types.Task, error) {
	return s.Create(ctx, familyID, req)
}
func (s *TaskService) UpdateTask(ctx context.Context, taskID uuid.UUID, req *types.UpdateTaskRequest) (*types.Task, error) {
	return s.Update(ctx, taskID, req)
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
		s.scheduler.RemoveJob(task.ID)
		return
	}

	b := &scheduler.JobBuilder{
		TaskID:       task.ID,
		ScheduleType: task.ScheduleType,
		ScheduleData: task.ScheduleData,
		Callback: func(taskID string, triggeredAt time.Time) error {
			return s.onTaskTriggered(taskID, task.FamilyID)
		},
	}

	// One-shot tasks auto-disable after firing (no recurrence).
	if task.ScheduleType == "once" {
		b.AfterFire = func(taskID string) {
			task.Enabled = false
			s.repo.UpdateTask(task)
		}
	}

	s.scheduler.RegisterJob(b)
}

func (s *TaskService) onTaskTriggered(taskID, familyID string) error {
	return s.createTodo(taskID, familyID, false)
}

func (s *TaskService) createTodo(taskID, familyID string, force bool) error {
	now := timeutil.Now()

	task, err := s.repo.FindTaskByID(taskID)
	if err != nil {
		return err
	}
	window := scheduleWindow(task.ScheduleType, task.ScheduleData)

	// Atomic check-and-create inside a transaction: prevents duplicate todos
	// when two requests (or scheduler + manual trigger) race.
	return s.repo.Tx(func(tx *repository.TaskRepo) error {
		if !force {
			has, _ := tx.HasPendingTodoForTaskToday(taskID, now)
			if has {
				return nil // already created — idempotent
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
		return tx.UpdateLastTodoAt(taskID, now)
	})
}

// ─── Task Template ───────────────────────────────────────────────

func (s *TaskService) GetTask(ctx context.Context, taskID uuid.UUID) (*types.Task, error) {
	t, err := s.repo.FindTaskByID(taskID.String())
	if err != nil {
		return nil, err
	}
	return taskModelToType(t), nil
}

func (s *TaskService) GetTaskWithExtra(ctx context.Context, taskID uuid.UUID) (map[string]any, error) {
	t, err := s.repo.FindTaskFull(taskID.String())
	if err != nil {
		return nil, err
	}
	result := map[string]any{
		"task": taskModelToType(t),
	}
	if h := taskkind.Get(t.Kind); h != nil {
		extra, _ := h.GetExtra(s.Ops, t)
		if extra != nil {
			result["extra"] = extra
		}
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
		FamilyID:     familyID.String(),
		GroupID:      req.Task.GroupID,
		LocationID:   req.Task.LocationID,
		IsRoot:       true,
		Name:         req.Task.Name,
		ScheduleType: req.Task.ScheduleType,
		ScheduleData: string(dataJSON),
		Enabled:      true,
		Kind:         kind,
		CreatedBy:    userID,
	}
	if err := s.repo.Tx(func(tx *repository.TaskRepo) error {
		if err := tx.CreateTask(t); err != nil {
			return err
		}
		if h := taskkind.Get(kind); h != nil {
			txOps := &taskkind.Ops{Repo: tx, Scheduler: s.scheduler}
			if err := h.OnCreate(txOps, t, req.Extra); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	s.registerToScheduler(t)
	s.onTaskTriggered(t.ID, t.FamilyID)
	t, _ = s.repo.FindTaskFull(t.ID)
	return taskModelToType(t), nil
}

func (s *TaskService) Trigger(ctx context.Context, taskID uuid.UUID) error {
	userID := ctx.Value("user_id").(string)
	task, err := s.repo.FindTaskByID(taskID.String())
	if err != nil {
		return err
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
	if req.Task != nil {
		f := req.Task
		if f.Name != nil {
			t.Name = *f.Name
		}
		if f.ScheduleType != nil {
			t.ScheduleType = *f.ScheduleType
		}
		if f.ScheduleData != nil {
			dataJSON, _ := json.Marshal(f.ScheduleData)
			t.ScheduleData = string(dataJSON)
		}
		if f.Enabled != nil {
			t.Enabled = *f.Enabled
		}
		if f.GroupID != nil {
			t.GroupID = *f.GroupID
		}
		if f.LocationID != nil {
			t.LocationID = *f.LocationID
		}
		if f.Kind != nil {
			t.Kind = *f.Kind
		}
	}
	if err := s.repo.Tx(func(tx *repository.TaskRepo) error {
		if err := tx.UpdateTask(t); err != nil {
			return err
		}
		if h := taskkind.Get(t.Kind); h != nil {
			txOps := &taskkind.Ops{Repo: tx, Scheduler: s.scheduler}
			if err := h.OnUpdate(txOps, t, req.Extra); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	s.registerToScheduler(t)
	t, _ = s.repo.FindTaskFull(taskID.String())
	return taskModelToType(t), nil
}

func (s *TaskService) Delete(ctx context.Context, taskID uuid.UUID) error {
	task, _ := s.repo.FindTaskByID(taskID.String())
	if task != nil {
		if err := s.repo.Tx(func(tx *repository.TaskRepo) error {
			if h := taskkind.Get(task.Kind); h != nil {
				txOps := &taskkind.Ops{Repo: tx, Scheduler: s.scheduler}
				if err := h.OnDelete(txOps, task); err != nil {
					return err
				}
			}
			return tx.DeleteTask(task.ID)
		}); err != nil {
			return err
		}
	}
	s.scheduler.RemoveJob(taskID.String())
	return nil
}

// ─── Calendar ──────────────────────────────────────────────────────

type CalendarDay struct {
	Date           string          `json:"date"`
	Weekday        int             `json:"weekday"`
	IsCurrentMonth bool            `json:"isCurrentMonth"`
	Events         []CalendarEvent `json:"events"`
}

type CalendarEvent struct {
	TaskID       string `json:"task_id"`
	Name         string `json:"name"`
	Kind         string `json:"kind"`
	Time         string `json:"time"`
	ScheduleType string `json:"schedule_type"`
	GroupName    string `json:"group_name,omitempty"`
}

func (s *TaskService) GetCalendar(ctx context.Context, familyID string, year, month int, groupID string) ([]CalendarDay, error) {
	tasks, err := s.repo.ListTasksByFamily(familyID)
	if err != nil {
		return nil, err
	}

	loc := time.UTC
	firstDay := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, loc)
	lastDay := firstDay.AddDate(0, 1, -1)

	// Build a map of day -> events
	dayEvents := make(map[int][]CalendarEvent)

	for _, task := range tasks {
		if !task.Enabled {
			continue
		}
		if groupID != "" && task.GroupID != "" && task.GroupID != groupID {
			continue
		}

		eventTime := parseEventTime(task.ScheduleData)
		days := expandSchedule(task.ScheduleType, task.ScheduleData, task.CreatedAt, year, month)
		for _, d := range days {
			dayEvents[d] = append(dayEvents[d], CalendarEvent{
				TaskID:       task.ID,
				Name:         task.Name,
				Kind:         task.Kind,
				Time:         eventTime,
				ScheduleType: task.ScheduleType,
				GroupName:    task.Group.Name,
			})
		}
	}

	// Build calendar grid: pad leading/trailing days from adjacent months
	var result []CalendarDay
	startWeekday := int(firstDay.Weekday()) // 0=Sun

	// Leading days from previous month (no events - just padding)
	prevMonth := firstDay.AddDate(0, 0, -1)
	prevLast := prevMonth.Day()
	for i := startWeekday - 1; i >= 0; i-- {
		d := prevLast - i
		date := time.Date(year, time.Month(month)-1, d, 0, 0, 0, 0, loc)
		result = append(result, CalendarDay{
			Date:           date.Format("2006-01-02"),
			Weekday:        int(date.Weekday()),
			IsCurrentMonth: false,
			Events:         []CalendarEvent{},
		})
	}

	// Current month days
	for d := 1; d <= lastDay.Day(); d++ {
		date := time.Date(year, time.Month(month), d, 0, 0, 0, 0, loc)
		events := dayEvents[d]
		if events == nil {
			events = []CalendarEvent{}
		}
		result = append(result, CalendarDay{
			Date:           date.Format("2006-01-02"),
			Weekday:        int(date.Weekday()),
			IsCurrentMonth: true,
			Events:         events,
		})
	}

	// Trailing days from next month
	remaining := 42 - len(result) // 6 rows × 7 cols
	for d := 1; d <= remaining; d++ {
		date := time.Date(year, time.Month(month)+1, d, 0, 0, 0, 0, loc)
		result = append(result, CalendarDay{
			Date:           date.Format("2006-01-02"),
			Weekday:        int(date.Weekday()),
			IsCurrentMonth: false,
			Events:         []CalendarEvent{},
		})
	}

	return result, nil
}

// expandSchedule returns the days of the month (1-31) that a task occurs on.
func expandSchedule(scheduleType, scheduleData string, createdAt time.Time, year, month int) []int {
	loc := time.UTC
	firstDay := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, loc)
	lastDay := firstDay.AddDate(0, 1, -1)

	switch scheduleType {
	case "daily":
		days := make([]int, lastDay.Day())
		for i := 1; i <= lastDay.Day(); i++ {
			days[i-1] = i
		}
		return days

	case "weekly":
		var data struct {
			Days []int `json:"days"`
		}
		json.Unmarshal([]byte(scheduleData), &data)
		daySet := make(map[int]bool)
		for _, wd := range data.Days {
			daySet[wd] = true // 1=Mon..7=Sun
		}
		if len(daySet) == 0 {
			// Default: every day of week (same as daily)
			days := make([]int, lastDay.Day())
			for i := 1; i <= lastDay.Day(); i++ {
				days[i-1] = i
			}
			return days
		}
		var result []int
		for d := 1; d <= lastDay.Day(); d++ {
			date := time.Date(year, time.Month(month), d, 0, 0, 0, 0, loc)
			wd := int(date.Weekday()) // 0=Sun
			iso := (wd+6)%7 + 1       // convert to 1=Mon..7=Sun
			if daySet[iso] {
				result = append(result, d)
			}
		}
		return result

	case "monthly":
		var data struct {
			Days []int `json:"days"`
		}
		json.Unmarshal([]byte(scheduleData), &data)
		if len(data.Days) == 0 {
			// Default: first day of month
			return []int{1}
		}
		var result []int
		for _, md := range data.Days {
			if md >= 1 && md <= lastDay.Day() {
				result = append(result, md)
			}
		}
		return result

	case "interval":
		var data struct {
			Days int `json:"days"`
		}
		json.Unmarshal([]byte(scheduleData), &data)
		interval := data.Days
		if interval <= 0 {
			interval = 1
		}
		// Walk forward from createdAt by interval, collect days in this month
		var result []int
		cursor := createdAt
		for cursor.Before(lastDay.AddDate(0, 0, 1)) {
			if cursor.Year() == year && int(cursor.Month()) == month {
				result = append(result, cursor.Day())
			}
			cursor = cursor.AddDate(0, 0, interval)
		}
		return result

	case "once":
		var data struct {
			Date string `json:"date"`
		}
		json.Unmarshal([]byte(scheduleData), &data)
		if data.Date == "" {
			return nil
		}
		parsed, err := time.ParseInLocation("2006-01-02", data.Date, time.UTC)
		if err != nil {
			return nil
		}
		if parsed.Year() == year && int(parsed.Month()) == month {
			return []int{parsed.Day()}
		}
		return nil
	}
	return nil
}

// parseEventTime extracts "HH:MM" from schedule data, defaulting to "09:00".
func parseEventTime(scheduleData string) string {
	var data struct {
		Time string `json:"time"`
	}
	json.Unmarshal([]byte(scheduleData), &data)
	if data.Time == "" {
		return "09:00"
	}
	return data.Time
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
