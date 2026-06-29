// Package engine provides a singleton wrapper around gocron.Scheduler.
// It is an internal package of scheduler and must not be imported directly
// by packages outside pkg/scheduler/.
package engine

import (
	"fmt"
	"sync"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
)

// ─── Singleton ──────────────────────────────────────────────────

var (
	instance *Engine
	initMu   sync.Mutex
)

// Init initializes the singleton Engine. Safe to call multiple times;
// subsequent calls are no-ops. Must be called before Get.
func Init(opts ...gocron.SchedulerOption) error {
	initMu.Lock()
	defer initMu.Unlock()
	if instance != nil {
		return nil
	}
	gs, err := gocron.NewScheduler(opts...)
	if err != nil {
		return fmt.Errorf("engine: create gocron: %w", err)
	}
	instance = &Engine{gs: gs}
	return nil
}

// Get returns the singleton Engine. Panics if Init has not been called.
func Get() *Engine {
	if instance == nil {
		panic("engine: not initialized – call Init first")
	}
	return instance
}

// ─── Engine ─────────────────────────────────────────────────────

// Engine wraps gocron.Scheduler behind a minimal, singleton-safe API.
type Engine struct {
	gs gocron.Scheduler
	mu sync.Mutex
}

// Start begins the scheduler (non-blocking).
func (e *Engine) Start() { e.gs.Start() }

// Stop gracefully shuts down the scheduler.
func (e *Engine) Stop() error { return e.gs.Shutdown() }

// AddJob adds a job identified by taskID. The taskID string is parsed as a UUID
// and passed to gocron via WithIdentifier, so RemoveJob can remove it in O(1).
func (e *Engine) AddJob(taskID string, def gocron.JobDefinition, taskFn gocron.Task) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	id, err := uuid.Parse(taskID)
	if err != nil {
		return fmt.Errorf("engine: invalid taskID %q: %w", taskID, err)
	}

	_, err = e.gs.NewJob(def, taskFn, gocron.WithIdentifier(id))
	return err
}

// RemoveJob removes a job by its taskID in O(1) via gocron's UUID-based removal.
func (e *Engine) RemoveJob(taskID string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	id, err := uuid.Parse(taskID)
	if err != nil {
		return
	}
	_ = e.gs.RemoveJob(id)
}

// ─── Job definition helpers ─────────────────────────────────────
// Thin wrappers that return gocron.JobDefinition directly.

// DurationJobDef creates a duration-based job definition.
func DurationJobDef(d time.Duration) gocron.JobDefinition {
	return gocron.DurationJob(d)
}

// CronJobDef creates a cron-based job definition.
func CronJobDef(expr string) gocron.JobDefinition {
	return gocron.CronJob(expr, false)
}

// OneTimeJobDef creates a one-shot job definition at the given time.
func OneTimeJobDef(at time.Time) gocron.JobDefinition {
	return gocron.OneTimeJob(gocron.OneTimeJobStartDateTime(at))
}
