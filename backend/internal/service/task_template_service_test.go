package service

import (
	"testing"
)

func TestRenderTemplate_BasicString(t *testing.T) {
	result, err := renderTemplate(`{"name":"{{.greeting}}"}`, map[string]interface{}{
		"greeting": "hello",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != `{"name":"hello"}` {
		t.Errorf("got %q, want %q", result, `{"name":"hello"}`)
	}
}

func TestRenderTemplate_IntValue(t *testing.T) {
	result, err := renderTemplate(`{"count":{{.n}}}`, map[string]interface{}{
		"n": 42,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != `{"count":42}` {
		t.Errorf("got %q, want %q", result, `{"count":42}`)
	}
}

func TestRenderTemplate_MultipleParams(t *testing.T) {
	result, err := renderTemplate(`{"name":"{{.area}} - check","time":"{{.t}}"}`, map[string]interface{}{
		"area": "kitchen",
		"t":    "09:00",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != `{"name":"kitchen - check","time":"09:00"}` {
		t.Errorf("got %q, want %q", result, `{"name":"kitchen - check","time":"09:00"}`)
	}
}

func TestRenderTemplate_RangeOverArray(t *testing.T) {
	result, err := renderTemplate(
		`"{{range $i, $v := .items}}{{if $i}}、{{end}}{{$v}}{{end}}"`,
		map[string]interface{}{
			"items": []interface{}{"主卧", "次卧", "儿童房"},
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != `"主卧、次卧、儿童房"` {
		t.Errorf("got %q, want %q", result, `"主卧、次卧、儿童房"`)
	}
}

func TestRenderTemplate_EmptyArray(t *testing.T) {
	result, err := renderTemplate(
		`"{{range .items}}{{.}}{{end}}"`,
		map[string]interface{}{
			"items": []interface{}{},
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != `""` {
		t.Errorf("got %q, want %q", result, `""`)
	}
}

func TestRenderTemplate_ScheduleType(t *testing.T) {
	result, err := renderTemplate(
		`{"schedule_type":"{{.sched}}","schedule_data":{"time":"{{.sched_time}}"}}`,
		map[string]interface{}{
			"sched":      "weekly",
			"sched_time": "10:00",
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := `{"schedule_type":"weekly","schedule_data":{"time":"10:00"}}`
	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}

func TestRenderTemplate_YAMLExtraSchema(t *testing.T) {
	// Simulates the | block in YAML extra_schema with range
	result, err := renderTemplate(
		`check_items:
  {{range $room := .rooms}}
    - name: "{{$room}}床品"
      branches:
        - name: "已换洗"
        - name: "未换洗"
          create_todo: true
  {{end}}`,
		map[string]interface{}{
			"rooms": []interface{}{"主卧", "次卧"},
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == "" {
		t.Fatal("expected non-empty YAML output")
	}
	// Should contain both rooms
	if !contains(result, "主卧床品") || !contains(result, "次卧床品") {
		t.Errorf("expected both rooms in output, got: %s", result)
	}
}

func TestRenderTemplate_InvalidSyntax(t *testing.T) {
	_, err := renderTemplate(`{"name":"{{.bad"`, nil)
	if err == nil {
		t.Fatal("expected error for invalid template syntax")
	}
}

func TestRenderTemplate_MissingKey(t *testing.T) {
	// Go text/template by default does NOT error on missing keys —
	// it renders "<no value>". This is acceptable behavior.
	result, err := renderTemplate(`{"name":"{{.missing}}"}`, map[string]interface{}{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should contain a rendered value (even if it's "<no value>")
	if result == "" {
		t.Fatal("expected non-empty result")
	}
}

func TestRenderTemplate_BoolFalse(t *testing.T) {
	// Go template prints "false" for false bool
	result, err := renderTemplate(`{"enabled":{{.flag}}}`, map[string]interface{}{
		"flag": false,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != `{"enabled":false}` {
		t.Errorf("got %q, want %q", result, `{"enabled":false}`)
	}
}

func TestRenderTemplate_NilParam(t *testing.T) {
	_, err := renderTemplate(`{"name":"{{.x}}"}`, map[string]interface{}{
		"x": nil,
	})
	// nil should render as "<nil>" or empty — either way it shouldn't crash
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && searchString(s, sub)
}

func searchString(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
