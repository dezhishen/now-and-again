package service

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuiltinFamilyDefaults_NotEmpty(t *testing.T) {
	cfg := builtinFamilyDefaults()
	if len(cfg.Locations) == 0 {
		t.Fatal("built-in locations should not be empty")
	}
	if len(cfg.Groups) == 0 {
		t.Fatal("built-in groups should not be empty")
	}
}

func TestBuiltinFamilyDefaults_HasExpectedKeys(t *testing.T) {
	cfg := builtinFamilyDefaults()

	names := make(map[string]bool)
	for _, l := range cfg.Locations {
		names[l.Name] = true
	}
	// These are the expected default locations
	for _, expected := range []string{"厨房", "客厅", "主卧", "次卧", "卫生间", "阳台"} {
		if !names[expected] {
			t.Errorf("missing expected location: %s", expected)
		}
	}

	groupNames := make(map[string]bool)
	for _, g := range cfg.Groups {
		groupNames[g.Name] = true
	}
	for _, expected := range []string{"大人", "小孩"} {
		if !groupNames[expected] {
			t.Errorf("missing expected group: %s", expected)
		}
	}
}

func TestLoadFamilyDefaults_ValidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "family_defaults.yaml")
	content := `
locations:
  - name: "办公室"
    kind: "indoor"
    color: "#000000"
groups:
  - name: "员工"
    description: "公司员工"
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	LoadFamilyDefaults(path)

	if len(familyDefaults.Locations) != 1 {
		t.Fatalf("expected 1 location, got %d", len(familyDefaults.Locations))
	}
	if familyDefaults.Locations[0].Name != "办公室" {
		t.Errorf("got %q, want 办公室", familyDefaults.Locations[0].Name)
	}
	if len(familyDefaults.Groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(familyDefaults.Groups))
	}
	if familyDefaults.Groups[0].Name != "员工" {
		t.Errorf("got %q, want 员工", familyDefaults.Groups[0].Name)
	}
}

func TestLoadFamilyDefaults_MissingFile(t *testing.T) {
	// Reset to known state
	familyDefaults = familyDefaultsConfig{}
	LoadFamilyDefaults("/nonexistent/path/family_defaults.yaml")

	if len(familyDefaults.Locations) == 0 {
		t.Fatal("should fall back to built-in defaults when file missing")
	}
	// Should have the built-in locations
	names := make(map[string]bool)
	for _, l := range familyDefaults.Locations {
		names[l.Name] = true
	}
	if !names["厨房"] {
		t.Error("missing built-in default location 厨房")
	}
}

func TestLoadFamilyDefaults_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yaml")
	if err := os.WriteFile(path, []byte("{{{invalid yaml!!!"), 0644); err != nil {
		t.Fatal(err)
	}

	LoadFamilyDefaults(path)

	// Should fall back to built-in defaults
	if len(familyDefaults.Locations) == 0 {
		t.Fatal("should fall back to built-in defaults when YAML is invalid")
	}
}

func TestLoadFamilyDefaults_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.yaml")
	if err := os.WriteFile(path, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	// Reset before test since familyDefaults is package-level
	familyDefaults = familyDefaultsConfig{}

	LoadFamilyDefaults(path)

	// Empty YAML unmarshals to zero struct → no locations/groups
	if len(familyDefaults.Locations) != 0 {
		t.Errorf("expected 0 locations for empty file, got %d", len(familyDefaults.Locations))
	}
}

func TestFamilyDefaultsEnabled_Default(t *testing.T) {
	// Unset env var → should be enabled by default
	os.Unsetenv("NA_FAMILY_DEFAULTS_INIT")
	if !familyDefaultsEnabled() {
		t.Error("should be enabled by default")
	}
}

func TestFamilyDefaultsEnabled_False(t *testing.T) {
	os.Setenv("NA_FAMILY_DEFAULTS_INIT", "false")
	defer os.Unsetenv("NA_FAMILY_DEFAULTS_INIT")
	if familyDefaultsEnabled() {
		t.Error("should be disabled when set to false")
	}
}

func TestFamilyDefaultsEnabled_Zero(t *testing.T) {
	os.Setenv("NA_FAMILY_DEFAULTS_INIT", "0")
	defer os.Unsetenv("NA_FAMILY_DEFAULTS_INIT")
	if familyDefaultsEnabled() {
		t.Error("should be disabled when set to 0")
	}
}

func TestFamilyDefaultsEnabled_True(t *testing.T) {
	os.Setenv("NA_FAMILY_DEFAULTS_INIT", "true")
	defer os.Unsetenv("NA_FAMILY_DEFAULTS_INIT")
	if !familyDefaultsEnabled() {
		t.Error("should be enabled when set to true")
	}
}
