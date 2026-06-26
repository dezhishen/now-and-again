package client

import (
	"github.com/dezhishen/now-and-again/shared/types"
)

// TaskClient provides CLI access to task endpoints.
type TaskClient struct {
	http *HTTPClient
}

func NewTaskClient(http *HTTPClient) *TaskClient {
	return &TaskClient{http: http}
}

func (c *TaskClient) Create(familyID string, req *types.CreateTaskRequest) (*types.TaskTemplate, error) {
	var t types.TaskTemplate
	if err := c.http.do("POST", "/api/families/"+familyID+"/tasks", req, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

func (c *TaskClient) List(familyID string) ([]types.TaskTemplate, error) {
	var tasks []types.TaskTemplate
	if err := c.http.do("GET", "/api/families/"+familyID+"/tasks", nil, &tasks); err != nil {
		return nil, err
	}
	return tasks, nil
}

func (c *TaskClient) Update(taskID string, req *types.UpdateTaskRequest) (*types.TaskTemplate, error) {
	var t types.TaskTemplate
	if err := c.http.do("PUT", "/api/tasks/"+taskID, req, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

func (c *TaskClient) Delete(taskID string) error {
	return c.http.do("DELETE", "/api/tasks/"+taskID, nil, nil)
}

func (c *TaskClient) ListTodos(familyID, status string) ([]types.Todo, error) {
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

func (c *TaskClient) CompleteTodo(todoID, status string) (*types.Todo, error) {
	var t types.Todo
	if err := c.http.do("PUT", "/api/todos/"+todoID, &types.CompleteTodoRequest{Status: status}, &t); err != nil {
		return nil, err
	}
	return &t, nil
}
