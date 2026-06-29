package repository

import (
	"fmt"

	"github.com/dezhishen/now-and-again/backend/pkg/tasktemplate"
	"gorm.io/gorm"
)

// TaskTemplateRepo provides CRUD for the task_templates table.
type TaskTemplateRepo struct {
	db *gorm.DB
}

func NewTaskTemplateRepo(db *gorm.DB) *TaskTemplateRepo {
	return &TaskTemplateRepo{db: db}
}

func (r *TaskTemplateRepo) DB() *gorm.DB { return r.db }

// ─── Queries (service-layer) ──────────────────────────────────────

// FindAllForFamily returns all enabled templates visible to a given family:
// system-level (family_id IS NULL) + family-level (family_id = familyID).
func (r *TaskTemplateRepo) FindAllForFamily(familyID, kind string) ([]TaskTemplateModel, error) {
	var models []TaskTemplateModel
	q := r.db.Where("enabled = ?", true).
		Where("family_id IS NULL OR family_id = ?", familyID).
		Order("sort_order ASC, name ASC")
	if kind != "" {
		q = q.Where("kind = ?", kind)
	}
	if err := q.Find(&models).Error; err != nil {
		return nil, fmt.Errorf("find task templates: %w", err)
	}
	return models, nil
}

// FindAllSystem returns system-level templates (family_id IS NULL).
func (r *TaskTemplateRepo) FindAllSystem(kind string) ([]TaskTemplateModel, error) {
	var models []TaskTemplateModel
	q := r.db.Where("enabled = ? AND family_id IS NULL", true).Order("sort_order ASC, name ASC")
	if kind != "" {
		q = q.Where("kind = ?", kind)
	}
	if err := q.Find(&models).Error; err != nil {
		return nil, fmt.Errorf("find system task templates: %w", err)
	}
	return models, nil
}

// FindForFamilyByCode looks up a template that is either system-level or belongs to the given family.
func (r *TaskTemplateRepo) FindForFamilyByCode(familyID, templateCode string) (*TaskTemplateModel, error) {
	var m TaskTemplateModel
	err := r.db.Where("template_code = ?", templateCode).
		Where("family_id IS NULL OR family_id = ?", familyID).
		First(&m).Error
	if err != nil {
		return nil, fmt.Errorf("find task template by code %s: %w", templateCode, err)
	}
	return &m, nil
}

// FindFamilyOwnedByCode looks up a family-level template (must belong to family).
func (r *TaskTemplateRepo) FindFamilyOwnedByCode(familyID, templateCode string) (*TaskTemplateModel, error) {
	var m TaskTemplateModel
	err := r.db.Where("template_code = ? AND family_id = ?", templateCode, familyID).First(&m).Error
	if err != nil {
		return nil, fmt.Errorf("find family template %s: %w", templateCode, err)
	}
	return &m, nil
}

// ─── Family-level CRUD ───────────────────────────────────────────

// CreateFamilyTemplate inserts a new family-level template.
func (r *TaskTemplateRepo) CreateFamilyTemplate(tmpl *TaskTemplateModel) error {
	return r.db.Create(tmpl).Error
}

// UpdateFamilyTemplate updates an existing family-level template.
func (r *TaskTemplateRepo) UpdateFamilyTemplate(tmpl *TaskTemplateModel) error {
	return r.db.Save(tmpl).Error
}

// DeleteFamilyTemplate deletes a family-level template.
func (r *TaskTemplateRepo) DeleteFamilyTemplate(familyID, templateCode string) error {
	return r.db.Where("template_code = ? AND family_id = ?", templateCode, familyID).
		Delete(&TaskTemplateModel{}).Error
}

// ─── TemplateStorage implementation (for provider Sync) ──────────

// UpsertTemplate inserts or updates a template record.
// Key: (family_id, provider_code, template_code). family_id may be nil.
func (r *TaskTemplateRepo) UpsertTemplate(tmpl *TaskTemplateModel) error {
	var existing TaskTemplateModel
	q := r.db.Where("provider_code = ? AND template_code = ?", tmpl.ProviderCode, tmpl.TemplateCode)
	if tmpl.FamilyID == nil {
		q = q.Where("family_id IS NULL")
	} else {
		q = q.Where("family_id = ?", *tmpl.FamilyID)
	}
	err := q.First(&existing).Error
	if err == nil {
		existing.Name = tmpl.Name
		existing.Description = tmpl.Description
		existing.Kind = tmpl.Kind
		existing.Icon = tmpl.Icon
		existing.SortOrder = tmpl.SortOrder
		existing.Enabled = tmpl.Enabled
		existing.Parameters = tmpl.Parameters
		existing.TaskDefaults = tmpl.TaskDefaults
		existing.ExtraSchema = tmpl.ExtraSchema
		existing.Version = tmpl.Version
		existing.Metadata = tmpl.Metadata
		return r.db.Save(&existing).Error
	}
	if err == gorm.ErrRecordNotFound {
		return r.db.Create(tmpl).Error
	}
	return fmt.Errorf("upsert task template: %w", err)
}

