package contracts

import (
	"context"

	"github.com/dezhishen/now-and-again/shared/types"
	"github.com/google/uuid"
)

// ─── User ─────────────────────────────────────────────────────────

type UserContract interface {
	Setup(ctx context.Context, req *types.SetupRequest) (*types.User, error)
	CheckInit(ctx context.Context) (*types.SystemStatus, error)
	Register(ctx context.Context, req *types.CreateUserRequest) (*types.User, error)
	Login(ctx context.Context, req *types.LoginRequest) (*types.TokenPair, error)
	Refresh(ctx context.Context, refreshToken string) (*types.TokenPair, error)
	Logout(ctx context.Context) error
	GetMe(ctx context.Context) (*types.User, error)
	UpdateMe(ctx context.Context, req *types.UpdateUserRequest) (*types.User, error)
	ListUsers(ctx context.Context) ([]types.User, error)
}

// ─── Family ───────────────────────────────────────────────────────

type FamilyContract interface {
	// Family CRUD
	Create(ctx context.Context, req *types.CreateFamilyRequest) (*types.Family, error)
	Join(ctx context.Context, req *types.JoinFamilyRequest) (*types.FamilyMember, error)
	Get(ctx context.Context, familyID uuid.UUID) (*types.Family, error)
	ListMyFamilies(ctx context.Context) ([]types.Family, error)

	// Member management
	ListMembers(ctx context.Context, familyID uuid.UUID) ([]types.FamilyMember, error)
	UpdateMemberRole(ctx context.Context, familyID, userID uuid.UUID, role types.FamilyRole) error
	RemoveMember(ctx context.Context, familyID, userID uuid.UUID) error
	LeaveFamily(ctx context.Context, familyID uuid.UUID) error

	// Join request review (owner/admin only)
	ListJoinRequests(ctx context.Context, familyID uuid.UUID) ([]types.FamilyMember, error)
	ReviewJoinRequest(ctx context.Context, familyID uuid.UUID, req *types.ReviewJoinRequest) error

	// Family Group
	CreateGroup(ctx context.Context, familyID uuid.UUID, req *types.CreateFamilyGroupRequest) (*types.FamilyGroup, error)
	ListGroups(ctx context.Context, familyID uuid.UUID) ([]types.FamilyGroup, error)
	JoinGroup(ctx context.Context, groupID uuid.UUID) (*types.FamilyGroupMember, error)
	LeaveGroup(ctx context.Context, groupID uuid.UUID) error
	ListGroupMembers(ctx context.Context, groupID uuid.UUID) ([]types.FamilyGroupMember, error)
	RemoveGroupMember(ctx context.Context, groupID, userID uuid.UUID) error
	ListGroupJoinRequests(ctx context.Context, groupID uuid.UUID) ([]types.FamilyGroupMember, error)
	ReviewGroupJoinRequest(ctx context.Context, groupID uuid.UUID, req *types.ReviewGroupJoinRequest) error
}

// ─── API Key ──────────────────────────────────────────────────────

type ApiKeyContract interface {
	Create(ctx context.Context, req *types.CreateApiKeyRequest) (*types.CreateApiKeyResponse, error)
	List(ctx context.Context) ([]types.ApiKey, error)
	Revoke(ctx context.Context, keyID uuid.UUID) error
}

// ─── Aggregate ────────────────────────────────────────────────────

type AllContracts struct {
	User   UserContract
	Family FamilyContract
	ApiKey ApiKeyContract
}
