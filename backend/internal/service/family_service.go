package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/dezhishen/now-and-again/backend/internal/repository"
	"github.com/dezhishen/now-and-again/shared/types"
)

// ─── Family CRUD ──────────────────────────────────────────────────

func (s *FamilyService) Create(ctx context.Context, req *types.CreateFamilyRequest) (*types.Family, error) {
	userID := ctx.Value("user_id").(string)

	f := &repository.FamilyModel{
		Name:       req.Name,
		InviteCode: repository.GenInviteCode(),
		CreatedBy:  userID,
	}
	if err := s.repo.CreateFamily(f); err != nil {
		return nil, fmt.Errorf("create family: %w", err)
	}

	m := &repository.FamilyMemberModel{
		FamilyID: f.ID,
		UserID:   userID,
		Role:     string(types.FamilyRoleOwner),
		Status:   "active",
		JoinedAt: time.Now(),
	}
	if err := s.repo.AddMember(m); err != nil {
		return nil, fmt.Errorf("add creator: %w", err)
	}

	return &types.Family{ID: f.ID, Name: f.Name, InviteCode: f.InviteCode, CreatedBy: f.CreatedBy, CreatedAt: f.CreatedAt, UpdatedAt: f.UpdatedAt}, nil
}

func (s *FamilyService) Join(ctx context.Context, req *types.JoinFamilyRequest) (*types.FamilyMember, error) {
	userID := ctx.Value("user_id").(string)

	f, err := s.repo.FindFamilyByInviteCode(req.InviteCode)
	if err != nil {
		return nil, fmt.Errorf("invalid invite code")
	}

	if existing, _ := s.repo.FindMember(f.ID, userID); existing != nil {
		if existing.Status == "active" {
			return nil, fmt.Errorf("already a member")
		}
		return nil, fmt.Errorf("join request already exists")
	}

	m := &repository.FamilyMemberModel{
		FamilyID: f.ID, UserID: userID, Role: string(types.FamilyRoleMember),
		Status: "pending", JoinedAt: time.Now(),
	}
	if err := s.repo.AddMember(m); err != nil {
		return nil, fmt.Errorf("join family: %w", err)
	}

	return &types.FamilyMember{ID: m.ID, FamilyID: m.FamilyID, UserID: m.UserID, Role: types.FamilyRole(m.Role), Status: types.MemberStatus(m.Status), JoinedAt: m.JoinedAt}, nil
}

func (s *FamilyService) Get(ctx context.Context, familyID uuid.UUID) (*types.Family, error) {
	f, err := s.repo.FindFamilyByID(familyID.String())
	if err != nil {
		return nil, fmt.Errorf("family not found")
	}
	return &types.Family{ID: f.ID, Name: f.Name, InviteCode: f.InviteCode, CreatedBy: f.CreatedBy, CreatedAt: f.CreatedAt, UpdatedAt: f.UpdatedAt}, nil
}

func (s *FamilyService) ListMyFamilies(ctx context.Context) ([]types.Family, error) {
	userID := ctx.Value("user_id").(string)
	families, err := s.repo.ListFamiliesByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("list families: %w", err)
	}
	result := make([]types.Family, len(families))
	for i, f := range families {
		result[i] = types.Family{ID: f.ID, Name: f.Name, InviteCode: f.InviteCode, CreatedBy: f.CreatedBy, CreatedAt: f.CreatedAt, UpdatedAt: f.UpdatedAt}
	}
	return result, nil
}

// ─── Members ──────────────────────────────────────────────────────

func toFamilyMember(m *repository.FamilyMemberModel) *types.FamilyMember {
	var u *types.User
	if m.User.ID != "" {
		u = userModelToUser(&m.User)
	}
	return &types.FamilyMember{ID: m.ID, FamilyID: m.FamilyID, UserID: m.UserID, Role: types.FamilyRole(m.Role), Status: types.MemberStatus(m.Status), JoinedAt: m.JoinedAt, User: u}
}

