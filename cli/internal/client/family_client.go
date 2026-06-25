package client

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/dezhishen/now-and-again/shared/types"
)

// ─── Family ──────────────────────────────────────────────────────

func (c *FamilyClient) Create(ctx context.Context, req *types.CreateFamilyRequest) (*types.Family, error) {
	return nil, fmt.Errorf("not implemented")
}
func (c *FamilyClient) Join(ctx context.Context, req *types.JoinFamilyRequest) (*types.FamilyMember, error) {
	return nil, fmt.Errorf("not implemented")
}
func (c *FamilyClient) Get(ctx context.Context, familyID uuid.UUID) (*types.Family, error) {
	return nil, fmt.Errorf("not implemented")
}
func (c *FamilyClient) ListMembers(ctx context.Context, familyID uuid.UUID) ([]types.FamilyMember, error) {
	return nil, fmt.Errorf("not implemented")
}
func (c *FamilyClient) UpdateMemberRole(ctx context.Context, familyID, userID uuid.UUID, role types.FamilyRole) error {
	return fmt.Errorf("not implemented")
}
func (c *FamilyClient) RemoveMember(ctx context.Context, familyID, userID uuid.UUID) error {
	return fmt.Errorf("not implemented")
}

func (c *FamilyClient) ListMyFamilies(ctx context.Context) ([]types.Family, error) {
	return nil, fmt.Errorf("not implemented")
}

func (c *FamilyClient) LeaveFamily(ctx context.Context, familyID uuid.UUID) error {
	return fmt.Errorf("not implemented")
}
