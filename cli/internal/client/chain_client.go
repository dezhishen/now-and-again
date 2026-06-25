package client

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/dezhishen/now-and-again/shared/types"
)

// ─── Chain ───────────────────────────────────────────────────────

func (c *ChainClient) Create(ctx context.Context, familyID uuid.UUID, req *types.CreateChainRequest) (*types.TaskChain, error) {
	return nil, fmt.Errorf("not implemented")
}
func (c *ChainClient) List(ctx context.Context, familyID uuid.UUID) ([]types.TaskChain, error) {
	return nil, fmt.Errorf("not implemented")
}
func (c *ChainClient) Get(ctx context.Context, chainID uuid.UUID) (*types.TaskChain, error) {
	return nil, fmt.Errorf("not implemented")
}
func (c *ChainClient) AddStep(ctx context.Context, chainID uuid.UUID, req *types.AddStepRequest) (*types.TaskChainStep, error) {
	return nil, fmt.Errorf("not implemented")
}
func (c *ChainClient) ReorderSteps(ctx context.Context, chainID uuid.UUID, req *types.ReorderStepsRequest) error {
	return fmt.Errorf("not implemented")
}
func (c *ChainClient) RemoveStep(ctx context.Context, chainID, stepID uuid.UUID) error {
	return fmt.Errorf("not implemented")
}
func (c *ChainClient) Start(ctx context.Context, chainID uuid.UUID) (*types.StartChainResponse, error) {
	return nil, fmt.Errorf("not implemented")
}
