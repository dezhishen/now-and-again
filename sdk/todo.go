package sdk

import (
	"context"
	"fmt"

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

// CompleteTodo marks a todo as done or skipped, with an optional remark.
func (na *NA) CompleteTodo(ctx context.Context, todoID string, status string, remark string) (*types.Todo, error) {
	id, err := uuid.Parse(todoID)
	if err != nil {
		return nil, fmt.Errorf("invalid todo id: %w", err)
	}
	req := &types.CompleteTodoRequest{
		Todo: &types.Todo{
			Status: status,
		},
	}
	if remark != "" {
		req.Todo.Remark = remark
	}
	return na.Task.CompleteTodo(ctx, id, req)
}

// CompleteTodoSimple marks a todo as done without extra data (for simple tasks).
func (na *NA) CompleteTodoSimple(ctx context.Context, todoID string, status string) (*types.Todo, error) {
	id, err := uuid.Parse(todoID)
	if err != nil {
		return nil, fmt.Errorf("invalid todo id: %w", err)
	}
	req := &types.CompleteTodoRequest{
		Todo: &types.Todo{Status: status},
	}
	return na.Task.CompleteTodo(ctx, id, req)
}

// SkipTodo marks a todo as skipped.
func (na *NA) SkipTodo(ctx context.Context, todoID string) (*types.Todo, error) {
	return na.CompleteTodoSimple(ctx, todoID, "skipped")
}

// DoneTodo marks a todo as done with a remark.
func (na *NA) DoneTodo(ctx context.Context, todoID string, remark string) (*types.Todo, error) {
	return na.CompleteTodo(ctx, todoID, "done", remark)
}
