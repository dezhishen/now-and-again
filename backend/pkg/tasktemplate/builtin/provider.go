package builtin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/dezhishen/now-and-again/backend/pkg/logger"
	"github.com/dezhishen/now-and-again/backend/pkg/model"
	"github.com/dezhishen/now-and-again/backend/pkg/tasktemplate"
)

// dataDir is set once at startup via SetDataDir.
var dataDir string

// SetDataDir sets the runtime data directory for reading templates from disk.
// Must be called before Sync().
func SetDataDir(dir string) { dataDir = dir }

// templatesDir returns the path where administrators can place custom .yaml templates.
// This is NOT populated by the application — admins manage files manually.
func templatesDir() string { return filepath.Join(dataDir, "templates") }

// ─── Provider ─────────────────────────────────────────────────────

type Provider struct {
	mu         sync.Mutex
	lastSync   time.Time
	syncStatus string
}

func init() {
	tasktemplate.Register(&Provider{syncStatus: "idle"})
}

func (p *Provider) Code() string        { return "builtin" }
func (p *Provider) Name() string        { return "内置模板" }
func (p *Provider) Description() string { return "系统预置的任务模板" }

func (p *Provider) LastSyncAt() *time.Time {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.lastSync.IsZero() {
		return nil
	}
	t := p.lastSync
	return &t
}

func (p *Provider) SyncStatus() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.syncStatus
}

func (p *Provider) Sync(ctx context.Context, storage tasktemplate.TemplateStorage) error {
	if dataDir == "" {
		return fmt.Errorf("builtin: data dir not set, call SetDataDir before Sync")
	}

	p.mu.Lock()
	p.syncStatus = "syncing"
	p.mu.Unlock()

	var syncErr error
	defer func() {
		p.mu.Lock()
		if syncErr != nil {
			p.syncStatus = "error"
		} else {
			p.syncStatus = "idle"
			p.lastSync = time.Now()
		}
		p.mu.Unlock()
	}()

	dir := templatesDir()

	// Ensure the templates directory exists (admins put .yaml files here).
	if err := os.MkdirAll(dir, 0755); err != nil {
		syncErr = fmt.Errorf("builtin: create templates dir %s: %w", dir, err)
		return syncErr
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		logger.Infof("[builtin] no templates directory at %s, skipping (this is normal if no custom templates)", dir)
		return nil
	}

	if len(entries) == 0 {
		logger.Infof("[builtin] templates directory %s is empty, skipping", dir)
		return nil
	}

	seen := make(map[string]bool)
	hadFailure := false

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) != ".yaml" && filepath.Ext(entry.Name()) != ".yml" {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			logger.Warnf("[builtin] read %s: %v (skipping)", entry.Name(), err)
			hadFailure = true
			continue
		}

		doc, err := parseYAMLDocument(data)
		if err != nil {
			logger.Warnf("[builtin] parse %s: %v (skipping)", entry.Name(), err)
			hadFailure = true
			continue
		}

		for _, t := range doc.Templates {
			m := yamlEntryToModel("builtin", &t)
			if err := storage.UpsertTemplate(m); err != nil {
				logger.Warnf("[builtin] upsert %s: %v", t.Code, err)
				continue
			}
			seen[t.Code] = true
		}
		logger.Infof("[builtin] synced %s (%d templates)", entry.Name(), len(doc.Templates))
	}

	if hadFailure {
		logger.Warnf("[builtin] one or more files failed, keeping existing template data")
		return nil
	}

	// Remove templates that are no longer in the templates directory.
	existing, err := storage.FindByProvider("builtin")
	if err != nil {
		syncErr = fmt.Errorf("builtin: list existing: %w", err)
		return syncErr
	}
	for _, e := range existing {
		if !seen[e.TemplateCode] {
			if err := storage.DeleteTemplate("builtin", e.TemplateCode); err != nil {
				syncErr = fmt.Errorf("builtin: delete stale %s: %w", e.TemplateCode, err)
				return syncErr
			}
		}
	}

	return nil
}

// SyncOne is not supported for the builtin provider (no subscriptions).
func (p *Provider) SyncOne(ctx context.Context, storage tasktemplate.TemplateStorage, subscriptionURL string) error {
	return fmt.Errorf("builtin: SyncOne not supported, use Sync instead")
}

// ─── helpers ──────────────────────────────────────────────────────

func parseYAMLDocument(data []byte) (*tasktemplate.TemplateYAMLDocument, error) {
	dec := yaml.NewDecoder(bytes.NewReader(data))
	var doc tasktemplate.TemplateYAMLDocument
	if err := dec.Decode(&doc); err != nil {
		return nil, err
	}
	if doc.Version == 0 {
		doc.Version = 1
	}
	return &doc, nil
}

func yamlEntryToModel(providerCode string, e *tasktemplate.TemplateYAMLEntry) *model.TaskTemplateModel {
	enabled := true
	if e.Enabled != nil {
		enabled = *e.Enabled
	}

	paramsJSON, _ := json.Marshal(e.Parameters)
	defaultsJSON, _ := json.Marshal(e.TaskDefaults)

	var extraJSON []byte
	if s, ok := e.ExtraSchema.(string); ok {
		// YAML literal block scalar (|) → raw Go template string
		extraJSON = []byte(s)
	} else {
		extraJSON, _ = json.Marshal(e.ExtraSchema)
	}

	return &model.TaskTemplateModel{
		ProviderCode: providerCode,
		TemplateCode: e.Code,
		Name:         e.Name,
		Description:  e.Description,
		Kind:         e.Kind,
		Icon:         e.Icon,
		SortOrder:    e.SortOrder,
		Enabled:      enabled,
		Parameters:   string(paramsJSON),
		TaskDefaults: string(defaultsJSON),
		ExtraSchema:  string(extraJSON),
		Version:      e.Version,
	}
}
