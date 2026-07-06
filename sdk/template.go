// Package sdk — Template operations.
//
// ⚠️ TIMEZONE POLICY: Template data is timezone-agnostic by design.
// "早上8点" means 8 AM everywhere in the world. Templates do not carry
// timezone semantics. Therefore:
//
//   - Template loading (ListTemplates, RenderTemplate) must NEVER apply
//     timezone conversion.
//   - CreateTaskFromTemplate must use the low-level na.Task.CreateTask()
//     which sends schedule_data as-is (already UTC from backend rendering).
//   - Do NOT route template-created tasks through na.CreateTask() or
//     na.UpdateTask() for schedule_data — those wrappers apply local→UTC
//     conversion which would double-convert template times.
package sdk

import (
	"context"
	"fmt"

	"github.com/dezhishen/now-and-again/backend/pkg/types"
	"github.com/google/uuid"
)

// ─── Template operations ──────────────────────────────────────────

// ListTemplates returns all templates visible to the active family.
func (na *NA) ListTemplates(ctx context.Context, kind string) ([]types.TaskTemplate, error) {
	fid, err := na.requireFamilyID()
	if err != nil {
		return nil, err
	}
	return na.callListTemplates(ctx, fid, kind)
}

// RenderTemplate renders a template with the given parameters.
func (na *NA) RenderTemplate(ctx context.Context, code string, params map[string]interface{}) (*types.RenderedTask, error) {
	fid, err := na.requireFamilyID()
	if err != nil {
		return nil, err
	}
	return na.callRenderTemplate(ctx, fid, code, params)
}

// CreateTaskFromTemplate renders a template and creates a task from it.
// This is the most common workflow: template → task.
//
// Example:
//
//	task, err := na.CreateTaskFromTemplate(ctx, "weekly_cleaning", map[string]interface{}{
//	    "area_name": "客厅",
//	})
func (na *NA) CreateTaskFromTemplate(ctx context.Context, code string, params map[string]interface{}) (*types.Task, error) {
	// Step 1: Render
	rendered, err := na.RenderTemplate(ctx, code, params)
	if err != nil {
		return nil, fmt.Errorf("render template: %w", err)
	}

	// Step 2: Extract task defaults
	td, ok := rendered.TaskDefaults.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid rendered task defaults")
	}

	// Step 3: Build CreateTaskRequest
	req := &types.CreateTaskRequest{
		Task:  types.Task{},
		Extra: rendered.ExtraSchema,
	}

	if name, ok := td["name"].(string); ok {
		req.Task.Name = name
	}
	if st, ok := td["schedule_type"].(string); ok {
		req.Task.ScheduleType = st
	}
	if sd, ok := td["schedule_data"]; ok {
		req.Task.ScheduleData = sd
	}
	if kind, ok := td["kind"].(string); ok {
		req.Task.Kind = kind
	}
	if gid, ok := td["group_id"].(string); ok {
		req.Task.GroupID = gid
	}
	if lid, ok := td["location_id"].(string); ok {
		req.Task.LocationID = lid
	}

	// Step 4: Create — use low-level call to skip timezone conversion.
	// Template rendering already returns UTC times from the backend.
	fid, err := na.requireFamilyID()
	if err != nil {
		return nil, err
	}
	return na.Task.CreateTask(ctx, fid, req)
}

// ─── Internal HTTP helpers (mirror the backend handler paths) ─────

func (na *NA) callListTemplates(_ context.Context, _ uuid.UUID, kind string) ([]types.TaskTemplate, error) {
	path := "/api/task-templates"
	if kind != "" {
		path += "?kind=" + kind
	}
	var result []types.TaskTemplate
	if err := na.http.Do("GET", path, nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (na *NA) callRenderTemplate(_ context.Context, _ uuid.UUID, code string, params map[string]interface{}) (*types.RenderedTask, error) {
	path := fmt.Sprintf("/api/task-templates/%s/render", code)
	var result types.RenderedTask
	if err := na.http.Do("POST", path, params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
