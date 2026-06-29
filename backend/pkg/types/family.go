package types

import "time"

// ─── Family ───────────────────────────────────────────────────────

type Family struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	InviteCode   string    `json:"invite_code"`
	CreatedBy    string    `json:"created_by"`
	Archived     bool      `json:"archived"`
	ThumbnailURL string    `json:"thumbnail_url,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CreateFamilyRequest struct {
	Name string `json:"name" binding:"required,max=128"`
}

type JoinFamilyRequest struct {
	InviteCode string `json:"invite_code" binding:"required,max=32"`
}

type UpdateFamilyRequest struct {
	Name string `json:"name" binding:"required,max=128"`
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
	Name        string `json:"name" binding:"required,max=128"`
	Description string `json:"description,omitempty" binding:"max=512"`
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

// ─── Floor Plan ──────────────────────────────────────────────────

type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type FloorPlan struct {
	ID        string     `json:"id"`
	FamilyID  string     `json:"family_id"`
	Label     string     `json:"label"`
	ImageURL  string     `json:"image_url"`
	IsCover   bool       `json:"is_cover"`
	Width     int        `json:"width"`
	Height    int        `json:"height"`
	Locations []Location `json:"locations,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type CreateFloorPlanRequest struct {
	Label string `json:"label"`
}

type Location struct {
	ID          string    `json:"id"`
	FamilyID    string    `json:"family_id"`
	FloorPlanID *string   `json:"floor_plan_id,omitempty"`
	Kind        string    `json:"kind"`
	Name        string    `json:"name"`
	Color       string    `json:"color"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateLocationRequest struct {
	Name        string `json:"name" binding:"required"`
	Kind        string `json:"kind"`
	FloorPlanID string `json:"floor_plan_id,omitempty"`
	Color       string `json:"color"`
}

type UpdateLocationRequest struct {
	Name        string  `json:"name"`
	Kind        string  `json:"kind"`
	FloorPlanID *string `json:"floor_plan_id,omitempty"`
	Color       string  `json:"color"`
}
