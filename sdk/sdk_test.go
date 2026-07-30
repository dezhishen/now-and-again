package sdk

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/dezhishen/now-and-again/backend/pkg/types"
	"github.com/dezhishen/now-and-again/sdk/internal/client"
)

// ─── Config Tests ───────────────────────────────────────────────

func TestDefaultConfigPath(t *testing.T) {
	path, err := DefaultConfigPath()
	if err != nil {
		t.Fatalf("DefaultConfigPath: %v", err)
	}
	if !strings.HasSuffix(path, ".na.yaml") {
		t.Errorf("expected path ending with .na.yaml, got %q", path)
	}
	home, _ := os.UserHomeDir()
	if !strings.HasPrefix(path, home) {
		t.Errorf("expected path under home dir %q, got %q", home, path)
	}
}

func TestConfig_SaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, ".na.yaml")

	cfg := &Config{
		ServerURL:        "http://test:1234",
		Token:            "na_test_token",
		ActiveFamilyID:   "fam-1",
		ActiveFamilyName: "测试家庭",
	}
	cfg.SetPath(cfgPath)

	if err := cfg.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := LoadConfigFromPath(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfigFromPath: %v", err)
	}

	if loaded.ServerURL != "http://test:1234" {
		t.Errorf("expected ServerURL 'http://test:1234', got %q", loaded.ServerURL)
	}
	if loaded.Token != "na_test_token" {
		t.Errorf("expected Token 'na_test_token', got %q", loaded.Token)
	}
	if loaded.ActiveFamilyID != "fam-1" {
		t.Errorf("expected ActiveFamilyID 'fam-1', got %q", loaded.ActiveFamilyID)
	}
	if loaded.ActiveFamilyName != "测试家庭" {
		t.Errorf("expected ActiveFamilyName '测试家庭', got %q", loaded.ActiveFamilyName)
	}
}

func TestLoadConfigFromPath_NotExist(t *testing.T) {
	cfg, err := LoadConfigFromPath("/nonexistent/path/.na.yaml")
	if err != nil {
		t.Fatalf("LoadConfigFromPath for missing file: %v", err)
	}
	if cfg.ServerURL != "http://localhost:8080" {
		t.Errorf("expected default ServerURL 'http://localhost:8080', got %q", cfg.ServerURL)
	}
}

func TestConfig_SaveAndLoad_EmptyFields(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, ".na.yaml")

	cfg := &Config{ServerURL: "http://localhost:3000"}
	cfg.SetPath(cfgPath)
	cfg.Save()

	loaded, err := LoadConfigFromPath(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfigFromPath: %v", err)
	}
	if loaded.ServerURL != "http://localhost:3000" {
		t.Errorf("expected ServerURL 'http://localhost:3000', got %q", loaded.ServerURL)
	}
	if loaded.Token != "" {
		t.Errorf("expected empty Token, got %q", loaded.Token)
	}
}

// ─── HTTP Client Tests ──────────────────────────────────────────

func TestHTTPClient_DetectAuthMethod(t *testing.T) {
	tests := []struct {
		token    string
		expected string
	}{
		{"na_abc123", "apikey"},
		{"na_longer_than_3", "apikey"},
		{"", "jwt"},
		{"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9", "jwt"},
	}
	for _, tc := range tests {
		c := client.NewHTTPClient("http://localhost", tc.token)
		if c.AuthMethod != tc.expected {
			t.Errorf("token=%q: expected AuthMethod=%q, got %q", tc.token, tc.expected, c.AuthMethod)
		}
	}
}

// writeSuccess writes a successful API envelope response.
func writeSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    data,
	})
}

// writeError writes an error API envelope response.
func writeError(w http.ResponseWriter, status int, msg string) {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"error":   msg,
	})
}

func TestHTTPClient_Do_GET_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer test-jwt" {
			t.Errorf("expected Bearer token, got %q", r.Header.Get("Authorization"))
		}
		writeSuccess(w, types.Task{ID: "task-1", Name: "测试任务"})
	}))
	defer ts.Close()

	c := client.NewHTTPClient(ts.URL, "test-jwt")
	var result types.Task
	if err := c.Do("GET", "/api/tasks/test", nil, &result); err != nil {
		t.Fatalf("Do: %v", err)
	}
	if result.ID != "task-1" {
		t.Errorf("expected ID 'task-1', got %q", result.ID)
	}
	if result.Name != "测试任务" {
		t.Errorf("expected Name '测试任务', got %q", result.Name)
	}
}

func TestHTTPClient_Do_POST_WithBody(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json, got %q", r.Header.Get("Content-Type"))
		}
		if r.Header.Get("X-API-Key") != "na_apikey" {
			t.Errorf("expected X-API-Key header, got %q", r.Header.Get("X-API-Key"))
		}
		w.WriteHeader(http.StatusCreated)
		writeSuccess(w, map[string]string{"id": "new-task", "name": "created"})
	}))
	defer ts.Close()

	c := client.NewHTTPClient(ts.URL, "na_apikey")
	var result map[string]string
	req := map[string]string{"name": "test"}
	if err := c.Do("POST", "/api/tasks", req, &result); err != nil {
		t.Fatalf("Do: %v", err)
	}
	if result["id"] != "new-task" {
		t.Errorf("expected id 'new-task', got %q", result["id"])
	}
}

func TestHTTPClient_Do_NotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotFound, "not found")
	}))
	defer ts.Close()

	c := client.NewHTTPClient(ts.URL, "test")
	var result interface{}
	err := c.Do("GET", "/api/notfound", nil, &result)
	if err == nil {
		t.Error("expected error for 404 response")
	}
}

