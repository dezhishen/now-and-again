package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/dezhishen/now-and-again/shared/types"
)

// ─── SubGroup ────────────────────────────────────────────────────

func (s *SubGroupService) Create(ctx context.Context, familyID uuid.UUID, req *types.CreateSubGroupRequest) (*types.SubGroup, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *SubGroupService) List(ctx context.Context, familyID uuid.UUID) ([]types.SubGroup, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *SubGroupService) AddMember(ctx context.Context, subGroupID uuid.UUID, req *types.AddSubGroupMemberRequest) (*types.SubGroupMember, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *SubGroupService) RemoveMember(ctx context.Context, subGroupID, userID uuid.UUID) error {
	return fmt.Errorf("not implemented")
}
