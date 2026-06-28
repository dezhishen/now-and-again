package handler

import (
	"github.com/dezhishen/now-and-again/backend/pkg/contracts"
	"github.com/dezhishen/now-and-again/backend/pkg/types"
	"github.com/gin-gonic/gin"
)

type TaskHandlers struct {
	TaskSvc contracts.TaskContract
	TodoSvc contracts.TodoContract
	LogSvc  contracts.LogContract
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
	t, err := h.TaskSvc.CreateTask(userCtx(c), familyID, req)
	if err != nil {
		serverError(c, err)
		return
	}
	created(c, t)
}

func (h *TaskHandlers) List(c *gin.Context) {
	familyID, err := paramUUID(c, "family_id")
	if err != nil {
		badRequest(c, "invalid family_id")
		return
	}
	tasks, err := h.TaskSvc.ListTasks(userCtx(c), familyID)
	if err != nil {
		serverError(c, err)
		return
	}
	ok(c, tasks)
}

func (h *TaskHandlers) Get(c *gin.Context) {
	taskID, err := paramUUID(c, "task_id")
	if err != nil {
		badRequest(c, "invalid task_id")
		return
	}
	withExtra := c.Query("with_extra") == "true"
	if withExtra {
		result, err := h.TaskSvc.GetTaskWithExtra(userCtx(c), taskID)
		if err != nil {
			serverError(c, err)
			return
		}
		ok(c, result)
		return
	}
	t, err := h.TaskSvc.GetTask(userCtx(c), taskID)
	if err != nil {
		serverError(c, err)
		return
	}
	ok(c, t)
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
	t, err := h.TaskSvc.UpdateTask(userCtx(c), taskID, req)
	if err != nil {
		serverError(c, err)
		return
	}
	ok(c, t)
}

func (h *TaskHandlers) Delete(c *gin.Context) {
	taskID, err := paramUUID(c, "task_id")
	if err != nil {
		badRequest(c, "invalid task_id")
		return
	}
	if err := h.TaskSvc.DeleteTask(userCtx(c), taskID); err != nil {
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
	todos, err := h.TodoSvc.ListTodos(userCtx(c), familyID, groupID, status)
	if err != nil {
		serverError(c, err)
		return
	}
	ok(c, todos)
}

func (h *TaskHandlers) GetTodo(c *gin.Context) {
	todoID, err := paramUUID(c, "todo_id")
	if err != nil {
		badRequest(c, "invalid todo_id")
		return
	}
	withExtra := c.Query("with_extra") == "true"
	if withExtra {
		result, err := h.TodoSvc.GetTodoWithExtra(userCtx(c), todoID)
		if err != nil {
			serverError(c, err)
			return
		}
		ok(c, result)
		return
	}
	todo, err := h.TodoSvc.GetTodo(userCtx(c), todoID)
	if err != nil {
		serverError(c, err)
		return
	}
	ok(c, todo)
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
	todo, err := h.TodoSvc.CompleteTodo(userCtx(c), todoID, req)
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
	offset := queryInt(c, "offset", 0)
	userOnly := c.Query("type") == "user"
	logs, err := h.LogSvc.ListLogs(userCtx(c), taskID, limit, offset, userOnly)
	if err != nil {
		serverError(c, err)
		return
	}
	ok(c, logs)
}

func (h *TaskHandlers) Trigger(c *gin.Context) {
	taskID, err := paramUUID(c, "task_id")
	if err != nil {
		badRequest(c, "invalid task_id")
		return
	}
	if err := h.TaskSvc.TriggerTask(userCtx(c), taskID); err != nil {
		serverError(c, err)
		return
	}
	ok(c, gin.H{"message": "todo generated"})
}

func (h *TaskHandlers) GetCalendar(c *gin.Context) {
	familyID := c.Param("family_id")
	year := queryInt(c, "year", 0)
	month := queryInt(c, "month", 0)
	groupID := c.Query("group_id")

	if year <= 0 || month <= 0 || month > 12 {
		badRequest(c, "year and month are required (1-12)")
		return
	}

	days, err := h.TaskSvc.GetCalendar(userCtx(c), familyID, year, month, groupID)
	if err != nil {
		serverError(c, err)
		return
	}
	ok(c, days)
}
