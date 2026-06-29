package types

import "time"

// ─── Task Template ────────────────────────────────────────────────

// TemplateParameter defines a single dynamic parameter for a task template.
type TemplateParameter struct {
	Key         string         `json:"key"`
	Label       string         `json:"label"`
	Type        string         `json:"type"` // string, int, float, bool, select
	Description string         `json:"description,omitempty"`
	Required    bool           `json:"required"`
	Default     any            `json:"default,omitempty"`
	Options     []SelectOption `json:"options,omitempty"`
	Placeholder string         `json:"placeholder,omitempty"`
}

// SelectOption is an option for "select" type parameters.
type SelectOption struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

// TaskTemplate is the API DTO for a task template.
type TaskTemplate struct {
	ID           string              `json:"id"`
	FamilyID     *string             `json:"family_id,omitempty"` // nil = system-level
	ProviderCode string              `json:"provider_code"`
	TemplateCode string              `json:"template_code"`
	Name         string              `json:"name"`
	Description  string              `json:"description,omitempty"`
	Kind         string              `json:"kind"`
	Icon         string              `json:"icon,omitempty"`
	SortOrder    int                 `json:"sort_order"`
	Enabled      bool                `json:"enabled"`
	Parameters   []TemplateParameter `json:"parameters,omitempty"`
	TaskDefaults any                 `json:"task_defaults,omitempty"`
	ExtraSchema  any                 `json:"extra_schema,omitempty"`
	Version      string              `json:"version,omitempty"`
	CreatedAt    time.Time           `json:"created_at"`
	UpdatedAt    time.Time           `json:"updated_at"`
}

// CreateTaskTemplateRequest is used by family owners to create a family-level template.
type CreateTaskTemplateRequest struct {
	TemplateCode string              `json:"template_code" binding:"required"`
	Name         string              `json:"name" binding:"required"`
	Description  string              `json:"description,omitempty"`
	Kind         string              `json:"kind" binding:"required"`
	Icon         string              `json:"icon,omitempty"`
	SortOrder    int                 `json:"sort_order,omitempty"`
	Enabled      bool                `json:"enabled"`
	Parameters   []TemplateParameter `json:"parameters,omitempty"`
	TaskDefaults any                 `json:"task_defaults,omitempty"`
	ExtraSchema  any                 `json:"extra_schema,omitempty"`
}

// UpdateTaskTemplateRequest is used by family owners to update a family-level template.
type UpdateTaskTemplateRequest struct {
	Name         *string             `json:"name,omitempty"`
	Description  *string             `json:"description,omitempty"`
	Kind         *string             `json:"kind,omitempty"`
	Icon         *string             `json:"icon,omitempty"`
	SortOrder    *int                `json:"sort_order,omitempty"`
	Enabled      *bool               `json:"enabled,omitempty"`
	Parameters   []TemplateParameter `json:"parameters,omitempty"`
	TaskDefaults any                 `json:"task_defaults,omitempty"`
	ExtraSchema  any                 `json:"extra_schema,omitempty"`
}

// TemplateProvider is the API DTO describing a registered provider.
type TemplateProvider struct {
	Code        string     `json:"code"`
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	LastSyncAt  *time.Time `json:"last_sync_at,omitempty"`
	SyncStatus  string     `json:"sync_status"`
}

// RenderedTask is the output of rendering a template with user-supplied parameters.
type RenderedTask struct {
	TaskDefaults any `json:"task_defaults"`
	ExtraSchema  any `json:"extra_schema,omitempty"`
}

// ─── Task Template Subscription ───────────────────────────────────

// TaskTemplateSubscription is the API DTO for a subscription.
type TaskTemplateSubscription struct {
	ID                   string    `json:"id"`
	FamilyID             *string   `json:"family_id,omitempty"` // nil = system-level
	ProviderCode         string    `json:"provider_code"`
	URL                  string    `json:"url"`
	Name                 string    `json:"name"`
	AutoRefresh          bool      `json:"auto_refresh"`
	RefreshIntervalHours int       `json:"refresh_interval_hours"`
	Enabled              bool      `json:"enabled"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// CreateSubscriptionRequest is used to create a subscription.
type CreateSubscriptionRequest struct {
	ProviderCode         string `json:"provider_code" binding:"required"`
	URL                  string `json:"url" binding:"required"`
	Name                 string `json:"name" binding:"required"`
	AutoRefresh          bool   `json:"auto_refresh"`
	RefreshIntervalHours int    `json:"refresh_interval_hours"`
}

// UpdateSubscriptionRequest is used to update a subscription.
type UpdateSubscriptionRequest struct {
	URL                  *string `json:"url,omitempty"`
	Name                 *string `json:"name,omitempty"`
	AutoRefresh          *bool   `json:"auto_refresh,omitempty"`
	RefreshIntervalHours *int    `json:"refresh_interval_hours,omitempty"`
	Enabled              *bool   `json:"enabled,omitempty"`
}
