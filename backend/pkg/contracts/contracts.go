package contracts

import (
	"context"
	"mime/multipart"

	"github.com/dezhishen/now-and-again/backend/pkg/types"
	"github.com/dezhishen/now-and-again/backend/pkg/types/task"
	"github.com/google/uuid"
)

// ─── User ─────────────────────────────────────────────────────────

type UserContract interface {
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
	Update(ctx context.Context, familyID uuid.UUID, req *types.UpdateFamilyRequest) (*types.Family, error)
	Delete(ctx context.Context, familyID uuid.UUID) error
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

// ─── Floor Plan ──────────────────────────────────────────────────

type FloorPlanContract interface {
	Upload(ctx context.Context, familyID uuid.UUID, label string, isCover bool, file multipart.File, header *multipart.FileHeader) (*types.FloorPlan, error)
	ListByFamily(ctx context.Context, familyID uuid.UUID) ([]types.FloorPlan, error)
	GetByID(ctx context.Context, planID uuid.UUID) (*types.FloorPlan, error)
	SetCover(ctx context.Context, planID uuid.UUID) error
	Delete(ctx context.Context, planID uuid.UUID) error

	// Locations (first-class, independent entity)
	CreateLocation(ctx context.Context, familyID uuid.UUID, req *types.CreateLocationRequest) (*types.Location, error)
	ListFamilyLocations(ctx context.Context, familyID uuid.UUID) ([]types.Location, error)
	ListFloorPlanLocations(ctx context.Context, floorPlanID uuid.UUID) ([]types.Location, error)
	UpdateLocation(ctx context.Context, locationID uuid.UUID, req *types.UpdateLocationRequest) (*types.Location, error)
	DeleteLocation(ctx context.Context, locationID uuid.UUID) error
}

// ─── Task ─────────────────────────────────────────────────────────

// TaskContract defines the core task/todo operations that both server and CLI must implement.
type TaskContract interface {
	CreateTask(ctx context.Context, familyID uuid.UUID, req *task.CreateTaskRequest) (*task.Task, error)
	UpdateTask(ctx context.Context, taskID uuid.UUID, req *task.UpdateTaskRequest) (*task.Task, error)
	DeleteTask(ctx context.Context, taskID uuid.UUID) error
	ListTasks(ctx context.Context, familyID uuid.UUID) ([]task.Task, error)
	TriggerTask(ctx context.Context, taskID uuid.UUID) error
}

type TodoContract interface {
	ListTodos(ctx context.Context, familyID uuid.UUID, groupID, status string) ([]types.Todo, error)
	CompleteTodo(ctx context.Context, todoID uuid.UUID, req *types.CompleteTodoRequest) (*types.Todo, error)
}

type LogContract interface {
	ListLogs(ctx context.Context, taskID uuid.UUID, limit int, userOnly bool) ([]types.TaskLog, error)
}

type AllContracts struct {
	User      UserContract
	Family    FamilyContract
	ApiKey    ApiKeyContract
	FloorPlan FloorPlanContract
	Task      TaskContract
	Todo      TodoContract
	Log       LogContract
}
