package types

import "time"

// ─── User ─────────────────────────────────────────────────────────

type User struct {
	ID          string    `json:"id"`
	DisplayName string    `json:"display_name"`
	Email       string    `json:"email"`
	Phone       string    `json:"phone"`
	AvatarURL   string    `json:"avatar_url"`
	Roles       []string  `json:"roles"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ─── Account ──────────────────────────────────────────────────────

type Account struct {
	ID                string    `json:"id"`
	UserID            string    `json:"user_id"`
	Provider          string    `json:"provider"`
	ProviderAccountID string    `json:"provider_account_id,omitempty"`
	Username          string    `json:"username,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
}

// ─── Auth ─────────────────────────────────────────────────────────

type CreateUserRequest struct {
	DisplayName string `json:"display_name" binding:"required"`
	Username    string `json:"username" binding:"required,min=3"`
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=8"`
	Phone       string `json:"phone,omitempty"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	User         *User  `json:"user"`
}

type UpdateUserRequest struct {
	DisplayName *string `json:"display_name,omitempty"`
	Email       *string `json:"email,omitempty"`
	Phone       *string `json:"phone,omitempty"`
	AvatarURL   *string `json:"avatar_url,omitempty"`
}

// ─── API Key ──────────────────────────────────────────────────────

type ApiKey struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	KeyPrefix  string     `json:"key_prefix"`
	RawKey     string     `json:"raw_key,omitempty"`
	Scopes     []string   `json:"scopes,omitempty"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

type CreateApiKeyRequest struct {
	Name      string     `json:"name" binding:"required"`
	Scopes    []string   `json:"scopes,omitempty"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

type CreateApiKeyResponse struct {
	ApiKey  *ApiKey `json:"api_key"`
	Message string  `json:"message"`
}
