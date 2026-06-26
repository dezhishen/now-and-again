package types

import "time"

// ─── Family ───────────────────────────────────────────────────────

type Family struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	InviteCode string    `json:"invite_code"`
	CreatedBy  string    `json:"created_by"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type CreateFamilyRequest struct {
	Name string `json:"name" binding:"required"`
}

type JoinFamilyRequest struct {
	InviteCode string `json:"invite_code" binding:"required"`
}

type FamilyMember struct {
	ID       string       `json:"id"`
	FamilyID string       `json:"family_id"`
	UserID   string       `json:"user_id"`
	Role     FamilyRole   `json:"role"`
	Status   MemberStatus `json:"status"`
	JoinedAt time.Time    `json:"joined_at"`
	User     *User        `json:"user,omitempty"`
}

type UpdateMemberRoleRequest struct {
	Role FamilyRole `json:"role" binding:"required"`
}

type ReviewJoinRequest struct {
	UserID string       `json:"user_id" binding:"required"`
	Action MemberStatus `json:"action" binding:"required"` // "active" to approve, "rejected" to reject
}

// ─── Family Group ─────────────────────────────────────────────────

type FamilyGroup struct {
	ID          string    `json:"id"`
	FamilyID    string    `json:"family_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateFamilyGroupRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description,omitempty"`
}

type FamilyGroupMember struct {
	ID       string       `json:"id"`
	GroupID  string       `json:"group_id"`
	UserID   string       `json:"user_id"`
	Role     GroupRole    `json:"role"`
	Status   MemberStatus `json:"status"`
	JoinedAt time.Time    `json:"joined_at"`
	User     *User        `json:"user,omitempty"`
}

type ReviewGroupJoinRequest struct {
	UserID string       `json:"user_id" binding:"required"`
	Action MemberStatus `json:"action" binding:"required"`
}
