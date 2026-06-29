package builtin

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"sync"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/dezhishen/now-and-again/backend/pkg/model"
	"github.com/dezhishen/now-and-again/backend/pkg/tasktemplate"
)

//go:embed templates/*.yaml
var embeddedYAML embed.FS

// Provider loads templates from embedded YAML files and syncs them into the DB.
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

	entries, err := readEmbeddedDir("templates")
	if err != nil {
		syncErr = fmt.Errorf("builtin: read embedded dir: %w", err)
		return syncErr
	}

	seen := make(map[string]bool)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		data, err := embeddedYAML.ReadFile("templates/" + entry.Name())
		if err != nil {
			syncErr = fmt.Errorf("builtin: read %s: %w", entry.Name(), err)
			return syncErr
		}

		doc, err := parseYAMLDocument(data)
		if err != nil {
			syncErr = fmt.Errorf("builtin: parse %s: %w", entry.Name(), err)
			return syncErr
		}

		for _, t := range doc.Templates {
			m := yamlEntryToModel("builtin", &t)
			if err := storage.UpsertTemplate(m); err != nil {
				syncErr = fmt.Errorf("builtin: upsert %s: %w", t.Code, err)
				return syncErr
			}
			seen[t.Code] = true
		}
	}

	// Remove templates that are no longer in the embedded YAML.
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

// ─── helpers ──────────────────────────────────────────────────────

func readEmbeddedDir(dir string) ([]fs.DirEntry, error) {
	return embeddedYAML.ReadDir("templates")
}

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
	extraJSON, _ := json.Marshal(e.ExtraSchema)

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
