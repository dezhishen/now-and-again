package sdk

import (
	"context"
	"fmt"
	"strings"

	"github.com/dezhishen/now-and-again/backend/pkg/types"
	"github.com/google/uuid"
)

// ─── Todo operations ──────────────────────────────────────────────

// GetPendingTodos returns all pending todos in the active family.
func (na *NA) GetPendingTodos(ctx context.Context) ([]types.Todo, error) {
	fid, err := na.requireFamilyID()
	if err != nil {
		return nil, err
	}
	return na.Task.ListTodos(ctx, fid, "", "pending")
}

// ResolveTodoID accepts a full UUID or a short prefix (at least 3 chars)
// and returns the matching todo ID. Returns an error if not found or ambiguous.
func (na *NA) ResolveTodoID(ctx context.Context, input string) (string, error) {
	if _, err := uuid.Parse(input); err == nil {
		return input, nil // already a full UUID
	}
	if len(input) < 3 {
		return "", fmt.Errorf("id prefix too short (need at least 3 characters): %s", input)
	}
	todos, err := na.GetPendingTodos(ctx)
	if err != nil {
		return "", err
	}
	var matches []string
	inputLower := strings.ToLower(input)
	for _, t := range todos {
		if strings.HasPrefix(strings.ToLower(t.ID), inputLower) {
			matches = append(matches, t.ID)
		}
	}
	if len(matches) == 0 {
		return "", fmt.Errorf("no pending todo matches prefix: %s", input)
	}
	if len(matches) > 1 {
		return "", fmt.Errorf("ambiguous prefix %q matches %d todos — use a longer prefix", input, len(matches))
	}
	return matches[0], nil
}

// DoneTodo marks a todo as done — accepts full UUID or short prefix.
func (na *NA) DoneTodo(ctx context.Context, idOrPrefix string, remark string) (*types.Todo, error) {
	id, err := na.ResolveTodoID(ctx, idOrPrefix)
	if err != nil {
		return nil, err
	}
	return na.CompleteTodo(ctx, id, "done", remark)
}

// SkipTodo marks a todo as skipped — accepts full UUID or short prefix.
func (na *NA) SkipTodo(ctx context.Context, idOrPrefix string) (*types.Todo, error) {
	id, err := na.ResolveTodoID(ctx, idOrPrefix)
	if err != nil {
		return nil, err
	}
	return na.CompleteTodoSimple(ctx, id, "skipped")
}

// CompleteTodo marks a todo as done or skipped by full UUID.
func (na *NA) CompleteTodo(ctx context.Context, todoID string, status string, remark string) (*types.Todo, error) {
	id, err := uuid.Parse(todoID)
	if err != nil {
		return nil, fmt.Errorf("invalid todo id: %w", err)
	}
	req := &types.CompleteTodoRequest{
		Todo: &types.Todo{Status: status},
	}
	if remark != "" {
		req.Todo.Remark = remark
	}
	return na.Task.CompleteTodo(ctx, id, req)
}

// CompleteTodoSimple marks a todo without extra data.
func (na *NA) CompleteTodoSimple(ctx context.Context, todoID string, status string) (*types.Todo, error) {
	id, err := uuid.Parse(todoID)
	if err != nil {
		return nil, fmt.Errorf("invalid todo id: %w", err)
	}
	req := &types.CompleteTodoRequest{Todo: &types.Todo{Status: status}}
	return na.Task.CompleteTodo(ctx, id, req)
}

// ResolveTaskID finds a task by full UUID or short prefix.
func (na *NA) ResolveTaskID(ctx context.Context, input string) (*types.Task, error) {
	if _, err := uuid.Parse(input); err == nil {
		tasks, err := na.ListTasks(ctx)
		if err != nil {
			return nil, err
		}
		for _, t := range tasks {
			if t.ID == input {
				return &t, nil
			}
		}
		return nil, fmt.Errorf("task not found: %s", input)
	}
	if len(input) < 3 {
		return nil, fmt.Errorf("id prefix too short (need at least 3 characters): %s", input)
	}
	tasks, err := na.ListTasks(ctx)
	if err != nil {
		return nil, err
	}
	var matches []types.Task
	inputLower := strings.ToLower(input)
	for _, t := range tasks {
		if strings.HasPrefix(strings.ToLower(t.ID), inputLower) {
			matches = append(matches, t)
		}
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("no task matches prefix: %s", input)
	}
	if len(matches) > 1 {
		return nil, fmt.Errorf("ambiguous prefix %q matches %d tasks", input, len(matches))
	}
	return &matches[0], nil
}

// TriggerTaskByName triggers a task — accepts ID, prefix, or name.
func (na *NA) TriggerTaskByName(ctx context.Context, nameOrID string) error {
	// Try as task ID/prefix first
	task, err := na.ResolveTaskID(ctx, nameOrID)
	if err == nil {
		return na.Task.TriggerTask(ctx, uuid.MustParse(task.ID))
	}
	// Try by name
	t, err := na.FindTaskByName(ctx, nameOrID)
	if err != nil {
		return err
	}
	if t == nil {
		return fmt.Errorf("no task found matching: %s", nameOrID)
	}
	return na.Task.TriggerTask(ctx, uuid.MustParse(t.ID))
}

