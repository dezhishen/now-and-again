package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/dezhishen/now-and-again/backend/pkg/model"
	"github.com/dezhishen/now-and-again/backend/pkg/tasktemplate"
)

// Provider fetches templates from remote HTTP(S) URLs and syncs them into the DB.
// Subscription URLs are read from TemplateStorage.ListSubscriptions(), which is
// scoped to system-level or family-level by the caller.
type Provider struct {
	mu         sync.Mutex
	lastSync   time.Time
	syncStatus string
}

func init() {
	tasktemplate.Register(&Provider{syncStatus: "idle"})
}

func (p *Provider) Code() string        { return "http" }
func (p *Provider) Name() string        { return "远程订阅" }
func (p *Provider) Description() string { return "通过 HTTP(S) 订阅远程任务模板" }

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
	subs, err := storage.ListSubscriptions("http")
	if err != nil {
		return fmt.Errorf("http: read subscriptions: %w", err)
	}
	if len(subs) == 0 {
		return nil
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

	seen := make(map[string]bool)

	for _, sub := range subs {
		data, err := fetchURL(ctx, sub.URL)
		if err != nil {
			syncErr = fmt.Errorf("http: fetch %s: %w", sub.URL, err)
			return syncErr
		}

		doc, err := parseYAMLDocument(data)
		if err != nil {
			syncErr = fmt.Errorf("http: parse %s: %w", sub.URL, err)
			return syncErr
		}

		for _, t := range doc.Templates {
			m := yamlEntryToModel("http", &t, sub.URL)
			if err := storage.UpsertTemplate(m); err != nil {
				syncErr = fmt.Errorf("http: upsert %s from %s: %w", t.Code, sub.URL, err)
				return syncErr
			}
			seen[t.Code] = true
		}
	}

	// Remove templates that are no longer in any subscription.
	existing, err := storage.FindByProvider("http")
	if err != nil {
		syncErr = fmt.Errorf("http: list existing: %w", err)
		return syncErr
	}
	for _, e := range existing {
		if !seen[e.TemplateCode] {
			if err := storage.DeleteTemplate("http", e.TemplateCode); err != nil {
				syncErr = fmt.Errorf("http: delete stale %s: %w", e.TemplateCode, err)
				return syncErr
			}
		}
	}

	return nil
}

// ─── helpers ──────────────────────────────────────────────────────

var httpClient = &http.Client{Timeout: 30 * time.Second}

func fetchURL(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
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

func yamlEntryToModel(providerCode string, e *tasktemplate.TemplateYAMLEntry, sourceURL string) *model.TaskTemplateModel {
	enabled := true
	if e.Enabled != nil {
		enabled = *e.Enabled
	}

	paramsJSON, _ := json.Marshal(e.Parameters)
	defaultsJSON, _ := json.Marshal(e.TaskDefaults)
	extraJSON, _ := json.Marshal(e.ExtraSchema)
	meta, _ := json.Marshal(map[string]string{"source_url": sourceURL})

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
		Metadata:     string(meta),
	}
}
