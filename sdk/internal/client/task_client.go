package client

import (
	"context"
	"fmt"
	"net/url"

	"github.com/dezhishen/now-and-again/backend/pkg/types"
	"github.com/google/uuid"
)

// Compile-time check: TaskClient satisfies the core task contract.

// TaskClient provides CLI access to task endpoints.
type TaskClient struct {
	http *HTTPClient
}

func NewTaskClient(http *HTTPClient) *TaskClient {
	return &TaskClient{http: http}
}

func (c *TaskClient) Create(familyID string, req *types.CreateTaskRequest) (*types.Task, error) {
	var t types.Task
	if err := c.http.do("POST", "/api/tasks", req, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

func (c *TaskClient) List(familyID string) ([]types.Task, error) {
	var tasks []types.Task
	if err := c.http.do("GET", "/api/tasks", nil, &tasks); err != nil {
		return nil, err
	}
	return tasks, nil
}

// ListFiltered returns tasks with optional archived/disabled/name filters via query params.
func (c *TaskClient) ListFiltered(familyID string, includeArchived, includeDisabled bool, name string) ([]types.Task, error) {
	path := "/api/tasks"
	parts := make([]string, 0, 3)
	if includeArchived {
		parts = append(parts, "archived=true")
	}
	if includeDisabled {
		parts = append(parts, "disabled=true")
	}
	if name != "" {
		parts = append(parts, "name="+url.QueryEscape(name))
	}
	for i, p := range parts {
		if i == 0 {
			path += "?" + p
		} else {
			path += "&" + p
		}
	}
	var tasks []types.Task
	if err := c.http.do("GET", path, nil, &tasks); err != nil {
		return nil, err
	}
	return tasks, nil
}

func (c *TaskClient) Update(taskID string, req *types.UpdateTaskRequest) (*types.Task, error) {
	var t types.Task
	if err := c.http.do("PUT", "/api/tasks/"+taskID, req, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

func (c *TaskClient) Delete(taskID string) error {
	return c.http.do("DELETE", "/api/tasks/"+taskID, nil, nil)
}

// Get returns a single task by ID.
func (c *TaskClient) Get(taskID string) (*types.Task, error) {
	var t types.Task
	if err := c.http.do("GET", "/api/tasks/"+taskID, nil, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

// GetWithExtra returns a single task with its extra data (chain steps, inspection check items, etc.).
func (c *TaskClient) GetWithExtra(taskID string) (*types.TaskWithExtra, error) {
	var t types.TaskWithExtra
	if err := c.http.do("GET", "/api/tasks/"+taskID+"?with_extra=true", nil, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

func (c *TaskClient) ListTodosSimple(familyID, status string) ([]types.Todo, error) {
	path := "/api/todos"
	if status != "" {
		path += "?status=" + status
	}
	var todos []types.Todo
	if err := c.http.do("GET", path, nil, &todos); err != nil {
		return nil, err
	}
	return todos, nil
}

func (c *TaskClient) CompleteTodoSimple(todoID, status string) (*types.Todo, error) {
	var t types.Todo
	req := &types.CompleteTodoRequest{
		Todo: &types.Todo{Status: status},
	}
	if err := c.http.do("PUT", "/api/todos/"+todoID, req, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

// ─── TaskContract delegates ──────────────────────────────────────

func (c *TaskClient) CreateTask(_ context.Context, familyID uuid.UUID, req *types.CreateTaskRequest) (*types.Task, error) {
	return c.Create(familyID.String(), req)
}
func (c *TaskClient) GetTask(_ context.Context, taskID uuid.UUID) (*types.Task, error) {
	return c.Get(taskID.String())
}
func (c *TaskClient) GetTaskWithExtra(_ context.Context, taskID uuid.UUID) (*types.TaskWithExtra, error) {
	return c.GetWithExtra(taskID.String())
}
func (c *TaskClient) UpdateTask(_ context.Context, taskID uuid.UUID, req *types.UpdateTaskRequest) (*types.Task, error) {
	return c.Update(taskID.String(), req)
}
func (c *TaskClient) DeleteTask(_ context.Context, taskID uuid.UUID) error {
	return c.Delete(taskID.String())
}
func (c *TaskClient) ListTasks(_ context.Context, familyID uuid.UUID) ([]types.Task, error) {
	return c.List(familyID.String())
}
func (c *TaskClient) ListTasksFiltered(_ context.Context, familyID uuid.UUID, includeArchived, includeDisabled bool, name string) ([]types.Task, error) {
	return c.ListFiltered(familyID.String(), includeArchived, includeDisabled, name)
}
func (c *TaskClient) TriggerTask(_ context.Context, taskID uuid.UUID) error {
	return c.http.do("POST", "/api/tasks/"+taskID.String()+"/trigger", nil, nil)
}
func (c *TaskClient) ListTodos(_ context.Context, familyID uuid.UUID, groupID, status string) ([]types.Todo, error) {
	return c.ListTodosSimple(familyID.String(), status)
}
func (c *TaskClient) CompleteTodo(_ context.Context, todoID uuid.UUID, req *types.CompleteTodoRequest) (*types.Todo, error) {
	var t types.Todo
	if err := c.http.do("PUT", "/api/todos/"+todoID.String(), req, &t); err != nil {
		return nil, err
	}
	return &t, nil
}
func (c *TaskClient) ListTaskLogs(_ context.Context, taskID uuid.UUID, limit int, userOnly bool) ([]types.TaskLog, error) {
	path := fmt.Sprintf("/api/tasks/%s/logs?limit=%d", taskID.String(), limit)
	if userOnly {
		path += "&type=user"
	}
	var logs []types.TaskLog
	if err := c.http.do("GET", path, nil, &logs); err != nil {
		return nil, err
	}
	return logs, nil
}