// DeleteTemplate removes templates by provider and template code (system-level only, used by builtin provider).
func (r *TaskTemplateRepo) DeleteTemplate(providerCode, templateCode string) error {
	return r.db.Where("provider_code = ? AND template_code = ? AND family_id IS NULL",
		providerCode, templateCode).Delete(&TaskTemplateModel{}).Error
}

// FindByProvider returns all system-level templates for a given provider (part of TemplateStorage).
func (r *TaskTemplateRepo) FindByProvider(providerCode string) ([]*TaskTemplateModel, error) {
	var models []TaskTemplateModel
	if err := r.db.Where("provider_code = ? AND family_id IS NULL", providerCode).
		Find(&models).Error; err != nil {
		return nil, fmt.Errorf("find templates by provider %s: %w", providerCode, err)
	}
	result := make([]*TaskTemplateModel, len(models))
	for i := range models {
		result[i] = &models[i]
	}
	return result, nil
}

// FindFamilyProviderTemplates returns all family-level templates for a given provider.
func (r *TaskTemplateRepo) FindFamilyProviderTemplates(familyID, providerCode string) ([]*TaskTemplateModel, error) {
	var models []TaskTemplateModel
	if err := r.db.Where("provider_code = ? AND family_id = ?", providerCode, familyID).
		Find(&models).Error; err != nil {
		return nil, fmt.Errorf("find family templates by provider %s: %w", providerCode, err)
	}
	result := make([]*TaskTemplateModel, len(models))
	for i := range models {
		result[i] = &models[i]
	}
	return result, nil
}

// ─── Subscription queries (TemplateStorage + CRUD) ───────────────

// ListSubscriptions returns active subscriptions for a provider.
// When called on the raw repo, returns system-level subscriptions (family_id IS NULL).
// When called via familyScopedStorage, returns family-level subscriptions.
func (r *TaskTemplateRepo) ListSubscriptions(providerCode string) ([]tasktemplate.SubscriptionInfo, error) {
	var models []TaskTemplateSubscriptionModel
	if err := r.db.Where("provider_code = ? AND enabled = ? AND family_id IS NULL",
		providerCode, true).Find(&models).Error; err != nil {
		return nil, fmt.Errorf("list subscriptions: %w", err)
	}
	result := make([]tasktemplate.SubscriptionInfo, len(models))
	for i, m := range models {
		result[i] = tasktemplate.SubscriptionInfo{
			URL:                  m.URL,
			Name:                 m.Name,
			AutoRefresh:          m.AutoRefresh,
			RefreshIntervalHours: m.RefreshIntervalHours,
		}
	}
	return result, nil
}

// ─── Subscription CRUD ───────────────────────────────────────────

// FindSubscriptions returns all subscriptions, optionally filtered by family.
func (r *TaskTemplateRepo) FindSubscriptions(familyID *string) ([]TaskTemplateSubscriptionModel, error) {
	var models []TaskTemplateSubscriptionModel
	q := r.db.Order("created_at ASC")
	if familyID == nil {
		q = q.Where("family_id IS NULL")
	} else {
		q = q.Where("family_id = ?", *familyID)
	}
	if err := q.Find(&models).Error; err != nil {
		return nil, fmt.Errorf("find subscriptions: %w", err)
	}
	return models, nil
}

// CreateSubscription inserts a new subscription.
func (r *TaskTemplateRepo) CreateSubscription(sub *TaskTemplateSubscriptionModel) error {
	return r.db.Create(sub).Error
}

// UpdateSubscription updates an existing subscription.
func (r *TaskTemplateRepo) UpdateSubscription(sub *TaskTemplateSubscriptionModel) error {
	return r.db.Save(sub).Error
}

// DeleteSubscription removes a subscription by ID.
func (r *TaskTemplateRepo) DeleteSubscription(id string) error {
	return r.db.Delete(&TaskTemplateSubscriptionModel{}, "id = ?", id).Error
}

// FindSubscriptionByID looks up a single subscription.
func (r *TaskTemplateRepo) FindSubscriptionByID(id string) (*TaskTemplateSubscriptionModel, error) {
	var m TaskTemplateSubscriptionModel
	if err := r.db.Where("id = ?", id).First(&m).Error; err != nil {
		return nil, fmt.Errorf("find subscription %s: %w", id, err)
	}
	return &m, nil
}
