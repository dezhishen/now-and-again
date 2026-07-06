package taskkind

import (
	"testing"
)

func TestNormalizeKind_Empty(t *testing.T) {
	if got := NormalizeKind(""); got != "simple" {
		t.Errorf("NormalizeKind('') = %q, want 'simple'", got)
	}
}

func TestNormalizeKind_NonEmpty(t *testing.T) {
	tests := []string{"simple", "inspection", "chain", "custom_kind"}
	for _, input := range tests {
		got := NormalizeKind(input)
		if got != input {
			t.Errorf("NormalizeKind(%q) = %q, want %q", input, got, input)
		}
	}
}

func TestIsDefaultKind(t *testing.T) {
	if !IsDefaultKind("") {
		t.Error("IsDefaultKind('') = false, want true")
	}
	if !IsDefaultKind("simple") {
		t.Error("IsDefaultKind('simple') = false, want true")
	}
	if IsDefaultKind("inspection") {
		t.Error("IsDefaultKind('inspection') = true, want false")
	}
}

func TestResolveDispatchKind(t *testing.T) {
	// When ownerKind is empty/default, fall back to taskKind
	if got := ResolveDispatchKind("", "inspection"); got != "inspection" {
		t.Errorf("ResolveDispatchKind('', 'inspection') = %q, want 'inspection'", got)
	}
	if got := ResolveDispatchKind("simple", "chain"); got != "chain" {
		t.Errorf("ResolveDispatchKind('simple', 'chain') = %q, want 'chain'", got)
	}
	// When ownerKind is non-default, use it
	if got := ResolveDispatchKind("inspection", "simple"); got != "inspection" {
		t.Errorf("ResolveDispatchKind('inspection', 'simple') = %q, want 'inspection'", got)
	}
}

func TestDefaultKind(t *testing.T) {
	if DefaultKind != "simple" {
		t.Errorf("DefaultKind = %q, want 'simple'", DefaultKind)
	}
}
