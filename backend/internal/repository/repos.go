package repository

import "gorm.io/gorm"

// ─── Repository Definitions ───────────────────────────────────────

type UserRepo struct{ db *gorm.DB }

func NewUserRepo(db *gorm.DB) *UserRepo { return &UserRepo{db} }

type FamilyRepo struct{ db *gorm.DB }

func NewFamilyRepo(db *gorm.DB) *FamilyRepo { return &FamilyRepo{db} }

type ApiKeyRepo struct{ db *gorm.DB }

func NewApiKeyRepo(db *gorm.DB) *ApiKeyRepo { return &ApiKeyRepo{db} }

type FloorPlanRepo struct{ db *gorm.DB }

func NewFloorPlanRepo(db *gorm.DB) *FloorPlanRepo { return &FloorPlanRepo{db} }

type ImageRepo struct{ db *gorm.DB }

func NewImageRepo(db *gorm.DB) *ImageRepo { return &ImageRepo{db} }

type SettingsRepo struct{ db *gorm.DB }

func NewSettingsRepo(db *gorm.DB) *SettingsRepo { return &SettingsRepo{db} }

type TaskRepo struct{ db *gorm.DB }

func NewTaskRepo(db *gorm.DB) *TaskRepo { return &TaskRepo{db} }

type IcsRepo struct{ db *gorm.DB }

func NewIcsRepo(db *gorm.DB) *IcsRepo { return &IcsRepo{db} }