func TestHTTPClient_Do_ServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusInternalServerError, "server error")
	}))
	defer ts.Close()

	c := client.NewHTTPClient(ts.URL, "test")
	var result interface{}
	err := c.Do("GET", "/api/error", nil, &result)
	if err == nil {
		t.Error("expected error for 500 response")
	}
}

func TestHTTPClient_FamilyID_Header(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fid := r.Header.Get("X-Family-Id")
		if fid != "test-family-id" {
			t.Errorf("expected X-Family-Id 'test-family-id', got %q", fid)
		}
		writeSuccess(w, map[string]string{"id": "ok"})
	}))
	defer ts.Close()

	c := client.NewHTTPClient(ts.URL, "test")
	c.SetFamilyID("test-family-id")

	var result map[string]string
	if err := c.Do("GET", "/api/test", nil, &result); err != nil {
		t.Fatalf("Do: %v", err)
	}
	if result["id"] != "ok" {
		t.Errorf("expected id='ok', got %q", result["id"])
	}
}

func TestHTTPClient_SetBaseURL(t *testing.T) {
	c := client.NewHTTPClient("http://old", "test")
	c.SetBaseURL("http://new")
	// Can't easily inspect BaseURL after change, but ensure no panic
	c.SetToken("new-token")
	if c.Token != "new-token" {
		t.Errorf("expected token 'new-token', got %q", c.Token)
	}
}

// ─── NA Config Integration ─────────────────────────────────────

func TestNewWithConfig_Default(t *testing.T) {
	cfg := &Config{ServerURL: "http://test:8080"}
	na := NewWithConfig(cfg)
	if na == nil {
		t.Fatal("expected non-nil NA")
	}
	if na.cfg.ServerURL != "http://test:8080" {
		t.Errorf("expected ServerURL 'http://test:8080', got %q", na.cfg.ServerURL)
	}
}

func TestNewWithConfig_Nil(t *testing.T) {
	// Should not panic
	cfg := &Config{}
	na := NewWithConfig(cfg)
	if na == nil {
		t.Fatal("expected non-nil NA")
	}
}

// ─── TaskClient Tests ──────────────────────────────────────────

func TestTaskClient_Get(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/tasks/task-1" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		writeSuccess(w, types.Task{ID: "task-1", Name: "测试", Enabled: true})
	}))
	defer ts.Close()

	c := client.NewHTTPClient(ts.URL, "test")
	tc := client.NewTaskClient(c)

	task, err := tc.Get("task-1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if task.ID != "task-1" {
		t.Errorf("expected ID 'task-1', got %q", task.ID)
	}
}

func TestTaskClient_SetEnabled(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/api/tasks/task-1/enabled" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		writeSuccess(w, types.Task{ID: "task-1", Enabled: true})
	}))
	defer ts.Close()

	c := client.NewHTTPClient(ts.URL, "test")
	tc := client.NewTaskClient(c)

	task, err := tc.SetEnabled("task-1", true)
	if err != nil {
		t.Fatalf("SetEnabled: %v", err)
	}
	if !task.Enabled {
		t.Error("expected Enabled=true")
	}
}

func TestTaskClient_List(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/tasks" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		writeSuccess(w, []types.Task{
			{ID: "task-1", Name: "A"},
			{ID: "task-2", Name: "B"},
		})
	}))
	defer ts.Close()

	c := client.NewHTTPClient(ts.URL, "test")
	tc := client.NewTaskClient(c)

	tasks, err := tc.List("family-1")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(tasks))
	}
	if tasks[0].Name != "A" || tasks[1].Name != "B" {
		t.Errorf("unexpected task order: %+v", tasks)
	}
}

// ─── NA Method Tests ───────────────────────────────────────────

func TestNA_GetTimezone_Default(t *testing.T) {
	na := NewWithConfig(&Config{})
	tz := na.GetTimezone()
	if tz == nil {
		t.Error("expected non-nil timezone")
	}
}

func TestNA_SetTimezone(t *testing.T) {
	na := NewWithConfig(&Config{})
	loc, _ := time.LoadLocation("Asia/Shanghai")
	na.SetTimezone(loc)
	tz := na.GetTimezone()
	if tz.String() != "Asia/Shanghai" {
		t.Errorf("expected 'Asia/Shanghai', got %q", tz.String())
	}
}

func TestNA_RequireFamilyID_NotSet(t *testing.T) {
	na := NewWithConfig(&Config{})
	_, err := na.requireFamilyID()
	if err == nil {
		t.Error("expected error when no active family")
	}
}

func TestNA_RequireFamilyID_Set(t *testing.T) {
	na := NewWithConfig(&Config{})
	na.SetActiveFamily("550e8400-e29b-41d4-a716-446655440000", "Test")
	fid, err := na.requireFamilyID()
	if err != nil {
		t.Fatalf("requireFamilyID: %v", err)
	}
	if fid.String() != "550e8400-e29b-41d4-a716-446655440000" {
		t.Errorf("expected '550e8400-e29b-41d4-a716-446655440000', got %q", fid.String())
	}
}

func TestNA_CompleteTodoSimple_InvalidID(t *testing.T) {
	na := NewWithConfig(&Config{})
	ctx := context.Background()
	_, err := na.CompleteTodoSimple(ctx, "not-a-uuid", "done")
	if err == nil {
		t.Error("expected error for invalid UUID")
	}
}