func (s *FamilyService) ListMembers(ctx context.Context, familyID uuid.UUID) ([]types.FamilyMember, error) {
	members, err := s.repo.ListMembers(familyID.String())
	if err != nil {
		return nil, err
	}
	result := make([]types.FamilyMember, len(members))
	for i, m := range members {
		result[i] = *toFamilyMember(&m)
	}
	return result, nil
}

func (s *FamilyService) UpdateMemberRole(ctx context.Context, familyID, userID uuid.UUID, role types.FamilyRole) error {
	m, err := s.repo.FindMember(familyID.String(), userID.String())
	if err != nil {
		return fmt.Errorf("member not found")
	}
	m.Role = string(role)
	return s.repo.UpdateMember(m)
}

func (s *FamilyService) RemoveMember(ctx context.Context, familyID, userID uuid.UUID) error {
	callerID := ctx.Value("user_id").(string)
	caller, err := s.repo.FindMember(familyID.String(), callerID)
	if err != nil || (caller.Role != "owner" && caller.Role != "admin") {
		return fmt.Errorf("only owner/admin can remove members")
	}
	return s.repo.DeleteMember(familyID.String(), userID.String())
}

func (s *FamilyService) LeaveFamily(ctx context.Context, familyID uuid.UUID) error {
	userID := ctx.Value("user_id").(string)
	return s.repo.DeleteMember(familyID.String(), userID)
}

// ─── Join Requests ────────────────────────────────────────────────

func (s *FamilyService) ListJoinRequests(ctx context.Context, familyID uuid.UUID) ([]types.FamilyMember, error) {
	members, err := s.repo.ListMembersByStatus(familyID.String(), "pending")
	if err != nil {
		return nil, err
	}
	result := make([]types.FamilyMember, len(members))
	for i, m := range members {
		result[i] = *toFamilyMember(&m)
	}
	return result, nil
}

func (s *FamilyService) ReviewJoinRequest(ctx context.Context, familyID uuid.UUID, req *types.ReviewJoinRequest) error {
	if req.Action != types.MemberStatusActive && req.Action != types.MemberStatusRejected {
		return fmt.Errorf("action must be 'active' or 'rejected'")
	}
	callerID := ctx.Value("user_id").(string)
	caller, err := s.repo.FindMember(familyID.String(), callerID)
	if err != nil || (caller.Role != "owner" && caller.Role != "admin") {
		return fmt.Errorf("only owner/admin can review join requests")
	}
	m, err := s.repo.FindMember(familyID.String(), req.UserID)
	if err != nil || m.Status != "pending" {
		return fmt.Errorf("join request not found")
	}
	m.Status = string(req.Action)
	if req.Action == types.MemberStatusActive {
		m.JoinedAt = time.Now()
	}
	return s.repo.UpdateMember(m)
}

// ─── Family Group ─────────────────────────────────────────────────

func toGroup(g *repository.FamilyGroupModel) *types.FamilyGroup {
	return &types.FamilyGroup{ID: g.ID, FamilyID: g.FamilyID, Name: g.Name, Description: g.Description, CreatedBy: g.CreatedBy, CreatedAt: g.CreatedAt, UpdatedAt: g.UpdatedAt}
}

func toGroupMember(m *repository.FamilyGroupMemberModel) *types.FamilyGroupMember {
	var u *types.User
	if m.User.ID != "" {
		u = userModelToUser(&m.User)
	}
	return &types.FamilyGroupMember{ID: m.ID, GroupID: m.GroupID, UserID: m.UserID, Role: types.GroupRole(m.Role), Status: types.MemberStatus(m.Status), JoinedAt: m.JoinedAt, User: u}
}

func (s *FamilyService) CreateGroup(ctx context.Context, familyID uuid.UUID, req *types.CreateFamilyGroupRequest) (*types.FamilyGroup, error) {
	userID := ctx.Value("user_id").(string)
	g := &repository.FamilyGroupModel{FamilyID: familyID.String(), Name: req.Name, Description: req.Description, CreatedBy: userID}
	if err := s.repo.CreateGroup(g); err != nil {
		return nil, err
	}
	gm := &repository.FamilyGroupMemberModel{GroupID: g.ID, UserID: userID, Role: "owner", Status: "active", JoinedAt: time.Now()}
	if err := s.repo.AddGroupMember(gm); err != nil {
		return nil, err
	}
	return toGroup(g), nil
}

