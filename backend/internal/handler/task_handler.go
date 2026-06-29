package handler

import (
	"github.com/dezhishen/now-and-again/backend/pkg/contracts"
	"github.com/dezhishen/now-and-again/backend/pkg/types"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TaskHandlers struct {
	Svc contracts.TaskContract
}

func (h *TaskHandlers) Create(c *gin.Context) {
	fid := familyID(c)
	familyID, err := uuid.Parse(fid)
	if err != nil {
		badRequest(c, "invalid family_id")
		return
	}
	req, err := bindJSON[types.CreateTaskRequest](c)
	if err != nil {
		badRequest(c, err.Error())
		return
	}
	t, err := h.Svc.CreateTask(userCtx(c), familyID, req)
	if err != nil {
		serverError(c, err)
		return
	}
	created(c, t)
}

func (h *TaskHandlers) List(c *gin.Context) {
	fid := familyID(c)
	familyID, err := uuid.Parse(fid)
	if err != nil {
		badRequest(c, "invalid family_id")
		return
	}
	tasks, err := h.Svc.ListTasks(userCtx(c), familyID)
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
		result, err := h.Svc.GetTaskWithExtra(userCtx(c), taskID)
		if err != nil {
			serverError(c, err)
			return
		}
		ok(c, result)
		return
	}
	t, err := h.Svc.GetTask(userCtx(c), taskID)
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
	t, err := h.Svc.UpdateTask(userCtx(c), taskID, req)
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
	if err := h.Svc.DeleteTask(userCtx(c), taskID); err != nil {
		serverError(c, err)
		return
	}
	ok(c, gin.H{"message": "task deleted"})
}

func (h *TaskHandlers) SetEnabled(c *gin.Context) {
	taskID, err := paramUUID(c, "task_id")
	if err != nil {
		badRequest(c, "invalid task_id")
		return
	}
	var body struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		badRequest(c, err.Error())
		return
	}
	t, err := h.Svc.SetTaskEnabled(userCtx(c), taskID, body.Enabled)
	if err != nil {
		serverError(c, err)
		return
	}
	ok(c, t)
}

func (h *TaskHandlers) Trigger(c *gin.Context) {
	taskID, err := paramUUID(c, "task_id")
	if err != nil {
		badRequest(c, "invalid task_id")
		return
	}
	if err := h.Svc.TriggerTask(userCtx(c), taskID); err != nil {
		serverError(c, err)
		return
	}
	ok(c, gin.H{"message": "todo generated"})
}
