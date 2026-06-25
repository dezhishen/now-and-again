package repository

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ─── Base Model ───────────────────────────────────────────────────
// BaseModel generates UUID in Go (works with both SQLite and PostgreSQL).

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

// ─── User ─────────────────────────────────────────────────────────

type UserModel struct {
	BaseModel
	Username     string `gorm:"uniqueIndex;size:64;not null"`
	Email        string `gorm:"uniqueIndex;size:255;not null"`
	Phone        string `gorm:"size:20"`
	PasswordHash string `gorm:"size:255;not null"`
	DisplayName  string `gorm:"size:128;not null"`
	AvatarURL    string `gorm:"type:text"`
	IsAdmin      bool   `gorm:"not null;default:false"`
}

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
	JoinedAt time.Time
	User     UserModel `gorm:"foreignKey:UserID"`
}

// ─── SubGroup ─────────────────────────────────────────────────────

type SubGroupModel struct {
	BaseModel
	FamilyID    string `gorm:"index;type:char(36);not null"`
	Name        string `gorm:"size:128;not null"`
	Description string `gorm:"type:text"`
	CreatedBy   string `gorm:"type:char(36);not null"`
}

type SubGroupMemberModel struct {
	BaseModel
	SubGroupID string `gorm:"uniqueIndex:idx_subgroup_user;type:char(36);not null"`
	UserID     string `gorm:"uniqueIndex:idx_subgroup_user;type:char(36);not null"`
	JoinedAt   time.Time
}

type TaskModel struct {
	BaseModel
	FamilyID         string     `gorm:"index:idx_task_family_status;type:char(36);not null"`
	SubGroupID       *string    `gorm:"type:char(36)"`
	TaskCode string `gorm:"size:64;not null"`
	ChainID          *string    `gorm:"type:char(36)"`
	Title            string     `gorm:"size:255;not null"`
	Description      string     `gorm:"type:text"`
	Status           string     `gorm:"index:idx_task_family_status;size:16;not null;default:todo"`
	Priority         string     `gorm:"size:16;not null;default:medium"`
	DueDate          *time.Time `gorm:"index:idx_task_due"`
	RecurrenceConfig string     `gorm:"type:text"`
	TypeSpecificData string     `gorm:"type:text"`
	CreatedBy        string     `gorm:"type:char(36);not null"`
	CompletedAt      *time.Time

	Assignees []TaskAssigneeModel `gorm:"foreignKey:TaskID"`
}

func (TaskModel) TableName() string { return "tasks" }

type TaskAssigneeModel struct {
	BaseModel
	TaskID     string `gorm:"uniqueIndex:idx_task_user;type:char(36);not null"`
	UserID     string `gorm:"uniqueIndex:idx_task_user;type:char(36);not null"`
	AssignedAt time.Time
	User       UserModel `gorm:"foreignKey:UserID"`
}

type TaskDependencyModel struct {
	BaseModel
	BlockedTaskID  string `gorm:"uniqueIndex:idx_dep_pair;type:char(36);not null"`
	BlockerTaskID  string `gorm:"uniqueIndex:idx_dep_pair;type:char(36);not null"`
	DependencyType string `gorm:"size:16;not null;default:blocks"`
}

// ─── Task Chain ───────────────────────────────────────────────────

type TaskChainModel struct {
	BaseModel
	FamilyID    string `gorm:"index;type:char(36);not null"`
	Name        string `gorm:"size:128;not null"`
	Description string `gorm:"type:text"`
	Icon        string `gorm:"size:64"`
	IsActive    bool   `gorm:"not null;default:true"`
	CreatedBy   string `gorm:"type:char(36);not null"`

	Steps []TaskChainStepModel `gorm:"foreignKey:ChainID"`
}

type TaskChainStepModel struct {
	BaseModel
	ChainID            string  `gorm:"index;type:char(36);not null"`
	SortOrder          int     `gorm:"not null;default:0"`
	Title              string  `gorm:"size:255;not null"`
	Description        string  `gorm:"type:text"`
	TaskCode string `gorm:"size:64;not null"`
	AssignedRole       string  `gorm:"size:16;not null;default:any"`
	AssignedSubGroupID *string `gorm:"type:char(36)"`
	DelayAfterPrevious string  `gorm:"size:16;not null;default:0h"`
	IsOptional         bool    `gorm:"not null;default:false"`
	Priority           string  `gorm:"size:16;not null;default:medium"`

}

