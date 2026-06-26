package service

import (
	"github.com/dezhishen/now-and-again/backend/internal/repository"
	"github.com/dezhishen/now-and-again/shared/contracts"
)

// ─── Compile-time contract compliance ─────────────────────────────

var (
	_ contracts.UserContract   = (*UserService)(nil)
	_ contracts.FamilyContract = (*FamilyService)(nil)
	_ contracts.ApiKeyContract = (*ApiKeyService)(nil)
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

// ─── All Contracts ────────────────────────────────────────────────

func NewAllContracts(user *UserService, family *FamilyService, apiKey *ApiKeyService) *contracts.AllContracts {
	return &contracts.AllContracts{
		User:   user,
		Family: family,
		ApiKey: apiKey,
	}
}
