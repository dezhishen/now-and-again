package chain

import (
	"context"
	"log"
)

// Engine handles the lifecycle of a task chain:
// - Instantiation: when StartChain is called, create all Tasks + Dependencies.
// - Progression: when a task is completed, unblock its successors (respecting delays).
//
// This is the core "事项链" engine. It works with the service layer,
// not directly with repositories, to ensure proper logging and notifications.

type Engine struct {
	// Injected dependencies (interfaces to be defined)
}

func New() *Engine {
	return &Engine{}
}

// StartChain instantiates all steps of a chain as actual Tasks.
// Returns the created tasks with dependencies wired up.
func (e *Engine) StartChain(ctx context.Context, chainID string, familyID string, createdBy string) ([]string, error) {
	// TODO:
	// 1. Load chain + steps
	// 2. For each step, create a Task (status = blocked except step[0])
	// 3. For step[i] and step[i-1], create TaskDependency (blocks)
	// 4. Fire notifications
	log.Printf("[chain] starting chain=%s family=%s by=%s", chainID, familyID, createdBy)
	return nil, nil
}

// OnTaskCompleted is called by the task service when a task is marked done.
// If the task is part of a chain, unblock the next step(s).
func (e *Engine) OnTaskCompleted(ctx context.Context, taskID string) error {
	// TODO:
	// 1. Find TaskDependency WHERE blocker_task_id = taskID
	// 2. For each blocked task, check if ALL blockers are done
	// 3. If yes, set status from blocked → todo (respect delay_after_previous)
	// 4. Fire task_unblocked notification
	log.Printf("[chain] task completed: %s", taskID)
	return nil
}
