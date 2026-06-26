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

func (s *TaskService) CreateTask(ctx context.Context, familyID uuid.UUID, req *types.CreateTaskRequest) (*types.TaskTemplate, error) {
	return s.Create(ctx, familyID, req)
}
func (s *TaskService) UpdateTask(ctx context.Context, taskID uuid.UUID, req *types.UpdateTaskRequest) (*types.TaskTemplate, error) {
	return s.Update(ctx, taskID, req)
}
func (s *TaskService) DeleteTask(ctx context.Context, taskID uuid.UUID) error {
	return s.Delete(ctx, taskID)
}
func (s *TaskService) ListTasks(ctx context.Context, familyID uuid.UUID) ([]types.TaskTemplate, error) {
	return s.List(ctx, familyID)
}
func (s *TaskService) TriggerTask(ctx context.Context, taskID uuid.UUID) error {
	return s.Trigger(ctx, taskID)
}
func (s *TaskService) ListTaskLogs(ctx context.Context, taskID uuid.UUID, limit int, userOnly bool) ([]types.TaskLog, error) {
	return s.ListLogs(ctx, taskID, limit, userOnly)
}

func (s *TaskService) registerToScheduler(task *repository.TaskTemplateModel) {
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
		CheckItems:   marshalJSON(req.CheckItems),
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

	// Avoid duplicate: don't create another todo if one is already pending today.
	if has, _ := s.repo.HasPendingTodoForTaskToday(taskID.String(), time.Now()); has {
		return fmt.Errorf("今天已有待办，无需重复生成")
	}

	s.repo.CreateUserLog(taskID.String(), "", userID, "manual", fmt.Sprintf("手动生成待办: %s", task.Name))
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
	if req.CheckItems != nil {
		t.CheckItems = marshalJSON(req.CheckItems)
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
			ID: l.ID, TaskID: l.TaskID, TodoID: l.TodoID, Status: l.Status,
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

	// Determine completion status
	status := req.Status
	branchName := req.BranchName

	// Always mark the todo done
	if err := s.repo.CompleteTodo(todoID.String(), userID, status, req.Remark); err != nil {
		return nil, err
	}

	// Always log the base completion (kind handlers may add extra logs).
	msg := fmt.Sprintf("完成待办: %s", todo.Task.Name)
	if req.Remark != "" {
		msg += fmt.Sprintf(" | 备注: %s", req.Remark)
	}
	s.repo.CreateUserLog(todo.TaskID, todoID.String(), userID, status, msg)

	// Dispatch kind-specific completion logic
	if h := taskkind.Get(todo.Task.Kind); h != nil {
		if len(req.Selections) > 0 {
			if insp, ok := h.(taskkind.Inspector); ok {
				selections := make([]taskkind.Selection, len(req.Selections))
				for i, s := range req.Selections {
					selections[i] = taskkind.Selection{Item: s.Item, Branch: s.Branch}
				}
				insp.OnInspect(s.Ops, todo, selections, userID)
			}
		} else {
			h.OnComplete(s.Ops, todo, branchName, userID)
		}
	}

	// One-shot tasks are done after first completion — disable and stop scheduling.
	s.disableCompletedOnceTask(todo)

	t, err := s.repo.FindTodoByID(todoID.String())
	if err != nil {
		return nil, err
	}
	return todoModelToType(t), nil
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

	loc := time.Local
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
	loc := time.Local
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
		parsed, err := time.ParseInLocation("2006-01-02", data.Date, loc)
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

// ─── Statistics ──────────────────────────────────────────────────

type StatsPeriod string

const (
	StatsWeek  StatsPeriod = "week"
	StatsMonth StatsPeriod = "month"
	StatsYear  StatsPeriod = "year"
)

type StatsResponse struct {
	Period    StatsPeriod   `json:"period"`
	StartDate string        `json:"start_date"`
	EndDate   string        `json:"end_date"`
	Summary   StatsSummary  `json:"summary"`
	Daily     []StatsDaily  `json:"daily"`
	ByTask    []StatsByTask `json:"by_task"`
}

type StatsSummary struct {
	TotalCompleted int     `json:"total_completed"`
	TotalSkipped   int     `json:"total_skipped"`
	TotalManual    int     `json:"total_manual"`
	TotalTasks     int     `json:"total_tasks"`
	CompletionRate float64 `json:"completion_rate"` // 0-1
}

type StatsDaily struct {
	Date      string `json:"date"`
	Completed int    `json:"completed"`
	Skipped   int    `json:"skipped"`
	Manual    int    `json:"manual"`
}

type StatsByTask struct {
	TaskID    string `json:"task_id"`
	TaskName  string `json:"task_name"`
	Completed int    `json:"completed"`
	Skipped   int    `json:"skipped"`
	Manual    int    `json:"manual"`
}

func (s *TaskService) GetStatistics(ctx context.Context, familyID string, period StatsPeriod, refDate string) (*StatsResponse, error) {
	loc := time.Local
	now := time.Now()

	// Parse reference date or use now
	var ref time.Time
	if refDate != "" {
		var err error
		ref, err = time.ParseInLocation("2006-01-02", refDate, loc)
		if err != nil {
			ref = now
		}
	} else {
		ref = now
	}

	// Compute date range
	var startDate, endDate time.Time
	switch period {
	case StatsWeek:
		// Monday of the week containing ref
		wd := ref.Weekday()
		if wd == time.Sunday {
			wd = 7
		}
		startDate = time.Date(ref.Year(), ref.Month(), ref.Day()-int(wd)+1, 0, 0, 0, 0, loc)
		endDate = startDate.AddDate(0, 0, 7)
	case StatsMonth:
		startDate = time.Date(ref.Year(), ref.Month(), 1, 0, 0, 0, 0, loc)
		endDate = startDate.AddDate(0, 1, 0)
	case StatsYear:
		startDate = time.Date(ref.Year(), 1, 1, 0, 0, 0, 0, loc)
		endDate = startDate.AddDate(1, 0, 0)
	default:
		startDate = time.Date(ref.Year(), ref.Month(), 1, 0, 0, 0, 0, loc)
		endDate = startDate.AddDate(0, 1, 0)
	}

	since := startDate.Format("2006-01-02")
	until := endDate.Format("2006-01-02")

	logs, err := s.repo.ListLogsByFamily(familyID, since, until)
	if err != nil {
		return nil, err
	}

	// Get tasks for name lookup
	tasks, _ := s.repo.ListTasksByFamily(familyID)
	taskNames := make(map[string]string)
	for _, t := range tasks {
		taskNames[t.ID] = t.Name
	}

	// Aggregate
	dailyMap := make(map[string]*StatsDaily)
	byTaskMap := make(map[string]*StatsByTask)
	var summary StatsSummary

	// Initialize daily buckets
	cursor := startDate
	for cursor.Before(endDate) {
		key := cursor.Format("2006-01-02")
		dailyMap[key] = &StatsDaily{Date: key}
		cursor = cursor.AddDate(0, 0, 1)
	}

	for _, log := range logs {
		dateKey := log.CreatedAt.Format("2006-01-02")

		// Ensure daily bucket exists
		if dailyMap[dateKey] == nil {
			dailyMap[dateKey] = &StatsDaily{Date: dateKey}
		}

		// Ensure task bucket exists
		if byTaskMap[log.TaskID] == nil {
			byTaskMap[log.TaskID] = &StatsByTask{
				TaskID:   log.TaskID,
				TaskName: taskNames[log.TaskID],
			}
		}

		switch log.Status {
		case "done", "completed":
			dailyMap[dateKey].Completed++
			byTaskMap[log.TaskID].Completed++
			summary.TotalCompleted++
		case "skipped":
			dailyMap[dateKey].Skipped++
			byTaskMap[log.TaskID].Skipped++
			summary.TotalSkipped++
		case "manual":
			dailyMap[dateKey].Manual++
			byTaskMap[log.TaskID].Manual++
			summary.TotalManual++
		}
	}

	summary.TotalTasks = summary.TotalCompleted + summary.TotalSkipped
	if summary.TotalTasks > 0 {
		summary.CompletionRate = float64(summary.TotalCompleted) / float64(summary.TotalTasks)
	}

	// Convert maps to sorted slices
	daily := make([]StatsDaily, 0, len(dailyMap))
	cursor = startDate
	for cursor.Before(endDate) {
		key := cursor.Format("2006-01-02")
		if d, ok := dailyMap[key]; ok {
			daily = append(daily, *d)
		}
		cursor = cursor.AddDate(0, 0, 1)
	}

	byTask := make([]StatsByTask, 0, len(byTaskMap))
	for _, bt := range byTaskMap {
		byTask = append(byTask, *bt)
	}

	return &StatsResponse{
		Period:    period,
		StartDate: since,
		EndDate:   until,
		Summary:   summary,
		Daily:     daily,
		ByTask:    byTask,
	}, nil
}

// disableCompletedOnceTask handles the one-shot task lifecycle: once a once-type
// task's todo is completed, the task is disabled and unscheduled.
func (s *TaskService) disableCompletedOnceTask(todo *repository.TodoModel) {
	if todo.Task.ScheduleType != "once" {
		return
	}
	s.repo.DisableTask(todo.TaskID)
	s.scheduler.RemoveJob(todo.TaskID)
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
	var checkItems any
	if t.CheckItems != "" {
		json.Unmarshal([]byte(t.CheckItems), &checkItems)
	}
	return &types.TaskTemplate{
		ID: t.ID, FamilyID: t.FamilyID, GroupID: t.GroupID,
		LocationID: t.LocationID,
		Name:       t.Name, ScheduleType: t.ScheduleType, ScheduleData: data,
		Enabled: t.Enabled, Kind: t.Kind, CheckItems: checkItems,
		LastTodoAt: t.LastTodoAt,
		CreatedBy:  t.CreatedBy, CreatedAt: t.CreatedAt, UpdatedAt: t.UpdatedAt,
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
		AssignedTo: t.AssignedTo, Status: t.Status, BranchName: t.BranchName, Remark: t.Remark,
		DueStart:    t.DueStart,
		DueDate:     t.DueDate,
		CompletedAt: t.CompletedAt, CompletedBy: t.CompletedBy,
		Task: task, User: user,
		CreatedAt: t.CreatedAt, UpdatedAt: t.UpdatedAt,
	}
}
