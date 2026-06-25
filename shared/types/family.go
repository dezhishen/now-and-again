package types

import "github.com/google/uuid"

// ─── Family ───────────────────────────────────────────────────────

type Family struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	InviteCode string    `json:"invite_code"`
	CreatedBy  uuid.UUID `json:"created_by"`
	Timestamps
}

type CreateFamilyRequest struct {
	Name string `json:"name" binding:"required,min=2,max=128"`
}

type JoinFamilyRequest struct {
	InviteCode string `json:"invite_code" binding:"required"`
}

// ─── FamilyMember ─────────────────────────────────────────────────

type FamilyMember struct {
	ID       uuid.UUID  `json:"id"`
	FamilyID uuid.UUID  `json:"family_id"`
	UserID   uuid.UUID  `json:"user_id"`
	Role     FamilyRole `json:"role"`
	JoinedAt string     `json:"joined_at"`
	// Expanded fields (populated on read)
	User *User `json:"user,omitempty"`
}

type UpdateMemberRoleRequest struct {
	Role FamilyRole `json:"role" binding:"required"`
}
