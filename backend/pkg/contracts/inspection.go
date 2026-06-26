package contracts

import (
	"context"

	"github.com/dezhishen/now-and-again/backend/pkg/types"
	"github.com/google/uuid"
)

// InspectionContract defines the inspection-specific API that both server
// and CLI must implement. It lives alongside the task-kind handler in
// backend/internal/taskkind/inspection.go — the two files form a cohesive
// plugin unit split only because shared/ is the sole package importable
// by both backend and CLI.
type InspectionContract interface {
	SubmitInspection(ctx context.Context, taskID uuid.UUID, req *types.SubmitInspectionRequest) (*types.Todo, error)
}
