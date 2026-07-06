package sdk

import (
	"context"
	"fmt"

	"github.com/dezhishen/now-and-again/backend/pkg/types"
	"github.com/google/uuid"
)

// ─── Task CRUD (convenience wrappers with active family) ──────────

// CreateTask creates a task in the active family.
// User-supplied schedule_data times are assumed to be in the configured timezone
// and are automatically converted to UTC before sending.
func (na *NA) CreateTask(ctx context.Context, req *types.CreateTaskRequest) (*types.Task, error) {
	fid, err := na.requireFamilyID()
	if err != nil {
		return nil, err
	}
	// Convert schedule_data times from local→UTC before sending.
	req.Task.ScheduleData = scheduleDataToUTC(req.Task.ScheduleData, na.GetTimezone())
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

// ListTasksFiltered returns tasks in the active family with optional archived/disabled/name filters.
// Set includeArchived=true to include archived tasks.
// Set includeDisabled=true to include disabled tasks.
// Set name to a non-empty string to filter by task name (server-side LIKE).
func (na *NA) ListTasksFiltered(ctx context.Context, includeArchived, includeDisabled bool, name string) ([]types.Task, error) {
	fid, err := na.requireFamilyID()
	if err != nil {
		return nil, err
	}
	return na.Task.ListTasksFiltered(ctx, fid, includeArchived, includeDisabled, name)
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
// User-supplied schedule_data times are assumed to be in the configured timezone
// and are automatically converted to UTC before sending.
func (na *NA) UpdateTask(ctx context.Context, taskID string, req *types.UpdateTaskRequest) (*types.Task, error) {
	id, err := uuid.Parse(taskID)
	if err != nil {
		return nil, fmt.Errorf("invalid task id: %w", err)
	}
	// Convert schedule_data times from local→UTC before sending.
	if req.Task != nil {
		req.Task.ScheduleData = scheduleDataToUTC(req.Task.ScheduleData, na.GetTimezone())
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
// matches exactly, or contains the given substring. Uses server-side LIKE filtering.
// Returns nil if not found.
// Priority: exact match > substring match.
func (na *NA) FindTaskByName(ctx context.Context, name string) (*types.Task, error) {
	tasks, err := na.ListTasksFiltered(ctx, true, true, name)
	if err != nil {
		return nil, err
	}
	if len(tasks) == 0 {
		return nil, nil
	}
	// Exact match first
	for _, t := range tasks {
		if t.Name == name {
			return &t, nil
		}
	}
	// Then first substring match
	return &tasks[0], nil
}

// ─── Helpers ──────────────────────────────────────────────────────

func (na *NA) requireFamilyID() (uuid.UUID, error) {
	fid := na.ActiveFamilyID()
	if fid == "" {
		return uuid.Nil, fmt.Errorf("no active family; run 'na init' first or call SetActiveFamily")
	}
	return uuid.Parse(fid)
}