func (s *FamilyService) ListGroups(ctx context.Context, familyID uuid.UUID) ([]types.FamilyGroup, error) {
	groups, err := s.repo.ListGroups(familyID.String())
	if err != nil {
		return nil, err
	}
	result := make([]types.FamilyGroup, len(groups))
	for i, g := range groups {
		result[i] = *toGroup(&g)
	}
	return result, nil
}

func (s *FamilyService) JoinGroup(ctx context.Context, groupID uuid.UUID) (*types.FamilyGroupMember, error) {
	userID := ctx.Value("user_id").(string)
	if _, err := s.repo.FindGroupByID(groupID.String()); err != nil {
		return nil, fmt.Errorf("group not found")
	}
	if existing, _ := s.repo.FindGroupMember(groupID.String(), userID); existing != nil {
		if existing.Status == "active" {
			return nil, fmt.Errorf("already a member")
		}
		return nil, fmt.Errorf("join request already exists")
	}
	m := &repository.FamilyGroupMemberModel{GroupID: groupID.String(), UserID: userID, Role: "member", Status: "pending", JoinedAt: time.Now()}
	if err := s.repo.AddGroupMember(m); err != nil {
		return nil, err
	}
	return toGroupMember(m), nil
}

func (s *FamilyService) LeaveGroup(ctx context.Context, groupID uuid.UUID) error {
	userID := ctx.Value("user_id").(string)
	return s.repo.DeleteGroupMember(groupID.String(), userID)
}

func (s *FamilyService) ListGroupMembers(ctx context.Context, groupID uuid.UUID) ([]types.FamilyGroupMember, error) {
	members, err := s.repo.ListGroupMembers(groupID.String())
	if err != nil {
		return nil, err
	}
	result := make([]types.FamilyGroupMember, len(members))
	for i, m := range members {
		result[i] = *toGroupMember(&m)
	}
	return result, nil
}

func (s *FamilyService) RemoveGroupMember(ctx context.Context, groupID, userID uuid.UUID) error {
	callerID := ctx.Value("user_id").(string)
	g, err := s.repo.FindGroupByID(groupID.String())
	if err != nil {
		return err
	}
	if g.CreatedBy != callerID {
		return fmt.Errorf("only group creator can remove members")
	}
	return s.repo.DeleteGroupMember(groupID.String(), userID.String())
}

func (s *FamilyService) ListGroupJoinRequests(ctx context.Context, groupID uuid.UUID) ([]types.FamilyGroupMember, error) {
	members, err := s.repo.ListGroupMembersByStatus(groupID.String(), "pending")
	if err != nil {
		return nil, err
	}
	result := make([]types.FamilyGroupMember, len(members))
	for i, m := range members {
		result[i] = *toGroupMember(&m)
	}
	return result, nil
}

func (s *FamilyService) ReviewGroupJoinRequest(ctx context.Context, groupID uuid.UUID, req *types.ReviewGroupJoinRequest) error {
	if req.Action != types.MemberStatusActive && req.Action != types.MemberStatusRejected {
		return fmt.Errorf("action must be 'active' or 'rejected'")
	}
	callerID := ctx.Value("user_id").(string)
	g, err := s.repo.FindGroupByID(groupID.String())
	if err != nil {
		return err
	}
	if g.CreatedBy != callerID {
		return fmt.Errorf("only group creator can review join requests")
	}
	m, err := s.repo.FindGroupMember(groupID.String(), req.UserID)
	if err != nil || m.Status != "pending" {
		return fmt.Errorf("join request not found")
	}
	m.Status = string(req.Action)
	if req.Action == types.MemberStatusActive {
		m.JoinedAt = time.Now()
	}
	return s.repo.UpdateGroupMember(m)
}
