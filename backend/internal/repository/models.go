package repository

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
	DisplayName string `gorm:"size:128;not null"`
	Email       string `gorm:"uniqueIndex;size:255"`
	Phone       string `gorm:"size:20"`
	AvatarURL   string `gorm:"type:text"`
	Timezone    string `gorm:"size:64;not null;default:Asia/Shanghai"` // IANA timezone

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
	Name               string `gorm:"size:128;not null"`
	InviteCode         string `gorm:"uniqueIndex;size:32;not null"`
	CreatedBy          string `gorm:"type:char(36);not null"`
	Timezone           string `gorm:"size:64;not null;default:Asia/Shanghai"` // IANA timezone, used for schedule resolution
	FloorPlanImagePath string `gorm:"->"`                                     // populated by subquery in ListFamiliesByUserID
}

func (FamilyModel) TableName() string { return "families" }

type FamilyMemberModel struct {
	BaseModel
	FamilyID string `gorm:"uniqueIndex:idx_family_user;type:char(36);not null"`
	UserID   string `gorm:"uniqueIndex:idx_family_user;type:char(36);not null"`
	Role     string `gorm:"size:16;not null;default:member"`
	Status   string `gorm:"size:16;not null;default:active"`
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
	IsCover  bool       `gorm:"not null;default:false"`
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
	ParentTaskID   string           `gorm:"index;type:char(36)"`
	IsRoot         bool             `gorm:"not null;default:false;index:idx_root_family,priority:1"`
	Name           string           `gorm:"size:128;not null"`
	ScheduleType   string           `gorm:"size:32;not null"`   // once/daily/weekly/monthly/interval
	ScheduleData   string           `gorm:"type:text;not null"` // JSON config
	Enabled        bool             `gorm:"not null;default:true"`
	Kind           string           `gorm:"size:16;not null;default:simple"` // simple | inspection (future: chain)
	DisplaySummary string           `gorm:"size:256"`                        // plugin-populated display text for list view
	LastTodoAt     *time.Time
	CreatedBy      string `gorm:"type:char(36);not null"`
	// Relations
	CheckItems  []CheckItemModel `gorm:"foreignKey:TaskID"`
	Children    []TaskModel      `gorm:"foreignKey:ParentTaskID"`
	CheckItems_ string           `gorm:"-"` // ignore old field in DB, kept for migration
}

func (TaskModel) TableName() string { return "tasks" }

type TodoModel struct {
	BaseModel
	TaskID      string    `gorm:"index;type:char(36);not null"`
	FamilyID    string    `gorm:"index;type:char(36);not null"`
	LocationID  string    `gorm:"index;type:char(36)"`
	AssignedTo  string    `gorm:"index;type:char(36)"`
	Status      string    `gorm:"size:16;not null;default:pending"` // pending/done/skipped
	BranchName  string    `gorm:"size:128"`                         // selected branch (only for kind=branched)
	Remark      string    `gorm:"type:text"`                        // user note on completion
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
	TaskID     string `gorm:"index;type:char(36);not null"`
	TodoID     string `gorm:"index;type:char(36)"` // linked todo (for completion/skip logs)
	Status     string `gorm:"size:32;not null"`    // registered/triggered/created/completed/manual etc
	Message    string `gorm:"type:text"`
	LogType    string `gorm:"size:16;not null;default:system"` // system / user
	OperatorID string `gorm:"index;type:char(36)"`             // user who triggered (empty for system)
}

func (TaskLogModel) TableName() string { return "task_logs" }

// ─── Inspection Result ───────────────────────────────────────────

// InspectionResultModel stores each inspection submission for audit trail.
type InspectionResultModel struct {
	BaseModel
	TaskID     string `gorm:"index;type:char(36);not null"` // which inspection task
	TodoID     string `gorm:"index;type:char(36);not null"` // which todo was completed
	FamilyID   string `gorm:"index;type:char(36);not null"`
	ItemName   string `gorm:"size:128;not null"`            // check item name
	BranchName string `gorm:"size:128;not null"`            // selected branch
	Remark     string `gorm:"size:512"`                     // optional remark
	CreatedBy  string `gorm:"index;type:char(36);not null"` // who submitted
}

func (InspectionResultModel) TableName() string { return "inspection_results" }

// ─── Check Items (巡检检查项) ─────────────────────────────────────

type CheckItemModel struct {
	BaseModel
	TaskID    string                 `gorm:"index;type:char(36);not null"` // inspection task ID
	Name      string                 `gorm:"size:128;not null"`            // item name
	SortOrder int                    `gorm:"not null;default:0"`
	Branches  []CheckItemBranchModel `gorm:"foreignKey:CheckItemID"`
}

func (CheckItemModel) TableName() string { return "check_items" }

type CheckItemBranchModel struct {
	BaseModel
	CheckItemID  string     `gorm:"index;type:char(36);not null"` // parent check item
	Name         string     `gorm:"size:128;not null"`            // branch name (e.g. "正常", "缺失")
	CreateTodo   bool       `gorm:"not null;default:false"`       // should create follow-up?
	BranchTaskID string     `gorm:"index;type:char(36)"`          // linked task template (null if create_todo=false)
	SortOrder    int        `gorm:"not null;default:0"`
	BranchTask   *TaskModel `gorm:"foreignKey:BranchTaskID"`
}

func (CheckItemBranchModel) TableName() string { return "check_item_branches" }

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
