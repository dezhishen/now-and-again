package client

import (
	"context"
	"strconv"

	"github.com/dezhishen/now-and-again/backend/pkg/types"
	"github.com/dezhishen/now-and-again/backend/pkg/types/task"
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

func (c *TaskClient) Create(familyID string, req *task.CreateTaskRequest) (*task.Task, error) {
	var t task.Task
	if err := c.http.do("POST", "/api/families/"+familyID+"/tasks", req, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

func (c *TaskClient) List(familyID string) ([]task.Task, error) {
	var tasks []task.Task
	if err := c.http.do("GET", "/api/families/"+familyID+"/tasks", nil, &tasks); err != nil {
		return nil, err
	}
	return tasks, nil
}

func (c *TaskClient) Update(taskID string, req *task.UpdateTaskRequest) (*task.Task, error) {
	var t task.Task
	if err := c.http.do("PUT", "/api/tasks/"+taskID, req, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

func (c *TaskClient) Delete(taskID string) error {
	return c.http.do("DELETE", "/api/tasks/"+taskID, nil, nil)
}

func (c *TaskClient) ListTodosSimple(familyID, status string) ([]types.Todo, error) {
	path := "/api/families/" + familyID + "/todos"
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

func (c *TaskClient) CreateTask(_ context.Context, familyID uuid.UUID, req *task.CreateTaskRequest) (*task.Task, error) {
	return c.Create(familyID.String(), req)
}
func (c *TaskClient) UpdateTask(_ context.Context, taskID uuid.UUID, req *task.UpdateTaskRequest) (*task.Task, error) {
	return c.Update(taskID.String(), req)
}
func (c *TaskClient) DeleteTask(_ context.Context, taskID uuid.UUID) error {
	return c.Delete(taskID.String())
}
func (c *TaskClient) ListTasks(_ context.Context, familyID uuid.UUID) ([]task.Task, error) {
	return c.List(familyID.String())
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
	path := "/api/tasks/" + taskID.String() + "/logs?limit=" + strconv.Itoa(limit)
	if userOnly {
		path += "&type=user"
	}
	var logs []types.TaskLog
	if err := c.http.do("GET", path, nil, &logs); err != nil {
		return nil, err
	}
	return logs, nil
}
