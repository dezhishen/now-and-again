package types

import "github.com/google/uuid"

// ─── User ─────────────────────────────────────────────────────────

type User struct {
	ID          uuid.UUID `json:"id"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	Phone       string    `json:"phone,omitempty"`
	DisplayName string    `json:"display_name"`
	AvatarURL   string    `json:"avatar_url,omitempty"`
	IsAdmin     bool      `json:"is_admin"`
	Timestamps
}

type CreateUserRequest struct {
	Username    string `json:"username" binding:"required,min=3,max=64"`
	Email       string `json:"email" binding:"required,email"`
	Phone       string `json:"phone,omitempty"`
	Password    string `json:"password" binding:"required,min=8"`
	DisplayName string `json:"display_name" binding:"required"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
	User      User   `json:"user"`
}

type UpdateUserRequest struct {
	DisplayName *string `json:"display_name,omitempty"`
	Email       *string `json:"email,omitempty"`
	Phone       *string `json:"phone,omitempty"`
	AvatarURL   *string `json:"avatar_url,omitempty"`
}

// ─── System Setup ─────────────────────────────────────────────────

type SetupRequest struct {
	Username    string `json:"username" binding:"required,min=3,max=64"`
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=8"`
	DisplayName string `json:"display_name" binding:"required"`
}

type SystemStatus struct {
	Initialized bool `json:"initialized"`
}

// ─── Auth Tokens ──────────────────────────────────────────────────

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"` // seconds until access token expires
	User         User   `json:"user"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// ─── API Key ──────────────────────────────────────────────────────

type ApiKey struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	KeyPrefix  string  `json:"key_prefix"`
	Scopes     *string `json:"scopes,omitempty"`
	LastUsedAt *string `json:"last_used_at,omitempty"`
	ExpiresAt  *string `json:"expires_at,omitempty"`
	Revoked    bool    `json:"revoked"`
	CreatedAt  string  `json:"created_at"`
}

type CreateApiKeyRequest struct {
	Name      string  `json:"name" binding:"required"`
	ExpiresAt *string `json:"expires_at,omitempty"`
}

type CreateApiKeyResponse struct {
	ApiKey     ApiKey `json:"api_key"`
	RawKey     string `json:"raw_key"` // only returned once!
}
