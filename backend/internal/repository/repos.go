package repository

import "gorm.io/gorm"

// ─── Repository Interfaces ───────────────────────────────────────
// Each repository is backed by a concrete GORM implementation.
// We define interfaces at the package level to enable test mocking.

// UserRepo handles user CRUD and auth lookups.
type UserRepo struct{ db *gorm.DB }
func NewUserRepo(db *gorm.DB) *UserRepo { return &UserRepo{db} }

// FamilyRepo handles family and membership CRUD.
type FamilyRepo struct{ db *gorm.DB }
func NewFamilyRepo(db *gorm.DB) *FamilyRepo { return &FamilyRepo{db} }

// SubGroupRepo handles sub-group and membership CRUD.
type SubGroupRepo struct{ db *gorm.DB }
func NewSubGroupRepo(db *gorm.DB) *SubGroupRepo { return &SubGroupRepo{db} }

// TaskRepo handles task, task type, assignee, and dependency CRUD.
type TaskRepo struct{ db *gorm.DB }
func NewTaskRepo(db *gorm.DB) *TaskRepo { return &TaskRepo{db} }

// ChainRepo handles task chain template and step CRUD.
// ApiKeyRepo handles API key CRUD.
type ApiKeyRepo struct{ db *gorm.DB }
func NewApiKeyRepo(db *gorm.DB) *ApiKeyRepo { return &ApiKeyRepo{db} }
type ChainRepo struct{ db *gorm.DB }
func NewChainRepo(db *gorm.DB) *ChainRepo { return &ChainRepo{db} }

// InspectionRepo handles inspection and inspection item CRUD.
type InspectionRepo struct{ db *gorm.DB }
func NewInspectionRepo(db *gorm.DB) *InspectionRepo { return &InspectionRepo{db} }

// LogRepo handles task log append and query.
type LogRepo struct{ db *gorm.DB }
func NewLogRepo(db *gorm.DB) *LogRepo { return &LogRepo{db} }

// NotificationRepo handles notification channel, template, user config, and delivery record CRUD.
type NotificationRepo struct{ db *gorm.DB }
func NewNotificationRepo(db *gorm.DB) *NotificationRepo { return &NotificationRepo{db} }
