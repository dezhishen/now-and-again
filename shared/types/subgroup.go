package types

import "github.com/google/uuid"

// ─── SubGroup ─────────────────────────────────────────────────────

type SubGroup struct {
	ID          uuid.UUID `json:"id"`
	FamilyID    uuid.UUID `json:"family_id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedBy   uuid.UUID `json:"created_by"`
	Timestamps
}

type CreateSubGroupRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=128"`
	Description string `json:"description,omitempty"`
}

// ─── SubGroupMember ───────────────────────────────────────────────

type SubGroupMember struct {
	ID         uuid.UUID `json:"id"`
	SubGroupID uuid.UUID `json:"sub_group_id"`
	UserID     uuid.UUID `json:"user_id"`
	JoinedAt   string    `json:"joined_at"`
	// Expanded
	User *User `json:"user,omitempty"`
}

type AddSubGroupMemberRequest struct {
	UserID uuid.UUID `json:"user_id" binding:"required"`
}
