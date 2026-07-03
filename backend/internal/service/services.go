package service

import (
	"github.com/dezhishen/now-and-again/backend/internal/repository"
	"github.com/dezhishen/now-and-again/backend/pkg/contracts"
)

// ─── Compile-time contract checks ────────────────────────────────

var (
	_ contracts.UserContract         = (*UserService)(nil)
	_ contracts.FamilyContract       = (*FamilyService)(nil)
	_ contracts.ApiKeyContract       = (*ApiKeyService)(nil)
	_ contracts.FloorPlanContract    = (*FloorPlanService)(nil)
	_ contracts.LocationContract     = (*FloorPlanService)(nil)
	_ contracts.TaskContract         = (*TaskService)(nil)
	_ contracts.TodoContract         = (*TodoService)(nil)
	_ contracts.LogContract          = (*LogService)(nil)
	_ contracts.CalendarContract     = (*CalendarService)(nil)
	_ contracts.TaskTemplateContract = (*TaskTemplateService)(nil)
)

// ─── User ─────────────────────────────────────────────────────────

type UserService struct {
	repo         *repository.UserRepo
	settingsRepo *repository.SettingsRepo
	jwtSecret    string
}

func NewUserService(repo *repository.UserRepo, settingsRepo *repository.SettingsRepo, jwtSecret string) *UserService {
	return &UserService{repo: repo, settingsRepo: settingsRepo, jwtSecret: jwtSecret}
}

// ─── Family ───────────────────────────────────────────────────────

type FamilyService struct {
	repo          *repository.FamilyRepo
	floorPlanRepo *repository.FloorPlanRepo
	userRepo      *repository.UserRepo
}

func NewFamilyService(repo *repository.FamilyRepo, floorPlanRepo *repository.FloorPlanRepo, userRepo *repository.UserRepo) *FamilyService {
	return &FamilyService{repo: repo, floorPlanRepo: floorPlanRepo, userRepo: userRepo}
}

// ─── API Key ──────────────────────────────────────────────────────

type ApiKeyService struct {
	repo *repository.ApiKeyRepo
}

func NewApiKeyService(repo *repository.ApiKeyRepo) *ApiKeyService {
	return &ApiKeyService{repo: repo}
}

// ─── Floor Plan ──────────────────────────────────────────────────

type FloorPlanService struct {
	repo       *repository.FloorPlanRepo
	familyRepo *repository.FamilyRepo
	userRepo   *repository.UserRepo
	imageSvc   *ImageService
	imageRepo  *repository.ImageRepo
}

func NewFloorPlanService(repo *repository.FloorPlanRepo, familyRepo *repository.FamilyRepo, userRepo *repository.UserRepo, imageSvc *ImageService, imageRepo *repository.ImageRepo) *FloorPlanService {
	return &FloorPlanService{repo: repo, familyRepo: familyRepo, userRepo: userRepo, imageSvc: imageSvc, imageRepo: imageRepo}
}

// ─── All Contracts ────────────────────────────────────────────────

func NewAllContracts(user *UserService, family *FamilyService, apiKey *ApiKeyService, floorPlan *FloorPlanService, task *TaskService, todo *TodoService, log *LogService, calendar *CalendarService, taskTemplate *TaskTemplateService) *contracts.AllContracts {
	return &contracts.AllContracts{
		User:         user,
		Family:       family,
		ApiKey:       apiKey,
		FloorPlan:    floorPlan,
		Location:     floorPlan, // FloorPlanService also implements LocationContract
		Task:         task,
		Todo:         todo,
		Log:          log,
		Calendar:     calendar,
		TaskTemplate: taskTemplate,
	}
}
