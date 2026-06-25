package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/dezhishen/now-and-again/shared/types"
)

func (h *LogHandlers) ListByTask(c *gin.Context) {
	taskID, err := paramUUID(c, "task_id")
	if err != nil {
		badRequest(c, "invalid task_id")
		return
	}
	logs, err := h.C.ListByTask(userCtx(c), taskID)
	if err != nil {
		serverError(c, err)
		return
	}
	ok(c, logs)
}
func (h *LogHandlers) AddComment(c *gin.Context) {
	taskID, err := paramUUID(c, "task_id")
	if err != nil {
		badRequest(c, "invalid task_id")
		return
	}
	req, err := bindJSON[types.AddCommentRequest](c)
	if err != nil {
		badRequest(c, err.Error())
		return
	}
	log, err := h.C.AddComment(userCtx(c), taskID, req)
	if err != nil {
		serverError(c, err)
		return
	}
	created(c, log)
}
