package client

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/dezhishen/now-and-again/shared/types"
)

// ─── Task ────────────────────────────────────────────────────────

func (c *TaskClient) Create(ctx context.Context, req *types.CreateTaskRequest) (*types.Task, error) {
	return nil, fmt.Errorf("not implemented")
}
func (c *TaskClient) List(ctx context.Context, familyID uuid.UUID, status *types.TaskStatus, assigneeID *uuid.UUID, page, pageSize int) ([]types.Task, int, error) {
	return nil, 0, fmt.Errorf("not implemented")
}
func (c *TaskClient) Get(ctx context.Context, taskID uuid.UUID) (*types.Task, error) {
	return nil, fmt.Errorf("not implemented")
}
func (c *TaskClient) Update(ctx context.Context, taskID uuid.UUID, req *types.UpdateTaskRequest) (*types.Task, error) {
	return nil, fmt.Errorf("not implemented")
}
func (c *TaskClient) SetAssignees(ctx context.Context, taskID uuid.UUID, req *types.SetAssigneesRequest) ([]types.TaskAssignee, error) {
	return nil, fmt.Errorf("not implemented")
}
func (c *TaskClient) AddDependency(ctx context.Context, taskID uuid.UUID, req *types.CreateDependencyRequest) (*types.TaskDependency, error) {
	return nil, fmt.Errorf("not implemented")
}
func (c *TaskClient) RemoveDependency(ctx context.Context, taskID, depID uuid.UUID) error {
	return fmt.Errorf("not implemented")
}
