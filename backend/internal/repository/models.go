package repository

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ─── Base Model ───────────────────────────────────────────────────

type BaseModel struct {
	ID        string `gorm:"primaryKey;type:char(36)"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (b *BaseModel) BeforeCreate(_ *gorm.DB) error {
	if b.ID == "" {
		b.ID = uuid.New().String()
	}
	return nil
}

// ─── Account / User / Role ────────────────────────────────────────
// Account n:1 User. Currently only "local" provider; future OAuth providers.

type AccountModel struct {
	BaseModel
	UserID            string `gorm:"index;type:char(36);not null"`
	Provider          string `gorm:"size:32;not null;default:local"`
	ProviderAccountID string `gorm:"size:255"`
	Username          string `gorm:"uniqueIndex;size:64"`
	PasswordHash      string `gorm:"size:255"`
}

func (AccountModel) TableName() string { return "accounts" }

type UserModel struct {
	BaseModel
	DisplayName string `gorm:"size:128;not null"`
	Email       string `gorm:"uniqueIndex;size:255"`
	Phone       string `gorm:"size:20"`
	AvatarURL   string `gorm:"type:text"`

	Accounts []AccountModel  `gorm:"foreignKey:UserID"`
	Roles    []UserRoleModel `gorm:"foreignKey:UserID"`
}

type RoleModel struct {
	BaseModel
	Name        string `gorm:"uniqueIndex;size:64;not null"`
	Description string `gorm:"type:text"`
}

type UserRoleModel struct {
	UserID string    `gorm:"primaryKey;type:char(36)"`
	RoleID string    `gorm:"primaryKey;type:char(36)"`
	Role   RoleModel `gorm:"foreignKey:RoleID"`
}

func (UserRoleModel) TableName() string { return "user_roles" }

// ─── Family ───────────────────────────────────────────────────────

type FamilyModel struct {
	BaseModel
	Name       string `gorm:"size:128;not null"`
	InviteCode string `gorm:"uniqueIndex;size:32;not null"`
	CreatedBy  string `gorm:"type:char(36);not null"`
}

type FamilyMemberModel struct {
	BaseModel
	FamilyID string `gorm:"uniqueIndex:idx_family_user;type:char(36);not null"`
	UserID   string `gorm:"uniqueIndex:idx_family_user;type:char(36);not null"`
	Role     string `gorm:"size:16;not null;default:member"`
	Status   string `gorm:"size:16;not null;default:active"`
	JoinedAt time.Time
	User     UserModel `gorm:"foreignKey:UserID"`
}

// ─── Family Group ─────────────────────────────────────────────────

type FamilyGroupModel struct {
	BaseModel
	FamilyID    string `gorm:"index;type:char(36);not null"`
	Name        string `gorm:"size:128;not null"`
	Description string `gorm:"type:text"`
	CreatedBy   string `gorm:"type:char(36);not null"`
}

func (FamilyGroupModel) TableName() string { return "family_groups" }

type FamilyGroupMemberModel struct {
	BaseModel
	GroupID  string `gorm:"uniqueIndex:idx_group_user;type:char(36);not null"`
	UserID   string `gorm:"uniqueIndex:idx_group_user;type:char(36);not null"`
	Role     string `gorm:"size:16;not null;default:member"`
	Status   string `gorm:"size:16;not null;default:pending"`
	JoinedAt time.Time
	User     UserModel `gorm:"foreignKey:UserID"`
}

func (FamilyGroupMemberModel) TableName() string { return "family_group_members" }

// ─── Refresh Token ────────────────────────────────────────────────

type RefreshTokenModel struct {
	ID        string    `gorm:"primaryKey;type:char(36)"`
	UserID    string    `gorm:"index;type:char(36);not null"`
	TokenHash string    `gorm:"uniqueIndex;size:255;not null"`
	ExpiresAt time.Time `gorm:"not null"`
	Revoked   bool      `gorm:"not null;default:false"`
	CreatedAt time.Time
}

// ─── API Key ──────────────────────────────────────────────────────

type ApiKeyModel struct {
	BaseModel
	UserID     string `gorm:"index;type:char(36);not null"`
	Name       string `gorm:"size:128;not null"`
	KeyPrefix  string `gorm:"uniqueIndex;size:32;not null"`
	KeyHash    string `gorm:"size:255;not null"`
	LastUsedAt *time.Time
	ExpiresAt  *time.Time
	Revoked    bool `gorm:"not null;default:false"`
}
