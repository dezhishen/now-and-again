package client

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/dezhishen/now-and-again/shared/types"
)

// ─── Log ─────────────────────────────────────────────────────────

func (c *LogClient) ListByTask(ctx context.Context, taskID uuid.UUID) ([]types.TaskLog, error) {
	return nil, fmt.Errorf("not implemented")
}
func (c *LogClient) AddComment(ctx context.Context, taskID uuid.UUID, req *types.AddCommentRequest) (*types.TaskLog, error) {
	return nil, fmt.Errorf("not implemented")
}
