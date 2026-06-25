// Package contracts defines the interface contracts that BOTH
// backend services and CLI API clients must implement.
//
// Adding a new method to any contract will cause compilation failures
// in BOTH backend and CLI — this enforces synchronous evolution.
//
// Each domain contract is implemented by:
//   - backend/internal/service  (business logic, database-backed)
//   - cli/internal/client       (HTTP API calls to backend)
//
// Compile-time assertions in both packages enforce compliance:
//
//	var _ contracts.UserContract = (*service.UserService)(nil)
//	var _ contracts.UserContract = (*client.UserClient)(nil)
package contracts

import (
	"context"

	"github.com/dezhishen/now-and-again/shared/types"
	"github.com/google/uuid"
)

// ─── User ─────────────────────────────────────────────────────────

type UserContract interface {
	// Setup creates the first admin user (only when system is uninitialized).
	Setup(ctx context.Context, req *types.SetupRequest) (*types.User, error)
	// CheckInit returns whether the system has been initialized.
	CheckInit(ctx context.Context) (*types.SystemStatus, error)
	Register(ctx context.Context, req *types.CreateUserRequest) (*types.User, error)
	Login(ctx context.Context, req *types.LoginRequest) (*types.TokenPair, error)
	// Refresh issues a new access + refresh token pair.
	Refresh(ctx context.Context, refreshToken string) (*types.TokenPair, error)
	// Logout invalidates the current refresh token.
	Logout(ctx context.Context) error
	GetMe(ctx context.Context) (*types.User, error)
	UpdateMe(ctx context.Context, req *types.UpdateUserRequest) (*types.User, error)
	// ListUsers returns all users (admin only).
	ListUsers(ctx context.Context) ([]types.User, error)
}

// ─── Family ──────────────────────────────────────────────────────

type FamilyContract interface {
	Create(ctx context.Context, req *types.CreateFamilyRequest) (*types.Family, error)
	Join(ctx context.Context, req *types.JoinFamilyRequest) (*types.FamilyMember, error)
	Get(ctx context.Context, familyID uuid.UUID) (*types.Family, error)
	// ListMyFamilies returns families the current user belongs to.
	ListMyFamilies(ctx context.Context) ([]types.Family, error)
	ListMembers(ctx context.Context, familyID uuid.UUID) ([]types.FamilyMember, error)
	UpdateMemberRole(ctx context.Context, familyID, userID uuid.UUID, role types.FamilyRole) error
	RemoveMember(ctx context.Context, familyID, userID uuid.UUID) error
	// LeaveFamily removes the current user from a family.
	LeaveFamily(ctx context.Context, familyID uuid.UUID) error
}

// ─── SubGroup ────────────────────────────────────────────────────

type SubGroupContract interface {
	Create(ctx context.Context, familyID uuid.UUID, req *types.CreateSubGroupRequest) (*types.SubGroup, error)
	List(ctx context.Context, familyID uuid.UUID) ([]types.SubGroup, error)
	AddMember(ctx context.Context, subGroupID uuid.UUID, req *types.AddSubGroupMemberRequest) (*types.SubGroupMember, error)
	RemoveMember(ctx context.Context, subGroupID, userID uuid.UUID) error
}

// ─── Task ────────────────────────────────────────────────────────

type TaskContract interface {
	Create(ctx context.Context, req *types.CreateTaskRequest) (*types.Task, error)
	List(ctx context.Context, familyID uuid.UUID, status *types.TaskStatus, assigneeID *uuid.UUID, page, pageSize int) ([]types.Task, int, error)
	Get(ctx context.Context, taskID uuid.UUID) (*types.Task, error)
	Update(ctx context.Context, taskID uuid.UUID, req *types.UpdateTaskRequest) (*types.Task, error)
	SetAssignees(ctx context.Context, taskID uuid.UUID, req *types.SetAssigneesRequest) ([]types.TaskAssignee, error)
	AddDependency(ctx context.Context, taskID uuid.UUID, req *types.CreateDependencyRequest) (*types.TaskDependency, error)
	RemoveDependency(ctx context.Context, taskID, depID uuid.UUID) error
}

// ─── Chain ───────────────────────────────────────────────────────

type ChainContract interface {
	Create(ctx context.Context, familyID uuid.UUID, req *types.CreateChainRequest) (*types.TaskChain, error)
	List(ctx context.Context, familyID uuid.UUID) ([]types.TaskChain, error)
	Get(ctx context.Context, chainID uuid.UUID) (*types.TaskChain, error)
	AddStep(ctx context.Context, chainID uuid.UUID, req *types.AddStepRequest) (*types.TaskChainStep, error)
	ReorderSteps(ctx context.Context, chainID uuid.UUID, req *types.ReorderStepsRequest) error
	RemoveStep(ctx context.Context, chainID, stepID uuid.UUID) error
	Start(ctx context.Context, chainID uuid.UUID) (*types.StartChainResponse, error)
}

// ─── Inspection ──────────────────────────────────────────────────

type InspectionContract interface {
	Create(ctx context.Context, familyID uuid.UUID, req *types.CreateInspectionRequest) (*types.Inspection, error)
	List(ctx context.Context, familyID uuid.UUID) ([]types.Inspection, error)
	Get(ctx context.Context, inspectionID uuid.UUID) (*types.Inspection, error)
	AddItem(ctx context.Context, inspectionID uuid.UUID, req *types.AddInspectionItemRequest) (*types.InspectionItem, error)
	UpdateItem(ctx context.Context, inspectionID, itemID uuid.UUID, req *types.UpdateInspectionItemRequest) (*types.InspectionItem, error)
	Complete(ctx context.Context, inspectionID uuid.UUID) (*types.Inspection, error)
}

// ─── Log ─────────────────────────────────────────────────────────

type LogContract interface {
	ListByTask(ctx context.Context, taskID uuid.UUID) ([]types.TaskLog, error)
	AddComment(ctx context.Context, taskID uuid.UUID, req *types.AddCommentRequest) (*types.TaskLog, error)
}

// ─── Notification ────────────────────────────────────────────────

type NotificationContract interface {
	List(ctx context.Context, page, pageSize int) ([]types.Notification, int, error)
	UpsertChannelConfig(ctx context.Context, req *types.UpsertUserChannelRequest) (*types.UserChannelConfig, error)
	ListTemplates(ctx context.Context, familyID uuid.UUID) ([]types.NotificationTemplate, error)
	UpsertTemplate(ctx context.Context, familyID uuid.UUID, req *types.UpsertTemplateRequest) (*types.NotificationTemplate, error)
}

// ─── API Key ─────────────────────────────────────────────────────

type ApiKeyContract interface {
	Create(ctx context.Context, req *types.CreateApiKeyRequest) (*types.CreateApiKeyResponse, error)
	List(ctx context.Context) ([]types.ApiKey, error)
	Revoke(ctx context.Context, keyID string) error
}

// ─── Aggregate (for handler injection) ───────────────────────────

// AllContracts bundles all domain contracts for dependency injection
// in the handler layer and CLI command layer.
type AllContracts struct {
	User         UserContract
	Family       FamilyContract
	SubGroup     SubGroupContract
	Task         TaskContract
	Chain        ChainContract
	Inspection   InspectionContract
	Log          LogContract
	Notification NotificationContract
	ApiKey       ApiKeyContract
}
