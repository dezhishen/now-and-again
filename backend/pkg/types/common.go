package types

import "time"

// ─── Enums ────────────────────────────────────────────────────────

type FamilyRole string

const (
	FamilyRoleOwner  FamilyRole = "owner"
	FamilyRoleAdmin  FamilyRole = "admin"
	FamilyRoleMember FamilyRole = "member"
)

type GroupRole string

const (
	GroupRoleOwner  GroupRole = "owner"
	GroupRoleMember GroupRole = "member"
)

type MemberStatus string

const (
	MemberStatusActive   MemberStatus = "active"
	MemberStatusPending  MemberStatus = "pending"
	MemberStatusRejected MemberStatus = "rejected"
)

// ─── Common ───────────────────────────────────────────────────────

type Timestamps struct {
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Pagination struct {
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// ─── Unified error types ────────────────────────────────────────

// ErrorCode identifies the category of an API error.
type ErrorCode string

const (
	ErrBadRequest   ErrorCode = "BAD_REQUEST"
	ErrValidation   ErrorCode = "VALIDATION_ERROR"
	ErrUnauthorized ErrorCode = "UNAUTHORIZED"
	ErrForbidden    ErrorCode = "FORBIDDEN"
	ErrNotFound     ErrorCode = "NOT_FOUND"
	ErrConflict     ErrorCode = "CONFLICT"
	ErrInternal     ErrorCode = "INTERNAL_ERROR"
)

// FieldError describes a single field-level validation failure.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// APIError is the unified error response body.
type APIError struct {
	Code    ErrorCode    `json:"code"`
	Summary string       `json:"summary"`
	Details []FieldError `json:"details,omitempty"`
}

type APIResponse[T any] struct {
	Success bool      `json:"success"`
	Data    *T        `json:"data,omitempty"`
	Error   *APIError `json:"error,omitempty"`
}

type PagedResponse[T any] struct {
	Success    bool       `json:"success"`
	Data       []T        `json:"data"`
	Pagination Pagination `json:"pagination"`
}
