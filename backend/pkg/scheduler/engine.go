package scheduler

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/go-co-op/gocron/v2"

	"github.com/dezhishen/now-and-again/backend/internal/logger"
)

// Callback is invoked when a scheduled task triggers.
type Callback func(taskID string, triggeredAt time.Time) error

// LogFunc records scheduler events to an external system (e.g. DB).
type LogFunc func(taskID, status, message string)

// JobBuilder packages parameters for building a gocron job,
// hiding gocron details from business code.
type JobBuilder struct {
	TaskID       string
	ScheduleType string
	ScheduleData string // raw JSON
	Callback     Callback
	AfterFire    func(taskID string) // optional, called after Callback succeeds
}

// Scheduler wraps gocron.Scheduler, providing Register/Remove by task ID
// and bridging handler definitions to gocron jobs.
type Scheduler struct {
	gs       gocron.Scheduler
	mu       sync.Mutex
	logFunc  LogFunc
	handlers map[string]Handler // populated via RegisterHandler
}

// New creates a gocron-backed Scheduler. opts are passed directly to gocron.
func New(logFunc LogFunc, opts ...gocron.SchedulerOption) (*Scheduler, error) {
	gs, err := gocron.NewScheduler(opts...)
	if err != nil {
		return nil, fmt.Errorf("create gocron: %w", err)
	}
	s := &Scheduler{
		gs:       gs,
		logFunc:  logFunc,
		handlers: make(map[string]Handler),
	}
	// Register built-in handlers from the global registry
	for _, h := range AllHandlers() {
		s.handlers[h.Code()] = h
	}
	return s, nil
}

// RegisterJob builds a gocron job from the JobBuilder and registers it.
// If a job with the same task_id tag already exists, it is replaced.
func (s *Scheduler) RegisterJob(b *JobBuilder) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Remove existing job tagged with this task_id
	s.removeByTag(b.TaskID)

	handler, ok := s.handlers[b.ScheduleType]
	if !ok {
		return fmt.Errorf("unknown schedule type: %s", b.ScheduleType)
	}

	var data map[string]interface{}
	json.Unmarshal([]byte(b.ScheduleData), &data)

	def := handler.BuildJob(data)
	if def == nil {
		return fmt.Errorf("handler %s returned nil def", b.ScheduleType)
	}

	// Build gocron task with callback
	taskFn := gocron.NewTask(func() {
		now := time.Now()
		s.log("triggered", b.TaskID, "")
		if err := b.Callback(b.TaskID, now); err != nil {
			s.log("error", b.TaskID, err.Error())
			return
		}
		if b.AfterFire != nil {
			b.AfterFire(b.TaskID)
		}
	})

	_, err := s.gs.NewJob(def.toGocronDefinition(), taskFn, gocron.WithTags(b.TaskID))
	if err != nil {
		return fmt.Errorf("register job: %w", err)
	}

	s.log("registered", b.TaskID, "schedule="+b.ScheduleType)
	return nil
}

// RemoveJob removes a job by task_id tag.
func (s *Scheduler) RemoveJob(taskID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.removeByTag(taskID)
	s.log("removed", taskID, "")
}

func (s *Scheduler) removeByTag(tag string) {
	for _, j := range s.gs.Jobs() {
		for _, t := range j.Tags() {
			if t == tag {
				s.gs.RemoveJob(j.ID())
				return
			}
		}
	}
}

// Start begins the scheduler (non-blocking).
func (s *Scheduler) Start() {
	s.gs.Start()
	logger.Infof("scheduler started")
}

// Stop gracefully shuts down the scheduler.
func (s *Scheduler) Stop() error {
	return s.gs.Shutdown()
}

func (s *Scheduler) log(status, taskID, message string) {
	if s.logFunc != nil {
		s.logFunc(taskID, status, message)
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
