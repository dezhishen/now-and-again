package client

import (
	"context"

	"github.com/dezhishen/now-and-again/backend/pkg/types"
	"github.com/google/uuid"
)

// Compile-time check: TaskClient must satisfy the inspection contract.

// SubmitInspection implements contracts.InspectionContract.
func (c *TaskClient) SubmitInspection(_ context.Context, taskID uuid.UUID, req *types.SubmitInspectionRequest) (*types.Todo, error) {
	var t types.Todo
	if err := c.http.do("POST", "/api/tasks/"+taskID.String()+"/inspection", req, &t); err != nil {
		return nil, err
	}
	return &t, nil
}
