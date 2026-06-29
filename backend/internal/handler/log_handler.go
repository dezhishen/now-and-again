package handler

import (
	"github.com/dezhishen/now-and-again/backend/pkg/contracts"
	"github.com/gin-gonic/gin"
)

type LogHandlers struct {
	Svc contracts.LogContract
}

func (h *LogHandlers) ListLogs(c *gin.Context) {
	taskID, err := paramUUID(c, "task_id")
	if err != nil {
		badRequest(c, "invalid task_id")
		return
	}
	limit := queryInt(c, "limit", 50)
	offset := queryInt(c, "offset", 0)
	userOnly := c.Query("type") == "user"
	logs, err := h.Svc.ListLogs(userCtx(c), taskID, limit, offset, userOnly)
	if err != nil {
		serverError(c, err)
		return
	}
	ok(c, logs)
}
