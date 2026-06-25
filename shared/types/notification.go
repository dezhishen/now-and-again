package types

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ─── NotificationChannel ──────────────────────────────────────────

type NotificationChannel struct {
	ID               uuid.UUID       `json:"id"`
	Code             string          `json:"code"`
	Name             string          `json:"name"`
	Description      string          `json:"description,omitempty"`
	Config           json.RawMessage `json:"config"`
	RateLimitPerHour *int            `json:"rate_limit_per_hour,omitempty"`
	IsActive         bool            `json:"is_active"`
	Timestamps
}

// ─── NotificationTemplate ─────────────────────────────────────────

type NotificationTemplate struct {
	ID           uuid.UUID  `json:"id"`
	FamilyID     *uuid.UUID `json:"family_id,omitempty"`    // null = system-level
	ScheduleTypeID   *uuid.UUID `json:"task_type_id,omitempty"` // null = global fallback
	TriggerEvent string     `json:"trigger_event"`
	ChannelCode  string     `json:"channel_code"`
	TitleTmpl    string     `json:"title_tmpl"`
	BodyTmpl     string     `json:"body_tmpl"`
	IsActive     bool       `json:"is_active"`
}

type UpsertTemplateRequest struct {
	TriggerEvent string     `json:"trigger_event" binding:"required"`
	ChannelCode  string     `json:"channel_code" binding:"required"`
	TitleTmpl    string     `json:"title_tmpl" binding:"required"`
	BodyTmpl     string     `json:"body_tmpl" binding:"required"`
	ScheduleTypeID   *uuid.UUID `json:"task_type_id,omitempty"`
}

// ─── UserChannelConfig ────────────────────────────────────────────

type UserChannelConfig struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	ChannelCode string    `json:"channel_code"`
	Destination string    `json:"destination"`
	IsEnabled   bool      `json:"is_enabled"`
	QuietStart  *string   `json:"quiet_start,omitempty"` // "22:00"
	QuietEnd    *string   `json:"quiet_end,omitempty"`   // "08:00"
	IsVerified  bool      `json:"is_verified"`
}

type UpsertUserChannelRequest struct {
	ChannelCode string  `json:"channel_code" binding:"required"`
	Destination string  `json:"destination" binding:"required"`
	IsEnabled   bool    `json:"is_enabled"`
	QuietStart  *string `json:"quiet_start,omitempty"`
	QuietEnd    *string `json:"quiet_end,omitempty"`
}

// ─── Notification ─────────────────────────────────────────────────

type Notification struct {
	ID           uuid.UUID          `json:"id"`
	UserID       uuid.UUID          `json:"user_id"`
	TaskID       uuid.UUID          `json:"task_id"`
	ChannelCode  string             `json:"channel_code"`
	TriggerEvent string             `json:"trigger_event"`
	Title        string             `json:"title"`
	Body         string             `json:"body"`
	Status       NotificationStatus `json:"status"`
	ErrorMsg     string             `json:"error_msg,omitempty"`
	SentAt       *time.Time         `json:"sent_at,omitempty"`
	CreatedAt    string             `json:"created_at"`
}
