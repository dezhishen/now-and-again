package service

import (
	"github.com/dezhishen/now-and-again/backend/internal/notifier"
	"github.com/dezhishen/now-and-again/backend/internal/repository"
	"github.com/dezhishen/now-and-again/shared/contracts"
)

// ─── Compile-time interface compliance ────────────────────────────
// These assertions ensure backend services always satisfy the shared contracts.
// Adding a method to contracts will break compilation here until implemented.

var (
	_ contracts.UserContract         = (*UserService)(nil)
	_ contracts.FamilyContract       = (*FamilyService)(nil)
	_ contracts.SubGroupContract     = (*SubGroupService)(nil)
	_ contracts.TaskContract         = (*TaskService)(nil)
	_ contracts.ChainContract        = (*ChainService)(nil)
	_ contracts.InspectionContract   = (*InspectionService)(nil)
	_ contracts.LogContract          = (*LogService)(nil)
	_ contracts.NotificationContract = (*NotificationService)(nil)
	_ contracts.ApiKeyContract       = (*ApiKeyService)(nil)
)

// ─── User ────────────────────────────────────────────────────────

type UserService struct {
	repo      *repository.UserRepo
	jwtSecret string
}

func NewUserService(repo *repository.UserRepo, jwtSecret string) *UserService {
	return &UserService{repo: repo, jwtSecret: jwtSecret}
}

// ─── Family ──────────────────────────────────────────────────────

type FamilyService struct {
	repo     *repository.FamilyRepo
	userRepo *repository.UserRepo
}

func NewFamilyService(repo *repository.FamilyRepo, userRepo *repository.UserRepo) *FamilyService {
	return &FamilyService{repo: repo, userRepo: userRepo}
}

// ─── SubGroup ────────────────────────────────────────────────────

type SubGroupService struct {
	repo       *repository.SubGroupRepo
	familyRepo *repository.FamilyRepo
}

func NewSubGroupService(repo *repository.SubGroupRepo, familyRepo *repository.FamilyRepo) *SubGroupService {
	return &SubGroupService{repo: repo, familyRepo: familyRepo}
}

// ─── Task ────────────────────────────────────────────────────────

type TaskService struct {
	repo    *repository.TaskRepo
	logRepo *repository.LogRepo
	notif   *notifier.NotificationEngine
}

func NewTaskService(repo *repository.TaskRepo, logRepo *repository.LogRepo, notif *notifier.NotificationEngine) *TaskService {
	return &TaskService{repo: repo, logRepo: logRepo, notif: notif}
}

// ─── Chain ───────────────────────────────────────────────────────

type ChainService struct {
	repo     *repository.ChainRepo
	taskRepo *repository.TaskRepo
	logRepo  *repository.LogRepo
	notif    *notifier.NotificationEngine
}

func NewChainService(repo *repository.ChainRepo, taskRepo *repository.TaskRepo, logRepo *repository.LogRepo, notif *notifier.NotificationEngine) *ChainService {
	return &ChainService{repo: repo, taskRepo: taskRepo, logRepo: logRepo, notif: notif}
}

// ─── Inspection ──────────────────────────────────────────────────

type InspectionService struct {
	repo     *repository.InspectionRepo
	taskRepo *repository.TaskRepo
	logRepo  *repository.LogRepo
	notif    *notifier.NotificationEngine
}

func NewInspectionService(repo *repository.InspectionRepo, taskRepo *repository.TaskRepo, logRepo *repository.LogRepo, notif *notifier.NotificationEngine) *InspectionService {
	return &InspectionService{repo: repo, taskRepo: taskRepo, logRepo: logRepo, notif: notif}
}

// ─── Log ─────────────────────────────────────────────────────────

type LogService struct {
	repo *repository.LogRepo
}

func NewLogService(repo *repository.LogRepo) *LogService {
	return &LogService{repo: repo}
}

// ─── Notification ────────────────────────────────────────────────

type NotificationService struct {
	repo     *repository.NotificationRepo
	userRepo *repository.UserRepo
}

func NewNotificationService(repo *repository.NotificationRepo, userRepo *repository.UserRepo) *NotificationService {
	return &NotificationService{repo: repo, userRepo: userRepo}
}

// ─── API Key ─────────────────────────────────────────────────────
// (ApiKeyService and NewApiKeyService are defined in apikey_service.go)

// ─── Bundle (satisfies contracts.AllContracts) ───────────────────

func NewAllContracts(
	user *UserService,
	family *FamilyService,
	subGroup *SubGroupService,
	task *TaskService,
	chain *ChainService,
	inspection *InspectionService,
	log *LogService,
	notif *NotificationService,
	apiKey *ApiKeyService,
) *contracts.AllContracts {
	return &contracts.AllContracts{
		User:         user,
		Family:       family,
		SubGroup:     subGroup,
		Task:         task,
		Chain:        chain,
		Inspection:   inspection,
		Log:          log,
		Notification: notif,
		ApiKey:       apiKey,
	}
}
