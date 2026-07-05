package sdk

import (
	"context"
	"fmt"

	"github.com/dezhishen/now-and-again/backend/pkg/types"
	"github.com/google/uuid"
)

// ─── Task CRUD (convenience wrappers with active family) ──────────

// CreateTask creates a task in the active family.
func (na *NA) CreateTask(ctx context.Context, req *types.CreateTaskRequest) (*types.Task, error) {
	fid, err := na.requireFamilyID()
	if err != nil {
		return nil, err
	}
	return na.Task.CreateTask(ctx, fid, req)
}

// ListTasks returns all tasks in the active family.
func (na *NA) ListTasks(ctx context.Context) ([]types.Task, error) {
	fid, err := na.requireFamilyID()
	if err != nil {
		return nil, err
	}
	return na.Task.ListTasks(ctx, fid)
}

// GetTask returns a single task by ID.
func (na *NA) GetTask(ctx context.Context, taskID string) (*types.Task, error) {
	id, err := uuid.Parse(taskID)
	if err != nil {
		return nil, fmt.Errorf("invalid task id: %w", err)
	}
	// Direct HTTP call since TaskClient doesn't have GetTask yet.
	var t types.Task
	path := fmt.Sprintf("/api/tasks/%s", id)
	if err := na.http.Do("GET", path, nil, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

// UpdateTask updates a task.
func (na *NA) UpdateTask(ctx context.Context, taskID string, req *types.UpdateTaskRequest) (*types.Task, error) {
	id, err := uuid.Parse(taskID)
	if err != nil {
		return nil, fmt.Errorf("invalid task id: %w", err)
	}
	return na.Task.UpdateTask(ctx, id, req)
}

// DeleteTask deletes a task.
func (na *NA) DeleteTask(ctx context.Context, taskID string) error {
	id, err := uuid.Parse(taskID)
	if err != nil {
		return fmt.Errorf("invalid task id: %w", err)
	}
	return na.Task.DeleteTask(ctx, id)
}

// TriggerTask immediately generates a todo for the given task.
func (na *NA) TriggerTask(ctx context.Context, taskID string) error {
	id, err := uuid.Parse(taskID)
	if err != nil {
		return fmt.Errorf("invalid task id: %w", err)
	}
	return na.Task.TriggerTask(ctx, id)
}

// FindTaskByName returns the first task in the active family whose name
// contains the given substring. Returns nil if not found.
func (na *NA) FindTaskByName(ctx context.Context, name string) (*types.Task, error) {
	tasks, err := na.ListTasks(ctx)
	if err != nil {
		return nil, err
	}
	for _, t := range tasks {
		if t.Name == name || containsSubstring(t.Name, name) {
			return &t, nil
		}
	}
	return nil, nil
}

// ─── Helpers ──────────────────────────────────────────────────────

func (na *NA) requireFamilyID() (uuid.UUID, error) {
	fid := na.ActiveFamilyID()
	if fid == "" {
		return uuid.Nil, fmt.Errorf("no active family; run 'na init' first or call SetActiveFamily")
	}
	return uuid.Parse(fid)
}

func containsSubstring(s, sub string) bool {
	return len(s) >= len(sub) && searchString(s, sub)
}

func searchString(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
