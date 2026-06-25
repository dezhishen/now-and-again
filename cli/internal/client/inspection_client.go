package client

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/dezhishen/now-and-again/shared/types"
)

// ─── Inspection ──────────────────────────────────────────────────

func (c *InspectionClient) Create(ctx context.Context, familyID uuid.UUID, req *types.CreateInspectionRequest) (*types.Inspection, error) {
	return nil, fmt.Errorf("not implemented")
}
func (c *InspectionClient) List(ctx context.Context, familyID uuid.UUID) ([]types.Inspection, error) {
	return nil, fmt.Errorf("not implemented")
}
func (c *InspectionClient) Get(ctx context.Context, inspectionID uuid.UUID) (*types.Inspection, error) {
	return nil, fmt.Errorf("not implemented")
}
func (c *InspectionClient) AddItem(ctx context.Context, inspectionID uuid.UUID, req *types.AddInspectionItemRequest) (*types.InspectionItem, error) {
	return nil, fmt.Errorf("not implemented")
}
func (c *InspectionClient) UpdateItem(ctx context.Context, inspectionID, itemID uuid.UUID, req *types.UpdateInspectionItemRequest) (*types.InspectionItem, error) {
	return nil, fmt.Errorf("not implemented")
}
func (c *InspectionClient) Complete(ctx context.Context, inspectionID uuid.UUID) (*types.Inspection, error) {
	return nil, fmt.Errorf("not implemented")
}
