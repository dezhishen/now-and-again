package scheduler

import (
	"context"
	"log"
	"time"

	"github.com/dezhishen/now-and-again/backend/internal/notifier"
	"github.com/dezhishen/now-and-again/backend/internal/repository"
)

// Engine drives the scheduling lifecycle.
// It hooks into task creation/completion and periodically scans for reminders.
type Engine struct {
	taskRepo *repository.TaskRepo
	notif    *notifier.NotificationEngine
	stopCh   chan struct{}
	ticker   *time.Ticker
}

func NewEngine(taskRepo *repository.TaskRepo, notif *notifier.NotificationEngine) *Engine {
	return &Engine{taskRepo: taskRepo, notif: notif, stopCh: make(chan struct{})}
}

func (e *Engine) Start() {
	e.ticker = time.NewTicker(5 * time.Minute)
	log.Println("[scheduler] engine started (interval: 5m)")
	for {
		select {
		case <-e.ticker.C:
			e.tick(context.Background())
		case <-e.stopCh:
			e.ticker.Stop()
			log.Println("[scheduler] engine stopped")
			return
		}
	}
}

func (e *Engine) Stop() { close(e.stopCh) }

func (e *Engine) tick(ctx context.Context) {
	log.Println("[scheduler] tick")
	// TODO: scan tasks due soon → fire reminders
	// TODO: scan overdue tasks → fire overdue notifications
}

// OnTaskCompleted is called by the task service when a task is completed.
// It finds the schedule handler and applies the behavior.
func (e *Engine) OnTaskCompleted(ctx context.Context, task *repository.TaskModel) error {
	handler := Get(task.TaskCode)
	if handler == nil {
		return nil // no handler registered for this type
	}

	reset := handler.OnComplete(task)
	if reset != nil {
		// Auto-reset: create a new task instance
		if err := e.taskRepo.Create(reset); err != nil {
			return err
		}
		log.Printf("[scheduler] task %s auto-reset → new task %s", task.ID, reset.ID)
	}
	return nil
}
