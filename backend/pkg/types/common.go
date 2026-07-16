package types

import (
	"errors"
	"time"
)

// ─── Sentinel errors ──────────────────────────────────────────────

// ErrRefreshTokenInvalid is returned when a refresh token is not found,
// has been revoked, or has expired. Callers should respond with 401.
var ErrRefreshTokenInvalid = errors.New("refresh token is invalid or expired")

// ErrInvalidInviteCode is returned when a family invite code does not match
// any family. Callers should respond with 400.
var ErrInvalidInviteCode = errors.New("invalid invite code")

// ─── TriState ────────────────────────────────────────────────────

// TriState is a three-state string for optional boolean filters.
//
//	"" (Unset)  = no filter (show all)
//	"true"      = only true
//	"false"     = only false
type TriState string

const (
	TriStateUnset TriState = ""      // no filter
	TriStateTrue  TriState = "true"  // only true
	TriStateFalse TriState = "false" // only false
)

// ParseTriState parses a query param into a TriState.
// Returns Unset if the param is absent.
func ParseTriState(v string, present bool) TriState {
	if !present {
		return TriStateUnset
	}
	switch v {
	case "true":
		return TriStateTrue
	case "false":
		return TriStateFalse
	default:
		return TriStateUnset
	}
}

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

type TodoStatus string

const (
	TodoStatusPending     TodoStatus = "pending"
	TodoStatusDone        TodoStatus = "done"
	TodoStatusSkipped     TodoStatus = "skipped"
	TodoStatusInterrupted TodoStatus = "interrupted"
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
	ErrBadRequest     ErrorCode = "BAD_REQUEST"
	ErrValidation     ErrorCode = "VALIDATION_ERROR"
	ErrUnauthorized   ErrorCode = "UNAUTHORIZED"
	ErrForbidden      ErrorCode = "FORBIDDEN"
	ErrNotFound       ErrorCode = "NOT_FOUND"
	ErrConflict       ErrorCode = "CONFLICT"
	ErrInternal       ErrorCode = "INTERNAL_ERROR"
	ErrFamilyNotFound ErrorCode = "FAMILY_NOT_FOUND"
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
