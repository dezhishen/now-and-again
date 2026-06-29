package handler

import (
	"github.com/dezhishen/now-and-again/backend/pkg/contracts"
	"github.com/dezhishen/now-and-again/backend/pkg/types"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TodoHandlers struct {
	Svc contracts.TodoContract
}

func (h *TodoHandlers) ListTodos(c *gin.Context) {
	familyID := familyID(c)
	status := c.Query("status")
	groupID := c.Query("group_id")
	todos, err := h.Svc.ListTodos(userCtx(c), uuid.MustParse(familyID), groupID, status)
	if err != nil {
		serverError(c, err)
		return
	}
	ok(c, todos)
}

func (h *TodoHandlers) GetTodo(c *gin.Context) {
	todoID, err := paramUUID(c, "todo_id")
	if err != nil {
		badRequest(c, "invalid todo_id")
		return
	}
	withExtra := c.Query("with_extra") == "true"
	if withExtra {
		result, err := h.Svc.GetTodoWithExtra(userCtx(c), todoID)
		if err != nil {
			serverError(c, err)
			return
		}
		ok(c, result)
		return
	}
	todo, err := h.Svc.GetTodo(userCtx(c), todoID)
	if err != nil {
		serverError(c, err)
		return
	}
	ok(c, todo)
}

func (h *TodoHandlers) CompleteTodo(c *gin.Context) {
	todoID, err := paramUUID(c, "todo_id")
	if err != nil {
		badRequest(c, "invalid todo_id")
		return
	}
	req, err := bindJSON[types.CompleteTodoRequest](c)
	if err != nil {
		validationError(c, err)
		return
	}
	todo, err := h.Svc.CompleteTodo(userCtx(c), todoID, req)
	if err != nil {
		serverError(c, err)
		return
	}
	ok(c, todo)
}
