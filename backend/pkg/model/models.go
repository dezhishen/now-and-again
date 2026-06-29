package model

import (
	"time"

	"github.com/dezhishen/now-and-again/backend/pkg/timeutil"
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
	now := timeutil.Now()
	b.CreatedAt = now
	b.UpdatedAt = now
	return nil
}

func (b *BaseModel) BeforeUpdate(_ *gorm.DB) error {
	b.UpdatedAt = timeutil.Now()
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
	DisplayName     string  `gorm:"size:128;not null"`
	Email           string  `gorm:"uniqueIndex;size:255"`
	Phone           string  `gorm:"size:20"`
	AvatarURL       string  `gorm:"type:text"`
	Timezone        string  `gorm:"size:64;not null;default:Asia/Shanghai"`
	DefaultFamilyID *string `gorm:"type:char(36)"`

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
	Name               string              `gorm:"size:128;not null"`
	InviteCode         string              `gorm:"uniqueIndex;size:32;not null"`
	CreatedBy          string              `gorm:"type:char(36);not null"`
	Timezone           string              `gorm:"size:64;not null;default:Asia/Shanghai"`
	Archived           bool                `gorm:"not null;default:false"`
	FloorPlanImagePath string              `gorm:"->"`
	Members            []FamilyMemberModel `gorm:"foreignKey:FamilyID"`
}

func (FamilyModel) TableName() string { return "families" }

type FamilyMemberModel struct {
	BaseModel
	FamilyID string `gorm:"uniqueIndex:idx_family_user;index:idx_family_status,priority:1;type:char(36);not null"`
	UserID   string `gorm:"uniqueIndex:idx_family_user;index:idx_user_status,priority:1;type:char(36);not null"`
	Role     string `gorm:"size:16;not null;default:member"`
	Status   string `gorm:"index:idx_family_status,priority:2;index:idx_user_status,priority:2;size:16;not null;default:active"`
	JoinedAt time.Time
	User     UserModel `gorm:"foreignKey:UserID"`
}

func (FamilyMemberModel) TableName() string { return "family_members" }

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
	GroupID  string `gorm:"uniqueIndex:idx_group_user;index:idx_group_status,priority:1;type:char(36);not null"`
	UserID   string `gorm:"uniqueIndex:idx_group_user;type:char(36);not null"`
	Role     string `gorm:"size:16;not null;default:member"`
	Status   string `gorm:"index:idx_group_status,priority:2;size:16;not null;default:pending"`
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
	Scopes     string `gorm:"type:text"` // JSON array of scope strings, e.g. ["family:read"]
	LastUsedAt *time.Time
	ExpiresAt  *time.Time
	Revoked    bool `gorm:"not null;default:false"`
}

// ─── Image ──────────────────────────────────────────────────────

type ImageModel struct {
	BaseModel
	StorageType  string `gorm:"size:32;not null;default:local"`
	FilePath     string `gorm:"type:text;not null"`
	OriginalName string `gorm:"size:255"`
	MimeType     string `gorm:"size:64"`
	Size         int64  `gorm:"not null;default:0"`
}

func (ImageModel) TableName() string { return "images" }

// ─── System Settings ────────────────────────────────────────────

type SystemSettingModel struct {
	Key   string `gorm:"primaryKey;size:128"`
	Value string `gorm:"type:text;not null"`
}

func (SystemSettingModel) TableName() string { return "system_settings" }

// ─── Floor Plan ──────────────────────────────────────────────────

type FloorPlanModel struct {
	BaseModel
	FamilyID string     `gorm:"index;type:char(36);not null"`
	Label    string     `gorm:"size:32;not null;default:'1F'"`
	ImageID  string     `gorm:"index;type:char(36);not null"`
	IsCover  bool       `gorm:"index;not null;default:false"`
	Width    int        `gorm:"not null;default:0"`
	Height   int        `gorm:"not null;default:0"`
	Image    ImageModel `gorm:"foreignKey:ImageID"`
}

func (FloorPlanModel) TableName() string { return "floor_plans" }

type LocationModel struct {
	BaseModel
	FamilyID    string  `gorm:"index;type:char(36);not null"`
	FloorPlanID *string `gorm:"index;type:char(36)"`
	Kind        string  `gorm:"size:32;not null;default:'indoor'"`
	Name        string  `gorm:"size:128;not null"`
	Color       string  `gorm:"size:16;not null;default:'#3b82f6'"`
}

func (LocationModel) TableName() string { return "locations" }

// ─── Task ────────────────────────────────────────────────────────

type TaskModel struct {
	BaseModel
	FamilyID       string           `gorm:"index:idx_root_family,priority:2;index;type:char(36);not null"`
	GroupID        string           `gorm:"index;type:char(36)"`
	Group          FamilyGroupModel `gorm:"foreignKey:GroupID"`
	LocationID     string           `gorm:"index;type:char(36)"`
	ParentTaskID   string           `gorm:"index:idx_parent_root,priority:1;type:char(36)"`
	RootTaskID     string           `gorm:"index;type:char(36)"` // root ancestor of the task tree, for log aggregation
	IsRoot         bool             `gorm:"not null;default:false;index:idx_root_family,priority:1;index:idx_parent_root,priority:2"`
	Name           string           `gorm:"size:128;not null"`
	ScheduleType   string           `gorm:"size:32;not null"`   // once/daily/weekly/monthly/interval
	ScheduleData   string           `gorm:"type:text;not null"` // JSON config
	Enabled        bool             `gorm:"not null;default:true;index:idx_enabled_archived,priority:1"`
	Kind           string           `gorm:"size:16;not null;default:simple"` // simple | inspection (future: chain)
	DisplaySummary string           `gorm:"size:256"`                        // plugin-populated display text for list view
	Archived       bool             `gorm:"not null;default:false;index:idx_enabled_archived,priority:2"`
	LastTodoAt     *time.Time
	CreatedBy      string `gorm:"type:char(36);not null"`
}

