package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"text/template"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/dezhishen/now-and-again/backend/pkg/model"

	"github.com/dezhishen/now-and-again/backend/internal/repository"
	"github.com/dezhishen/now-and-again/backend/pkg/tasktemplate"
	"github.com/dezhishen/now-and-again/backend/pkg/types"

	// Blank imports trigger init() registration of task template providers.
	_ "github.com/dezhishen/now-and-again/backend/pkg/tasktemplate/builtin"
	_ "github.com/dezhishen/now-and-again/backend/pkg/tasktemplate/http"
)

// TaskTemplateService handles template listing, rendering, and provider sync.
type TaskTemplateService struct {
	repo *repository.TaskTemplateRepo
}

func NewTaskTemplateService(repo *repository.TaskTemplateRepo) *TaskTemplateService {
	return &TaskTemplateService{repo: repo}
}

// ─── Family-visible queries ───────────────────────────────────────

// List returns all templates visible to a given family (system-level + family-level).
func (s *TaskTemplateService) List(ctx context.Context, familyID uuid.UUID, kind string) ([]types.TaskTemplate, error) {
	models, err := s.repo.FindAllForFamily(familyID.String(), kind)
	if err != nil {
		return nil, fmt.Errorf("list task templates: %w", err)
	}
	result := make([]types.TaskTemplate, len(models))
	for i, m := range models {
		result[i] = modelToDTO(&m)
	}
	return result, nil
}

// GetByCode returns a single template visible to the family.
func (s *TaskTemplateService) GetByCode(ctx context.Context, familyID uuid.UUID, code string) (*types.TaskTemplate, error) {
	m, err := s.repo.FindForFamilyByCode(familyID.String(), code)
	if err != nil {
		return nil, fmt.Errorf("get task template %s: %w", code, err)
	}
	dto := modelToDTO(m)
	return &dto, nil
}

// ─── Rendering ────────────────────────────────────────────────────

// Render resolves the template's TaskDefaults with the provided parameters.
func (s *TaskTemplateService) Render(ctx context.Context, familyID uuid.UUID, code string, params map[string]interface{}) (*types.RenderedTask, error) {
	m, err := s.repo.FindForFamilyByCode(familyID.String(), code)
	if err != nil {
		return nil, fmt.Errorf("render task template %s: %w", code, err)
	}

	var taskDefaults any
	if m.TaskDefaults != "" {
		rendered, err := renderTemplate(m.TaskDefaults, params)
		if err != nil {
			return nil, fmt.Errorf("render task defaults: %w", err)
		}
		if err := json.Unmarshal([]byte(rendered), &taskDefaults); err != nil {
			return nil, fmt.Errorf("parse rendered task defaults: %w", err)
		}
	}

	var extraSchema any
	if m.ExtraSchema != "" {
		if err := json.Unmarshal([]byte(m.ExtraSchema), &extraSchema); err != nil {
			return nil, fmt.Errorf("parse extra schema: %w", err)
		}
	}

	return &types.RenderedTask{
		TaskDefaults: taskDefaults,
		ExtraSchema:  extraSchema,
	}, nil
}

// ─── Family-level CRUD (owner only) ──────────────────────────────

// CreateFamily creates a new family-level template.
func (s *TaskTemplateService) CreateFamily(ctx context.Context, familyID uuid.UUID, req *types.CreateTaskTemplateRequest) (*types.TaskTemplate, error) {
	fid := familyID.String()
	paramsJSON, _ := json.Marshal(req.Parameters)
	defaultsJSON, _ := json.Marshal(req.TaskDefaults)
	extraJSON, _ := json.Marshal(req.ExtraSchema)

	m := &repository.TaskTemplateModel{
		FamilyID:     &fid,
		ProviderCode: "family", // family-owned templates
		TemplateCode: req.TemplateCode,
		Name:         req.Name,
		Description:  req.Description,
		Kind:         req.Kind,
		Icon:         req.Icon,
		SortOrder:    req.SortOrder,
		Enabled:      req.Enabled,
		Parameters:   string(paramsJSON),
		TaskDefaults: string(defaultsJSON),
		ExtraSchema:  string(extraJSON),
	}

	if err := s.repo.CreateFamilyTemplate(m); err != nil {
		return nil, fmt.Errorf("create family template: %w", err)
	}

	dto := modelToDTO(m)
	return &dto, nil
}

