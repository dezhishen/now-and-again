package scopes

import "testing"

func TestValid_KnownScope(t *testing.T) {
	for _, s := range All() {
		if !Valid(s) {
			t.Errorf("Valid(%q) = false, want true", s)
		}
	}
}

func TestValid_UnknownScope(t *testing.T) {
	if Valid("unknown:scope") {
		t.Error("unknown scope should be invalid")
	}
	if Valid("") {
		t.Error("empty scope should be invalid")
	}
}

func TestHas_Match(t *testing.T) {
	if !Has([]string{"task:read", "task:write"}, "task:read") {
		t.Error("should match")
	}
}

func TestHas_NoMatch(t *testing.T) {
	if Has([]string{"task:read"}, "task:write") {
		t.Error("should not match")
	}
}

func TestHas_Empty(t *testing.T) {
	if Has(nil, "task:read") {
		t.Error("nil => false")
	}
	if Has([]string{}, "task:read") {
		t.Error("empty => false")
	}
}

func TestAll_MinSize(t *testing.T) {
	if n := len(All()); n < 10 {
		t.Errorf("got %d scopes, want >= 10", n)
	}
}

func TestExpandGroups_Read(t *testing.T) {
	r := ExpandGroups([]string{"read"})
	if len(r) != len(ReadScopes) {
		t.Errorf("got %d scopes, want %d", len(r), len(ReadScopes))
	}
}

func TestExpandGroups_Individual(t *testing.T) {
	r := ExpandGroups([]string{"task:read"})
	if len(r) != 1 || r[0] != "task:read" {
		t.Errorf("got %v", r)
	}
}

func TestRouteScope_Known(t *testing.T) {
	// Match keys that have single space between method and path
	if s := RouteScope("POST", "/api/tasks/:task_id/trigger"); s != "task:write" {
		t.Errorf("got %q", s)
	}
	if s := RouteScope("DELETE", "/api/tasks/:task_id"); s != "task:write" {
		t.Errorf("got %q", s)
	}
	if s := RouteScope("POST", "/api/task-templates"); s != "task:write" {
		t.Errorf("got %q", s)
	}
}

func TestRouteScope_Unknown(t *testing.T) {
	if s := RouteScope("GET", "/api/nonexistent"); s != "" {
		t.Errorf("got %q, want empty", s)
	}
}

func TestDescriptions_AllScopes(t *testing.T) {
	d := Descriptions()
	for _, s := range All() {
		if _, ok := d[s]; !ok {
			t.Errorf("missing desc for %s", s)
		}
	}
}
