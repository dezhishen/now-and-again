// Package state provides action-id-keyed temporary state storage for the CLI.
//
// When name resolution is ambiguous (e.g., "大人" matches both "大人一" and "大人二"),
// the CLI saves the full command context and candidate options to a tmp file.
// The caller (human or AI) can then inspect the candidates and retry with an
// exact name or ID. The state file is automatically cleaned up after the action
// completes or when explicitly discarded.
//
// Storage location (cross-platform):
//
//	Unix:    /tmp/na-action-<actionID>.json
//	Windows: %TEMP%\na-action-<actionID>.json
package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ─── Types ────────────────────────────────────────────────────────

// ActionState holds the context of an in-progress CLI action that needs
// user input to resolve ambiguous entities.
type ActionState struct {
	ActionID string `json:"action_id"`
	Step     string `json:"step"` // "resolve_group", "resolve_location"

	// Full command context so the user can retry without retyping everything.
	Command string            `json:"command"` // e.g. "task create"
	Args    map[string]string `json:"args"`    // all flag values from the original invocation

	// Candidates for resolution.
	Candidates []EntityOption `json:"candidates"`

	// Already-resolved values.
	Resolved map[string]string `json:"resolved,omitempty"`
}

// EntityOption represents one selectable entity (group or location).
type EntityOption struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ─── File paths ──────────────────────────────────────────────────

// statePath returns the tmp file path for a given action ID.
func statePath(actionID string) string {
	dir := os.TempDir() // cross-platform: /tmp on Unix, %TEMP% on Windows
	return filepath.Join(dir, "na-action-"+actionID+".json")
}

// ─── Save / Load / Delete ────────────────────────────────────────

// Save writes the action state to a temporary file.
func Save(s *ActionState) error {
	path := statePath(s.ActionID)
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("write state: %w", err)
	}
	return nil
}

// Load reads the action state from the temporary file.
// Returns nil if the file does not exist.
func Load(actionID string) (*ActionState, error) {
	path := statePath(actionID)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read state: %w", err)
	}
	var s ActionState
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parse state: %w", err)
	}
	return &s, nil
}

// Delete removes the action state file. Safe to call even if the file doesn't exist.
func Delete(actionID string) {
	path := statePath(actionID)
	_ = os.Remove(path)
}

// Exists checks whether a state file exists for the given action ID.
func Exists(actionID string) bool {
	path := statePath(actionID)
	_, err := os.Stat(path)
	return err == nil
}

// ─── Resolution helpers ──────────────────────────────────────────

// ResolveAmbiguousGroup generates an error message and saves state when
// multiple groups match the given input.
func ResolveAmbiguousGroup(actionID string, cmd string, args map[string]string, input string, candidates []EntityOption) error {
	s := &ActionState{
		ActionID:   actionID,
		Step:       "resolve_group",
		Command:    cmd,
		Args:       args,
		Candidates: candidates,
	}
	if err := Save(s); err != nil {
		return fmt.Errorf("save state: %w", err)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("找到 %d 个匹配的小组 %q:\n", len(candidates), input))
	for i, c := range candidates {
		sb.WriteString(fmt.Sprintf("  %d. %s  (%s)\n", i+1, c.Name, c.ID[:8]))
	}
	sb.WriteString(fmt.Sprintf("\n→ 请使用精确名称或 --group-id 重试，如:\n"))
	if len(candidates) > 0 {
		sb.WriteString(fmt.Sprintf("  na --action-id %s %s --group \"%s\" ...\n", actionID, cmd, candidates[0].Name))
		sb.WriteString(fmt.Sprintf("  na --action-id %s %s --group-id %s ...\n", actionID, cmd, candidates[0].ID))
	}
	sb.WriteString(fmt.Sprintf("\n状态已保存至: %s", statePath(actionID)))
	return fmt.Errorf("%s", sb.String())
}

// ResolveAmbiguousLocation generates an error message and saves state when
// multiple locations match the given input.
func ResolveAmbiguousLocation(actionID string, cmd string, args map[string]string, input string, candidates []EntityOption) error {
	s := &ActionState{
		ActionID:   actionID,
		Step:       "resolve_location",
		Command:    cmd,
		Args:       args,
		Candidates: candidates,
	}
	if err := Save(s); err != nil {
		return fmt.Errorf("save state: %w", err)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("找到 %d 个匹配的地址 %q:\n", len(candidates), input))
	for i, c := range candidates {
		sb.WriteString(fmt.Sprintf("  %d. %s  (%s)\n", i+1, c.Name, c.ID[:8]))
	}
	sb.WriteString(fmt.Sprintf("\n→ 请使用精确名称或 --location-id 重试，如:\n"))
	if len(candidates) > 0 {
		sb.WriteString(fmt.Sprintf("  na --action-id %s %s --location \"%s\" ...\n", actionID, cmd, candidates[0].Name))
		sb.WriteString(fmt.Sprintf("  na --action-id %s %s --location-id %s ...\n", actionID, cmd, candidates[0].ID))
	}
	sb.WriteString(fmt.Sprintf("\n状态已保存至: %s", statePath(actionID)))
	return fmt.Errorf("%s", sb.String())
}

// CleanupIfDone deletes the state file if the current step is "done"
// (i.e., no more pending resolutions).
func CleanupIfDone(actionID string) {
	_ = os.Remove(statePath(actionID))
}
