package inspection

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/dezhishen/now-and-again/backend/pkg/taskkind"
)

// RegisterRoutes wires inspection-specific API endpoints.
// POST /api/tasks/:task_id/inspection
func (Handler) RegisterRoutes(router *gin.RouterGroup, ops *taskkind.Ops) {
	router.POST("/", func(c *gin.Context) {
		taskID := c.Param("task_id")
		userID, _ := c.Get("user_id")

		var req struct {
			TodoID     string               `json:"todo_id"`
			Selections []taskkind.Selection `json:"selections"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid request"})
			return
		}

		todo, err := ops.Repo.FindTodoByID(req.TodoID)
		if err != nil || todo.TaskID != taskID {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "todo not found"})
			return
		}

		if err := ops.Repo.CompleteTodo(req.TodoID, userID.(string), "done", ""); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
			return
		}

		var h Handler
		if err := h.OnInspect(ops, todo, req.Selections, userID.(string)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
			return
		}

		if todo.Task.ScheduleType == "once" {
			ops.Repo.DisableTask(todo.TaskID)
			ops.Scheduler.RemoveJob(todo.TaskID)
		}

		todo, _ = ops.Repo.FindTodoByID(req.TodoID)
		c.JSON(http.StatusOK, gin.H{"success": true, "data": todo})
	})
}