// UpdateFamily updates an existing family-level template.
func (s *TaskTemplateService) UpdateFamily(ctx context.Context, familyID uuid.UUID, code string, req *types.UpdateTaskTemplateRequest) (*types.TaskTemplate, error) {
	m, err := s.repo.FindFamilyOwnedByCode(familyID.String(), code)
	if err != nil {
		return nil, fmt.Errorf("find family template %s: %w", code, err)
	}

	if req.Name != nil {
		m.Name = *req.Name
	}
	if req.Description != nil {
		m.Description = *req.Description
	}
	if req.Kind != nil {
		m.Kind = *req.Kind
	}
	if req.Icon != nil {
		m.Icon = *req.Icon
	}
	if req.SortOrder != nil {
		m.SortOrder = *req.SortOrder
	}
	if req.Enabled != nil {
		m.Enabled = *req.Enabled
	}
	if req.Parameters != nil {
		b, _ := json.Marshal(req.Parameters)
		m.Parameters = string(b)
	}
	if req.TaskDefaults != nil {
		b, _ := json.Marshal(req.TaskDefaults)
		m.TaskDefaults = string(b)
	}
	if req.ExtraSchema != nil {
		b, _ := json.Marshal(req.ExtraSchema)
		m.ExtraSchema = string(b)
	}

	if err := s.repo.UpdateFamilyTemplate(m); err != nil {
		return nil, fmt.Errorf("update family template: %w", err)
	}

	dto := modelToDTO(m)
	return &dto, nil
}

// DeleteFamily deletes a family-level template.
func (s *TaskTemplateService) DeleteFamily(ctx context.Context, familyID uuid.UUID, code string) error {
	return s.repo.DeleteFamilyTemplate(familyID.String(), code)
}

// ─── Provider management ──────────────────────────────────────────

// ListProviders returns metadata for every registered provider.
func (s *TaskTemplateService) ListProviders(ctx context.Context) ([]types.TemplateProvider, error) {
	all := tasktemplate.AllProviders()
	result := make([]types.TemplateProvider, len(all))
	for i, p := range all {
		result[i] = types.TemplateProvider{
			Code:        p.Code(),
			Name:        p.Name(),
			Description: p.Description(),
			LastSyncAt:  p.LastSyncAt(),
			SyncStatus:  p.SyncStatus(),
		}
	}
	return result, nil
}

// RefreshSystemProvider triggers Sync on the named provider (system-level, admin-only).
func (s *TaskTemplateService) RefreshSystemProvider(ctx context.Context, providerCode string) error {
	p := tasktemplate.GetProvider(providerCode)
	if p == nil {
		return fmt.Errorf("unknown provider: %s", providerCode)
	}
	return p.Sync(ctx, s.repo)
}

// RefreshFamilyProvider triggers Sync for a family-level provider (e.g., family HTTP subs).
// Uses a family-scoped storage wrapper so the provider writes templates with the correct FamilyID.
func (s *TaskTemplateService) RefreshFamilyProvider(ctx context.Context, familyID uuid.UUID, providerCode string) error {
	p := tasktemplate.GetProvider(providerCode)
	if p == nil {
		return fmt.Errorf("unknown provider: %s", providerCode)
	}
	if providerCode == "builtin" {
		return fmt.Errorf("builtin provider is system-level only")
	}
	scoped := &familyScopedStorage{inner: s.repo, familyID: familyID.String()}
	return p.Sync(ctx, scoped)
}

// ─── Subscription management ──────────────────────────────────────

// ListSubscriptions returns subscriptions, system or family scoped.
func (s *TaskTemplateService) ListSubscriptions(ctx context.Context, familyID *uuid.UUID) ([]types.TaskTemplateSubscription, error) {
	var fid *string
	if familyID != nil {
		s := familyID.String()
		fid = &s
	}
	models, err := s.repo.FindSubscriptions(fid)
	if err != nil {
		return nil, fmt.Errorf("list subscriptions: %w", err)
	}
	result := make([]types.TaskTemplateSubscription, len(models))
	for i, m := range models {
		result[i] = subModelToDTO(&m)
	}
	return result, nil
}

// CreateSubscription creates a new subscription (system or family).
func (s *TaskTemplateService) CreateSubscription(ctx context.Context, familyID *uuid.UUID, req *types.CreateSubscriptionRequest) (*types.TaskTemplateSubscription, error) {
	var fid *string
	if familyID != nil {
		s := familyID.String()
		fid = &s
	}
	m := &repository.TaskTemplateSubscriptionModel{
		FamilyID:             fid,
		ProviderCode:         req.ProviderCode,
		URL:                  req.URL,
		Name:                 req.Name,
		AutoRefresh:          req.AutoRefresh,
		RefreshIntervalHours: req.RefreshIntervalHours,
		Enabled:              true,
	}
	if err := s.repo.CreateSubscription(m); err != nil {
		return nil, fmt.Errorf("create subscription: %w", err)
	}
	dto := subModelToDTO(m)
	return &dto, nil
}

