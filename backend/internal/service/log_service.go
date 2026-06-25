package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/dezhishen/now-and-again/shared/types"
)

// ─── Log ─────────────────────────────────────────────────────────

func (s *LogService) ListByTask(ctx context.Context, taskID uuid.UUID) ([]types.TaskLog, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *LogService) AddComment(ctx context.Context, taskID uuid.UUID, req *types.AddCommentRequest) (*types.TaskLog, error) {
	return nil, fmt.Errorf("not implemented")
}
