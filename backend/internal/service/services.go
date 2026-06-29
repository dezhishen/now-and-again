package service

import (
	"github.com/dezhishen/now-and-again/backend/internal/repository"
	"github.com/dezhishen/now-and-again/backend/pkg/contracts"
)

// ─── Compile-time contract checks ────────────────────────────────

var (
	_ contracts.UserContract      = (*UserService)(nil)
	_ contracts.FamilyContract    = (*FamilyService)(nil)
	_ contracts.ApiKeyContract    = (*ApiKeyService)(nil)
	_ contracts.FloorPlanContract = (*FloorPlanService)(nil)
	_ contracts.LocationContract  = (*FloorPlanService)(nil)
	_ contracts.TaskContract      = (*TaskService)(nil)
	_ contracts.TodoContract      = (*TodoService)(nil)
	_ contracts.LogContract       = (*LogService)(nil)
	_ contracts.CalendarContract  = (*CalendarService)(nil)
)

// ─── User ─────────────────────────────────────────────────────────

type UserService struct {
	repo      *repository.UserRepo
	jwtSecret string
}

func NewUserService(repo *repository.UserRepo, jwtSecret string) *UserService {
	return &UserService{repo: repo, jwtSecret: jwtSecret}
}

// ─── Family ───────────────────────────────────────────────────────

type FamilyService struct {
	repo     *repository.FamilyRepo
	userRepo *repository.UserRepo
}

func NewFamilyService(repo *repository.FamilyRepo, userRepo *repository.UserRepo) *FamilyService {
	return &FamilyService{repo: repo, userRepo: userRepo}
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
	repo      *repository.FloorPlanRepo
	userRepo  *repository.UserRepo
	imageSvc  *ImageService
	imageRepo *repository.ImageRepo
}

func NewFloorPlanService(repo *repository.FloorPlanRepo, userRepo *repository.UserRepo, imageSvc *ImageService, imageRepo *repository.ImageRepo) *FloorPlanService {
	return &FloorPlanService{repo: repo, userRepo: userRepo, imageSvc: imageSvc, imageRepo: imageRepo}
}

// ─── All Contracts ────────────────────────────────────────────────

func NewAllContracts(user *UserService, family *FamilyService, apiKey *ApiKeyService, floorPlan *FloorPlanService, task *TaskService, todo *TodoService, log *LogService, calendar *CalendarService) *contracts.AllContracts {
	return &contracts.AllContracts{
		User:      user,
		Family:    family,
		ApiKey:    apiKey,
		FloorPlan: floorPlan,
		Location:  floorPlan, // FloorPlanService also implements LocationContract
		Task:      task,
		Todo:      todo,
		Log:       log,
		Calendar:  calendar,
	}
}
