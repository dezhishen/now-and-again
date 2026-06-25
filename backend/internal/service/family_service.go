package service

import (
	"context"
	"fmt"
	"time"

	"github.com/dezhishen/now-and-again/backend/internal/repository"
	"github.com/dezhishen/now-and-again/shared/types"
	"github.com/google/uuid"
)

// ─── FamilyService ────────────────────────────────────────────────

func (s *FamilyService) Create(ctx context.Context, req *types.CreateFamilyRequest) (*types.Family, error) {
	userID := userIDFromCtx(ctx)

	f := &repository.FamilyModel{
		Name:      req.Name,
		CreatedBy: userID,
	}
	if err := s.repo.Create(f); err != nil {
		return nil, fmt.Errorf("create family: %w", err)
	}
	// Creator is automatically owner
	_ = s.repo.AddMember(f.ID, userID, string(types.FamilyRoleOwner))

	return &types.Family{
		ID:         uuid.MustParse(f.ID),
		Name:       f.Name,
		InviteCode: f.InviteCode,
		CreatedBy:  uuid.MustParse(f.CreatedBy),
		Timestamps: types.Timestamps{CreatedAt: f.CreatedAt, UpdatedAt: f.UpdatedAt},
	}, nil
}

func (s *FamilyService) Join(ctx context.Context, req *types.JoinFamilyRequest) (*types.FamilyMember, error) {
	userID := userIDFromCtx(ctx)

	family, err := s.repo.FindByInviteCode(req.InviteCode)
	if err != nil || family == nil {
		return nil, fmt.Errorf("invalid invite code")
	}

	existing, _ := s.repo.FindMember(family.ID, userID)
	if existing != nil {
		return nil, fmt.Errorf("already a member")
	}

	if err := s.repo.AddMember(family.ID, userID, string(types.FamilyRoleMember)); err != nil {
		return nil, fmt.Errorf("join family: %w", err)
	}

	return &types.FamilyMember{
		FamilyID: uuid.MustParse(family.ID),
		UserID:   uuid.MustParse(userID),
		Role:     types.FamilyRoleMember,
		JoinedAt: time.Now().Format(time.RFC3339),
	}, nil
}

func (s *FamilyService) Get(ctx context.Context, familyID uuid.UUID) (*types.Family, error) {
	f, err := s.repo.FindByID(familyID.String())
	if err != nil || f == nil {
		return nil, fmt.Errorf("family not found")
	}
	return &types.Family{
		ID:         uuid.MustParse(f.ID),
		Name:       f.Name,
		InviteCode: f.InviteCode,
		CreatedBy:  uuid.MustParse(f.CreatedBy),
		Timestamps: types.Timestamps{CreatedAt: f.CreatedAt, UpdatedAt: f.UpdatedAt},
	}, nil
}

func (s *FamilyService) ListMyFamilies(ctx context.Context) ([]types.Family, error) {
	userID := userIDFromCtx(ctx)
	models, err := s.repo.ListFamiliesByUser(userID)
	if err != nil {
		return nil, err
	}
	families := make([]types.Family, len(models))
	for i, m := range models {
		families[i] = types.Family{
			ID:         uuid.MustParse(m.ID),
			Name:       m.Name,
			InviteCode: m.InviteCode,
			CreatedBy:  uuid.MustParse(m.CreatedBy),
			Timestamps: types.Timestamps{CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt},
		}
	}
	return families, nil
}

func (s *FamilyService) ListMembers(ctx context.Context, familyID uuid.UUID) ([]types.FamilyMember, error) {
	models, err := s.repo.ListMembers(familyID.String())
	if err != nil {
		return nil, err
	}
	members := make([]types.FamilyMember, len(models))
	for i, m := range models {
		members[i] = types.FamilyMember{
			ID:       uuid.MustParse(m.ID),
			FamilyID: uuid.MustParse(m.FamilyID),
			UserID:   uuid.MustParse(m.UserID),
			Role:     types.FamilyRole(m.Role),
			JoinedAt: m.JoinedAt.Format(time.RFC3339),
			User:     modelToUser(&m.User),
		}
	}
	return members, nil
}

func (s *FamilyService) UpdateMemberRole(ctx context.Context, familyID, userID uuid.UUID, role types.FamilyRole) error {
	return s.repo.UpdateMemberRole(familyID.String(), userID.String(), string(role))
}

func (s *FamilyService) RemoveMember(ctx context.Context, familyID, userID uuid.UUID) error {
	return s.repo.RemoveMember(familyID.String(), userID.String())
}

func (s *FamilyService) LeaveFamily(ctx context.Context, familyID uuid.UUID) error {
	userID := userIDFromCtx(ctx)
	return s.repo.RemoveMember(familyID.String(), userID)
}
