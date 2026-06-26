package handler

import (
	"github.com/dezhishen/now-and-again/backend/internal/service"
	"github.com/dezhishen/now-and-again/shared/types"
	"github.com/gin-gonic/gin"
)

type TaskHandlers struct {
	Svc *service.TaskService
}

func (h *TaskHandlers) Create(c *gin.Context) {
	familyID, err := paramUUID(c, "family_id")
	if err != nil {
		badRequest(c, "invalid family_id")
		return
	}
	req, err := bindJSON[types.CreateTaskRequest](c)
	if err != nil {
		badRequest(c, err.Error())
		return
	}
	task, err := h.Svc.Create(userCtx(c), familyID, req)
	if err != nil {
		serverError(c, err)
		return
	}
	created(c, task)
}

func (h *TaskHandlers) List(c *gin.Context) {
	familyID, err := paramUUID(c, "family_id")
	if err != nil {
		badRequest(c, "invalid family_id")
		return
	}
	tasks, err := h.Svc.List(userCtx(c), familyID)
	if err != nil {
		serverError(c, err)
		return
	}
	ok(c, tasks)
}

func (h *TaskHandlers) Update(c *gin.Context) {
	taskID, err := paramUUID(c, "task_id")
	if err != nil {
		badRequest(c, "invalid task_id")
		return
	}
	req, err := bindJSON[types.UpdateTaskRequest](c)
	if err != nil {
		badRequest(c, err.Error())
		return
	}
	task, err := h.Svc.Update(userCtx(c), taskID, req)
	if err != nil {
		serverError(c, err)
		return
	}
	ok(c, task)
}

func (h *TaskHandlers) Delete(c *gin.Context) {
	taskID, err := paramUUID(c, "task_id")
	if err != nil {
		badRequest(c, "invalid task_id")
		return
	}
	if err := h.Svc.Delete(userCtx(c), taskID); err != nil {
		serverError(c, err)
		return
	}
	ok(c, gin.H{"message": "task deleted"})
}

func (h *TaskHandlers) ListTodos(c *gin.Context) {
	familyID, err := paramUUID(c, "family_id")
	if err != nil {
		badRequest(c, "invalid family_id")
		return
	}
	status := c.Query("status")
	groupID := c.Query("group_id")
	todos, err := h.Svc.ListTodos(userCtx(c), familyID, groupID, status)
	if err != nil {
		serverError(c, err)
		return
	}
	ok(c, todos)
}

func (h *TaskHandlers) CompleteTodo(c *gin.Context) {
	todoID, err := paramUUID(c, "todo_id")
	if err != nil {
		badRequest(c, "invalid todo_id")
		return
	}
	req, err := bindJSON[types.CompleteTodoRequest](c)
	if err != nil {
		badRequest(c, err.Error())
		return
	}
	todo, err := h.Svc.CompleteTodo(userCtx(c), todoID, req)
	if err != nil {
		serverError(c, err)
		return
	}
	ok(c, todo)
}

func (h *TaskHandlers) ListLogs(c *gin.Context) {
	taskID, err := paramUUID(c, "task_id")
	if err != nil {
		badRequest(c, "invalid task_id")
		return
	}
	limit := queryInt(c, "limit", 50)
	logs, err := h.Svc.ListLogs(userCtx(c), taskID, limit)
	if err != nil {
		serverError(c, err)
		return
	}
	ok(c, logs)
}