// ─── Task Log ─────────────────────────────────────────────────────

type TaskLogModel struct {
	BaseModel
	TaskID string    `gorm:"index:idx_log_task;type:char(36);not null"`
	UserID string    `gorm:"type:char(36);not null"`
	Action string    `gorm:"size:32;not null"`
	Detail string    `gorm:"type:text"`
	User   UserModel `gorm:"foreignKey:UserID"`
}

// ─── Inspection ───────────────────────────────────────────────────

type InspectionModel struct {
	BaseModel
	FamilyID    string `gorm:"index;type:char(36);not null"`
	Title       string `gorm:"size:255;not null"`
	Description string `gorm:"type:text"`
	Status      string `gorm:"size:16;not null;default:in_progress"`
	CreatedBy   string `gorm:"type:char(36);not null"`
	CompletedAt *time.Time

	Items []InspectionItemModel `gorm:"foreignKey:InspectionID"`
	User  UserModel             `gorm:"foreignKey:CreatedBy"`
}

type InspectionItemModel struct {
	BaseModel
	InspectionID    string  `gorm:"index;type:char(36);not null"`
	CheckPoint      string  `gorm:"size:255;not null"`
	Result          string  `gorm:"size:16;not null;default:ok"`
	Note            string  `gorm:"type:text"`
	GeneratedTaskID *string `gorm:"type:char(36)"`
	CheckedAt       *time.Time
}

// ─── Notification ─────────────────────────────────────────────────

type NotificationChannelModel struct {
	BaseModel
	Code             string `gorm:"uniqueIndex;size:32;not null"`
	Name             string `gorm:"size:64;not null"`
	Description      string `gorm:"type:text"`
	Config           string `gorm:"type:text"`
	RateLimitPerHour *int
	IsActive         bool `gorm:"not null;default:true"`
}

type NotificationTemplateModel struct {
	BaseModel
	FamilyID     *string `gorm:"type:char(36)"`
	TaskCode   *string `gorm:"type:char(36)"`
	TriggerEvent string  `gorm:"size:64;not null"`
	ChannelCode  string  `gorm:"size:32;not null"`
	TitleTmpl    string  `gorm:"size:255;not null"`
	BodyTmpl     string  `gorm:"type:text;not null"`
	IsActive     bool    `gorm:"not null;default:true"`
}

type UserChannelConfigModel struct {
	BaseModel
	UserID      string  `gorm:"uniqueIndex:idx_user_channel;type:char(36);not null"`
	ChannelCode string  `gorm:"uniqueIndex:idx_user_channel;size:32;not null"`
	Destination string  `gorm:"size:255;not null"`
	IsEnabled   bool    `gorm:"not null;default:true"`
	QuietStart  *string `gorm:"size:8"`
	QuietEnd    *string `gorm:"size:8"`
	IsVerified  bool    `gorm:"not null;default:false"`
}

type NotificationModel struct {
	BaseModel
	UserID       string `gorm:"index:idx_notif_user;type:char(36);not null"`
	TaskID       string `gorm:"type:char(36);not null"`
	ChannelCode  string `gorm:"size:32;not null"`
	TriggerEvent string `gorm:"size:64;not null"`
	Title        string `gorm:"size:255;not null"`
	Body         string `gorm:"type:text;not null"`
	Status       string `gorm:"size:16;not null;default:pending"`
	ErrorMsg     string `gorm:"type:text"`
	SentAt       *time.Time
}

// ─── Auth ──────────────────────────────────────────────────────────

type RefreshTokenModel struct {
	BaseModel
	UserID    string    `gorm:"index;type:char(36);not null"`
	TokenHash string    `gorm:"uniqueIndex;size:64;not null"`
	ExpiresAt time.Time `gorm:"not null"`
	Revoked   bool      `gorm:"not null;default:false"`
}

// ─── API Key ──────────────────────────────────────────────────────

type ApiKeyModel struct {
	BaseModel
	UserID     string     `gorm:"index;type:char(36);not null"`
	Name       string     `gorm:"size:128;not null"`
	KeyHash    string     `gorm:"uniqueIndex;size:64;not null"`
	KeyPrefix  string     `gorm:"size:12;not null"` // first 8 chars for display
	Scopes     string     `gorm:"type:text"`        // JSON array of scopes, null = full access
	LastUsedAt *time.Time
	ExpiresAt  *time.Time
	Revoked    bool       `gorm:"not null;default:false"`
}
