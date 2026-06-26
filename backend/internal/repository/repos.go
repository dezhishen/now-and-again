package repository

import "gorm.io/gorm"

// ─── Repository Definitions ───────────────────────────────────────

type UserRepo struct{ db *gorm.DB }

func NewUserRepo(db *gorm.DB) *UserRepo { return &UserRepo{db} }

type FamilyRepo struct{ db *gorm.DB }

func NewFamilyRepo(db *gorm.DB) *FamilyRepo { return &FamilyRepo{db} }

type ApiKeyRepo struct{ db *gorm.DB }

func NewApiKeyRepo(db *gorm.DB) *ApiKeyRepo { return &ApiKeyRepo{db} }
