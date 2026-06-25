// Package types defines shared data structures used by both backend and CLI.
// These types serve as the contract between the API server and its clients.
package types

import "time"

// Pagination wraps list responses.
type Pagination struct {
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// APIResponse is the standard API envelope.
type APIResponse[T any] struct {
	Success bool   `json:"success"`
	Data    T      `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

// PagedResponse is the standard paginated envelope.
type PagedResponse[T any] struct {
	Success    bool       `json:"success"`
	Data       []T        `json:"data"`
	Pagination Pagination `json:"pagination"`
	Error      string     `json:"error,omitempty"`
}

// ─── Enums ────────────────────────────────────────────────────────

// TaskStatus represents the lifecycle of a task.
type TaskStatus string

const (
	TaskStatusTodo       TaskStatus = "todo"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusDone       TaskStatus = "done"
	TaskStatusBlocked    TaskStatus = "blocked"
	TaskStatusArchived   TaskStatus = "archived"
)

// TaskCategory discriminates one-off vs recurring tasks.
type TaskCategory string

const (
	TaskCategoryNow   TaskCategory = "now"
	TaskCategoryAgain TaskCategory = "again"
)

// Priority levels.
type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
	PriorityUrgent Priority = "urgent"
)

// FamilyRole defines permissions within a family.
type FamilyRole string

const (
	FamilyRoleOwner  FamilyRole = "owner"
	FamilyRoleAdmin  FamilyRole = "admin"
	FamilyRoleMember FamilyRole = "member"
)

// AssignRole for chain steps.
type AssignRole string

const (
	AssignRoleAny        AssignRole = "any"
	AssignRoleChainOwner AssignRole = "chain_owner"
	AssignRoleSubGroup   AssignRole = "sub_group"
)

// DependencyType for task dependencies.
type DependencyType string

const (
	DependencyBlocks  DependencyType = "blocks"
	DependencyRelates DependencyType = "relates_to"
)

// InspectionResult for inspection items.
type InspectionResult string

const (
	InspectionOK    InspectionResult = "ok"
	InspectionIssue InspectionResult = "issue_found"
)

// NotificationStatus for delivery tracking.
type NotificationStatus string

const (
	NotifPending NotificationStatus = "pending"
	NotifSent    NotificationStatus = "sent"
	NotifFailed  NotificationStatus = "failed"
	NotifSkipped NotificationStatus = "skipped"
)

// ─── Trigger Events ───────────────────────────────────────────────

type TriggerEvent string

const (
	EventTaskCreated     TriggerEvent = "task_created"
	EventTaskAssigned    TriggerEvent = "task_assigned"
	EventChainStarted    TriggerEvent = "task_chain_started"
	EventTaskUnblocked   TriggerEvent = "task_unblocked"
	EventTaskDueSoon     TriggerEvent = "task_due_soon"
	EventTaskOverdue     TriggerEvent = "task_overdue"
	EventTaskCompleted   TriggerEvent = "task_completed"
	EventInspectionIssue TriggerEvent = "inspection_issue"
	EventDailyDigest     TriggerEvent = "daily_digest"
	EventWeeklyReport    TriggerEvent = "weekly_report"
)

// ValidTaskStatuses is the canonical set of mutable statuses.
var ValidTaskStatuses = []TaskStatus{
	TaskStatusTodo, TaskStatusInProgress, TaskStatusDone, TaskStatusBlocked,
}

// IsTerminal returns true if the status represents a final state.
func (s TaskStatus) IsTerminal() bool {
	return s == TaskStatusDone || s == TaskStatusArchived
}

// Timestamps is embedded in every entity.
type Timestamps struct {
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