// UpdateSubscription updates an existing subscription.
func (s *TaskTemplateService) UpdateSubscription(ctx context.Context, id string, req *types.UpdateSubscriptionRequest) (*types.TaskTemplateSubscription, error) {
	m, err := s.repo.FindSubscriptionByID(id)
	if err != nil {
		return nil, fmt.Errorf("find subscription %s: %w", id, err)
	}
	if req.URL != nil {
		m.URL = *req.URL
	}
	if req.Name != nil {
		m.Name = *req.Name
	}
	if req.AutoRefresh != nil {
		m.AutoRefresh = *req.AutoRefresh
	}
	if req.RefreshIntervalHours != nil {
		m.RefreshIntervalHours = *req.RefreshIntervalHours
	}
	if req.Enabled != nil {
		m.Enabled = *req.Enabled
	}
	if err := s.repo.UpdateSubscription(m); err != nil {
		return nil, fmt.Errorf("update subscription: %w", err)
	}
	dto := subModelToDTO(m)
	return &dto, nil
}

// DeleteSubscription removes a subscription.
func (s *TaskTemplateService) DeleteSubscription(ctx context.Context, id string) error {
	return s.repo.DeleteSubscription(id)
}

// SyncAll synchronises every registered provider at system level. Called at startup.
func (s *TaskTemplateService) SyncAll(ctx context.Context) error {
	for _, p := range tasktemplate.AllProviders() {
		if err := p.Sync(ctx, s.repo); err != nil {
			return fmt.Errorf("sync provider %s: %w", p.Code(), err)
		}
	}
	return nil
}

// ─── helpers ──────────────────────────────────────────────────────

func modelToDTO(m *repository.TaskTemplateModel) types.TaskTemplate {
	var params []types.TemplateParameter
	if m.Parameters != "" {
		json.Unmarshal([]byte(m.Parameters), &params)
	}
	var taskDefaults any
	if m.TaskDefaults != "" {
		json.Unmarshal([]byte(m.TaskDefaults), &taskDefaults)
	}
	var extraSchema any
	if m.ExtraSchema != "" {
		json.Unmarshal([]byte(m.ExtraSchema), &extraSchema)
	}

	return types.TaskTemplate{
		ID:           m.ID,
		FamilyID:     m.FamilyID,
		ProviderCode: m.ProviderCode,
		TemplateCode: m.TemplateCode,
		Name:         m.Name,
		Description:  m.Description,
		Kind:         m.Kind,
		Icon:         m.Icon,
		SortOrder:    m.SortOrder,
		Enabled:      m.Enabled,
		Parameters:   params,
		TaskDefaults: taskDefaults,
		ExtraSchema:  extraSchema,
		Version:      m.Version,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}

// renderTemplate applies Go text/template on a JSON string using the given params.
func renderTemplate(tmplStr string, params map[string]interface{}) (string, error) {
	tmpl, err := template.New("task").Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, params); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}
	return buf.String(), nil
}

func subModelToDTO(m *repository.TaskTemplateSubscriptionModel) types.TaskTemplateSubscription {
	return types.TaskTemplateSubscription{
		ID:                   m.ID,
		FamilyID:             m.FamilyID,
		ProviderCode:         m.ProviderCode,
		URL:                  m.URL,
		Name:                 m.Name,
		AutoRefresh:          m.AutoRefresh,
		RefreshIntervalHours: m.RefreshIntervalHours,
		Enabled:              m.Enabled,
		CreatedAt:            m.CreatedAt,
		UpdatedAt:            m.UpdatedAt,
	}
}

// ─── familyScopedStorage wraps TemplateStorage to auto-scope writes to a family ──

var _ tasktemplate.TemplateStorage = (*familyScopedStorage)(nil)

type familyScopedStorage struct {
	inner    tasktemplate.TemplateStorage
	familyID string
}

func (s *familyScopedStorage) UpsertTemplate(tmpl *model.TaskTemplateModel) error {
	tmpl.FamilyID = &s.familyID
	return s.inner.UpsertTemplate(tmpl)
}

func (s *familyScopedStorage) DeleteTemplate(providerCode, templateCode string) error {
	// Family-scoped delete: the repo's DeleteTemplate is system-only, so skip.
	// Family templates are managed via DeleteFamilyTemplate instead.
	return nil
}

func (s *familyScopedStorage) FindByProvider(providerCode string) ([]*model.TaskTemplateModel, error) {
	// Use family-scoped query instead of system-level.
	return s.inner.(*repository.TaskTemplateRepo).FindFamilyProviderTemplates(s.familyID, providerCode)
}

func (s *familyScopedStorage) ListSubscriptions(providerCode string) ([]tasktemplate.SubscriptionInfo, error) {
	// Override to return family-scoped subscriptions from the repo.
	models, err := s.inner.(*repository.TaskTemplateRepo).FindSubscriptions(&s.familyID)
	if err != nil {
		return nil, err
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

func (s *familyScopedStorage) DB() *gorm.DB {
	return s.inner.DB()
}
