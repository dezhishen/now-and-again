package scheduler

import (
	"fmt"
	"sync"
	"time"

	"github.com/go-co-op/gocron/v2"

	"github.com/dezhishen/now-and-again/backend/pkg/scheduler/engine"
	"github.com/dezhishen/now-and-again/backend/pkg/timeutil"
)

// ─── Package-level state ────────────────────────────────────────

var (
	logFunc   LogFunc
	mu        sync.Mutex
	scheduled = make(map[string]scheduledTask)
)

// scheduledTask is the internal record stored per scheduled task.
type scheduledTask struct {
	handler Handler
	onDone  func(taskID string)
}

// ─── Public types ───────────────────────────────────────────────

// LogFunc records scheduler events to an external system (e.g. DB).
type LogFunc func(taskID, status, message string)

// TaskInfo is the contract between service layer and scheduler.
// All type-specific lifecycle logic is handled internally by each Handler.
type TaskInfo struct {
	TaskID       string
	ScheduleType string
	ScheduleData string // raw JSON
	OnFire       func(taskID string, triggeredAt time.Time) error
	OnDone       func(taskID string) // called when a one-shot task finishes (fired or manually completed)
}

// ─── Lifecycle ──────────────────────────────────────────────────

// Init initializes the singleton engine. Idempotent — safe to call multiple times.
func Init(opts ...gocron.SchedulerOption) error {
	return engine.Init(opts...)
}

// SetLogger sets the logging callback used by all scheduling operations.
func SetLogger(lf LogFunc) {
	logFunc = lf
}

// Start begins the scheduler (non-blocking).
func Start() {
	engine.Get().Start()
	log("started", "", "")
}

// Stop gracefully shuts down the scheduler.
func Stop() error {
	return engine.Get().Stop()
}

// ─── Scheduling ─────────────────────────────────────────────────

// Schedule registers or replaces a job for the given task.
// All type-specific logic (job definition, task function, engine interaction)
// is delegated to the Handler.
func Schedule(t TaskInfo) error {
	mu.Lock()
	defer mu.Unlock()

	handler := defaultRegistry.Get(t.ScheduleType)
	if handler == nil {
		return fmt.Errorf("unknown schedule type: %s", t.ScheduleType)
	}

	// Remove existing job before re-adding.
	handler.Unschedule(t.TaskID)

	if err := handler.Schedule(t); err != nil {
		return err
	}

	scheduled[t.TaskID] = scheduledTask{
		handler: handler,
		onDone:  t.OnDone,
	}

	log("registered", t.TaskID, "schedule="+t.ScheduleType)
	return nil
}

// Unschedule removes a job by task ID.
func Unschedule(taskID string) {
	mu.Lock()
	defer mu.Unlock()
	rec, ok := scheduled[taskID]
	if ok {
		rec.handler.Unschedule(taskID)
		delete(scheduled, taskID)
	}
	log("removed", taskID, "")
}

// MarkCompleted signals that a task's todo was completed by the user.
// Delegates to the handler, which decides the appropriate action
// (one-shot handlers unschedule, recurring handlers ignore).
func MarkCompleted(taskID string) {
	mu.Lock()
	rec, ok := scheduled[taskID]
	mu.Unlock()
	if !ok {
		return
	}
	rec.handler.OnManualComplete(taskID, rec.onDone)
}

// ─── Shared helpers (used by handler implementations) ───────────

// log writes a scheduler event through the registered LogFunc.
func log(status, taskID, message string) {
	if logFunc != nil {
		logFunc(taskID, status, message)
	}
}

// defaultTaskFn returns a task function for recurring schedule types.
// It fires OnFire and logs the event; the job is kept alive by gocron.
func defaultTaskFn(t TaskInfo) func() {
	return func() {
		now := timeutil.Now()
		log("triggered", t.TaskID, "")
		if err := t.OnFire(t.TaskID, now); err != nil {
			log("error", t.TaskID, err.Error())
		}
	}
}
