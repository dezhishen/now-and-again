package tasktemplate

import (
	"context"
	"time"

	"github.com/dezhishen/now-and-again/backend/pkg/model"
	"gorm.io/gorm"
)

// TemplateStorage is the database abstraction injected into providers.
// Each provider calls UpsertTemplate / DeleteTemplate to sync its data into
// the common task_templates table.  Pattern follows taskkind.TaskStorage.
type TemplateStorage interface {
	UpsertTemplate(tmpl *model.TaskTemplateModel) error
	DeleteTemplate(providerCode, templateCode string) error
	FindByProvider(providerCode string) ([]*model.TaskTemplateModel, error)

	// ListSubscriptions returns the subscription URLs for a given provider.
	// The returned subscriptions are scoped to the storage instance (system or family).
	ListSubscriptions(providerCode string) ([]SubscriptionInfo, error)

	DB() *gorm.DB
}

// SubscriptionInfo describes a single subscription source.
type SubscriptionInfo struct {
	URL                  string
	Name                 string
	AutoRefresh          bool
	RefreshIntervalHours int
}

// Provider is the unified interface every task-template source must implement.
//
// The main flow MUST NOT type-assert or branch on a concrete provider type.
// All provider-specific behaviour is encapsulated inside Sync().
type Provider interface {
	// Code returns a unique provider identifier (e.g. "builtin", "http").
	Code() string

	// Name returns a human-readable label shown in the UI.
	Name() string

	// Description returns optional descriptive text.
	Description() string

	// Sync pulls the provider's templates and writes them into the database
	// via the supplied TemplateStorage.  Built-in providers read embedded
	// YAML; HTTP providers fetch remote URLs; both produce the same YAML
	// format and call the same storage methods.
	Sync(ctx context.Context, storage TemplateStorage) error

	// LastSyncAt returns the timestamp of the last successful sync.
	LastSyncAt() *time.Time

	// SyncStatus returns a short status string ("idle", "syncing", "error").
	SyncStatus() string
}