func (TaskModel) TableName() string { return "tasks" }

type TodoModel struct {
	BaseModel
	TaskID      string    `gorm:"index;type:char(36);not null"`
	FamilyID    string    `gorm:"index;type:char(36);not null"`
	LocationID  string    `gorm:"index;type:char(36)"`
	AssignedTo  string    `gorm:"index;type:char(36)"`
	Status      string    `gorm:"index;size:16;not null;default:pending"` // pending/done/skipped
	Remark      string    `gorm:"type:text"`                              // user note on completion
	DueStart    time.Time `gorm:"not null"`
	DueDate     time.Time `gorm:"not null"`
	CompletedAt *time.Time
	CompletedBy string    `gorm:"type:char(36)"`
	Task        TaskModel `gorm:"foreignKey:TaskID"`
	User        UserModel `gorm:"foreignKey:AssignedTo"`
}

func (TodoModel) TableName() string { return "todos" }

// ─── Task Log ────────────────────────────────────────────────────

type TaskLogModel struct {
	BaseModel
	TaskID     string    `gorm:"index;type:char(36);not null"`
	TodoID     string    `gorm:"index;type:char(36)"`
	Status     string    `gorm:"size:32;not null"`
	Message    string    `gorm:"type:text"`
	LogType    string    `gorm:"index;size:16;not null;default:system"`
	OperatorID string    `gorm:"index;type:char(36)"`
	Task       TaskModel `gorm:"foreignKey:TaskID"` // for model-driven JOINs
}

func (TaskLogModel) TableName() string { return "task_logs" }

// ─── ICS Feed ────────────────────────────────────────────────────

type IcsFeedModel struct {
	BaseModel
	FamilyID        string       `gorm:"index;type:char(36);not null"`
	Name            string       `gorm:"size:128;not null"`
	Description     string       `gorm:"size:512"`
	FilterDays      int          `gorm:"not null;default:7"`               // how many days ahead
	FilterGroupID   string       `gorm:"index;type:char(36)"`              // optional group filter
	FilterType      string       `gorm:"size:16;not null;default:all"`     // all / personal
	AuthType        string       `gorm:"size:16;not null;default:api_key"` // api_key / basic
	ApiKeyID        string       `gorm:"index;type:char(36)"`
	AppUsername     string       `gorm:"size:64"`
	AppPasswordHash string       `gorm:"size:255"`
	AccessToken     string       `gorm:"uniqueIndex;size:36;not null"` // URL token
	Enabled         bool         `gorm:"not null;default:true"`
	CreatedBy       string       `gorm:"type:char(36);not null"`
	ApiKey          *ApiKeyModel `gorm:"foreignKey:ApiKeyID"`
}

func (IcsFeedModel) TableName() string { return "ics_feeds" }

// ─── Task Template ───────────────────────────────────────────────

// TaskTemplateModel stores resolved task templates in the database.
// Providers (builtin, http, …) sync their YAML definitions into this table;
// the main flow reads only from here — never from providers directly.
//
// FamilyID is nil for system-level templates (managed by admin, visible to all
// families) and non-nil for family-level templates (managed by family owner).
type TaskTemplateModel struct {
	BaseModel
	FamilyID     *string `gorm:"index;type:char(36)"` // nil = system-level
	ProviderCode string  `gorm:"index:idx_provider_template,priority:1;size:32;not null"`
	TemplateCode string  `gorm:"index:idx_provider_template,priority:2;size:64;not null"`
	Name         string  `gorm:"size:128;not null"`
	Description  string  `gorm:"size:512"`
	Kind         string  `gorm:"size:16;not null;default:simple"` // task kind selector
	Icon         string  `gorm:"size:32"`
	SortOrder    int     `gorm:"not null;default:0"`
	Enabled      bool    `gorm:"not null;default:true"`

	// Parameters is a JSON array of parameter definitions.
	// Example: [{"key":"area","label":"区域","type":"string","required":true}]
	Parameters string `gorm:"type:text"`

	// TaskDefaults is a JSON object with default task field values.
	// Go-template {{.param_key}} placeholders are resolved at render time.
	TaskDefaults string `gorm:"type:text"`

	// ExtraSchema is a JSON object with task-kind-specific extra defaults.
	// The main flow passes it through as-is; the frontend renders the form.
	ExtraSchema string `gorm:"type:text"`

	// Version is the template version string from the provider.
	Version  string `gorm:"size:32"`
	Metadata string `gorm:"type:text"` // opaque provider metadata (e.g. source URL)
}

func (TaskTemplateModel) TableName() string { return "task_templates" }

// ─── Task Template Subscription ──────────────────────────────────

// TaskTemplateSubscriptionModel stores HTTP(S) subscription sources for task templates.
// FamilyID is nil for system-level subscriptions (admin-managed); non-nil for family-level.
type TaskTemplateSubscriptionModel struct {
	BaseModel
	FamilyID             *string `gorm:"index;type:char(36)"` // nil = system-level
	ProviderCode         string  `gorm:"index;size:32;not null"`
	URL                  string  `gorm:"size:2048;not null"`
	Name                 string  `gorm:"size:128;not null"`
	AutoRefresh          bool    `gorm:"not null;default:false"`
	RefreshIntervalHours int     `gorm:"not null;default:24"`
	Enabled              bool    `gorm:"not null;default:true"`
}

func (TaskTemplateSubscriptionModel) TableName() string { return "task_template_subscriptions" }
