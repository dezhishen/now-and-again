package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/dezhishen/now-and-again/shared/types"
)

// ─── Chain ───────────────────────────────────────────────────────

func (s *ChainService) Create(ctx context.Context, familyID uuid.UUID, req *types.CreateChainRequest) (*types.TaskChain, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *ChainService) List(ctx context.Context, familyID uuid.UUID) ([]types.TaskChain, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *ChainService) Get(ctx context.Context, chainID uuid.UUID) (*types.TaskChain, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *ChainService) AddStep(ctx context.Context, chainID uuid.UUID, req *types.AddStepRequest) (*types.TaskChainStep, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *ChainService) ReorderSteps(ctx context.Context, chainID uuid.UUID, req *types.ReorderStepsRequest) error {
	return fmt.Errorf("not implemented")
}

func (s *ChainService) RemoveStep(ctx context.Context, chainID, stepID uuid.UUID) error {
	return fmt.Errorf("not implemented")
}

func (s *ChainService) Start(ctx context.Context, chainID uuid.UUID) (*types.StartChainResponse, error) {
	return nil, fmt.Errorf("not implemented")
}
