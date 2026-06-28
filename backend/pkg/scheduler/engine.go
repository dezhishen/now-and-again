package scheduler

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/go-co-op/gocron/v2"

	"github.com/dezhishen/now-and-again/backend/pkg/timeutil"
)

// LogFunc records scheduler events to an external system (e.g. DB).
type LogFunc func(taskID, status, message string)

// TaskInfo is the contract between service layer and scheduler.
// The scheduler handles all type-specific logic (one-shot, recurring, etc.)
// internally. The service layer only provides task metadata and callbacks.
type TaskInfo struct {
	TaskID       string
	ScheduleType string
	ScheduleData string // raw JSON
	OnFire       func(taskID string, triggeredAt time.Time) error
	OnDone       func(taskID string) // called when a one-shot task finishes (fired or manually completed)
}

// scheduledTask is the internal record stored per scheduled task.
type scheduledTask struct {
	scheduleType string
	onDone       func(taskID string)
}

// Scheduler wraps gocron.Scheduler, providing closed-loop task scheduling.
// All schedule-type-specific logic (one-shot auto-remove, etc.) is handled
// internally. The service layer only calls Schedule/Unschedule/MarkCompleted.
type Scheduler struct {
	gs        gocron.Scheduler
	mu        sync.Mutex
	logFunc   LogFunc
	registry  *Registry                // handler registry (defaults to global)
	scheduled map[string]scheduledTask // taskID → internal record
}

// New creates a gocron-backed Scheduler using the default global handler registry.
func New(logFunc LogFunc, opts ...gocron.SchedulerOption) (*Scheduler, error) {
	return NewWithRegistry(logFunc, defaultRegistry, opts...)
}

// NewWithRegistry creates a Scheduler with a specific handler registry.
func NewWithRegistry(logFunc LogFunc, reg *Registry, opts ...gocron.SchedulerOption) (*Scheduler, error) {
	gs, err := gocron.NewScheduler(opts...)
	if err != nil {
		return nil, fmt.Errorf("create gocron: %w", err)
	}
	return &Scheduler{
		gs:        gs,
		logFunc:   logFunc,
		registry:  reg,
		scheduled: make(map[string]scheduledTask),
	}, nil
}

// Schedule registers or replaces a job for the given task.
// The scheduler internally handles one-shot auto-remove after firing.
func (s *Scheduler) Schedule(t TaskInfo) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Remove existing job for this task ID
	s.removeByTag(t.TaskID)

	handler := s.registry.Get(t.ScheduleType)
	if handler == nil {
		return fmt.Errorf("unknown schedule type: %s", t.ScheduleType)
	}

	var data map[string]interface{}
	json.Unmarshal([]byte(t.ScheduleData), &data)

	def := handler.BuildJob(data)
	if def == nil {
		return fmt.Errorf("handler %s returned nil def", t.ScheduleType)
	}

	isOneShot := handler.IsOneShot()

	// Build gocron task. The scheduler wraps the callback with one-shot logic.
	taskFn := gocron.NewTask(func() {
		now := timeutil.Now()
		s.log("triggered", t.TaskID, "")
		if err := t.OnFire(t.TaskID, now); err != nil {
			s.log("error", t.TaskID, err.Error())
			return
		}
		// One-shot tasks auto-unschedule after successful fire.
		if isOneShot {
			s.removeByTagLocked(t.TaskID)
			delete(s.scheduled, t.TaskID)
			if t.OnDone != nil {
				t.OnDone(t.TaskID)
			}
		}
	})

	_, err := s.gs.NewJob(def.toGocronDefinition(), taskFn, gocron.WithTags(t.TaskID))
	if err != nil {
		return fmt.Errorf("register job: %w", err)
	}

	// Store internal record for MarkCompleted lookup.
	s.scheduled[t.TaskID] = scheduledTask{
		scheduleType: t.ScheduleType,
		onDone:       t.OnDone,
	}

	s.log("registered", t.TaskID, "schedule="+t.ScheduleType)
	return nil
}

// Unschedule removes a job by task ID.
func (s *Scheduler) Unschedule(taskID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.removeByTagLocked(taskID)
	delete(s.scheduled, taskID)
	s.log("removed", taskID, "")
}

// MarkCompleted signals that a task's todo was completed by the user.
// For one-shot tasks this unschedules and calls OnDone.
// For recurring tasks this is a no-op.
func (s *Scheduler) MarkCompleted(taskID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rec, ok := s.scheduled[taskID]
	if !ok {
		return
	}

	// Only one-shot tasks are affected by manual completion.
	if h := s.registry.Get(rec.scheduleType); h == nil || !h.IsOneShot() {
		return
	}

	s.removeByTagLocked(taskID)
	delete(s.scheduled, taskID)
	s.log("completed", taskID, "")

	if rec.onDone != nil {
		rec.onDone(taskID)
	}
}

// Start begins the scheduler (non-blocking).
func (s *Scheduler) Start() {
	s.gs.Start()
	s.log("started", "", "")
}

// Stop gracefully shuts down the scheduler.
func (s *Scheduler) Stop() error {
	return s.gs.Shutdown()
}

// ─── Internal helpers ──────────────────────────────────────────

func (s *Scheduler) log(status, taskID, message string) {
	if s.logFunc != nil {
		s.logFunc(taskID, status, message)
	}
}

func (s *Scheduler) removeByTag(tag string) {
	// Lock must be held by caller.
	s.removeByTagLocked(tag)
}

func (s *Scheduler) removeByTagLocked(tag string) {
	for _, j := range s.gs.Jobs() {
		for _, t := range j.Tags() {
			if t == tag {
				s.gs.RemoveJob(j.ID())
				return
			}
		}
	}
}

// ─── Job Definition ────────────────────────────────────────────

// JobDef is returned by Handler.BuildJob. It wraps gocron job definitions.
type JobDef struct {
	duration  time.Duration
	cronExpr  string
	oneTimeAt *time.Time // non-nil for one-shot jobs
}

func (d *JobDef) toGocronDefinition() gocron.JobDefinition {
	if d.oneTimeAt != nil {
		return gocron.OneTimeJob(gocron.OneTimeJobStartDateTime(*d.oneTimeAt))
	}
	if d.duration > 0 {
		return gocron.DurationJob(d.duration)
	}
	if d.cronExpr != "" {
		return gocron.CronJob(d.cronExpr, false)
	}
	return gocron.DurationJob(time.Hour) // fallback
}

// DurationJobDef creates a duration-based job definition.
func DurationJobDef(d time.Duration) *JobDef {
	return &JobDef{duration: d}
}

// CronJobDef creates a cron-based job definition.
func CronJobDef(expr string) *JobDef {
	return &JobDef{cronExpr: expr}
}

// OneTimeJobDef creates a one-shot job definition at the given time.
func OneTimeJobDef(at time.Time) *JobDef {
	return &JobDef{oneTimeAt: &at}
}
