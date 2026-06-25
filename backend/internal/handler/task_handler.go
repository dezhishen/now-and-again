package handler

import (
	"github.com/dezhishen/now-and-again/backend/internal/scheduler"
	"github.com/dezhishen/now-and-again/shared/types"
	"github.com/gin-gonic/gin"
)

func (h *TaskHandlers) Create(c *gin.Context) {
	req, err := bindJSON[types.CreateTaskRequest](c)
	if err != nil {
		badRequest(c, err.Error())
		return
	}
	task, err := h.C.Create(userCtx(c), req)
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
	status := queryStatus(c, "status")
	assigneeID := queryUUID(c, "assignee_id")
	page := queryInt(c, "page", 1)
	pageSize := queryInt(c, "page_size", 20)
	tasks, total, err := h.C.List(userCtx(c), familyID, status, assigneeID, page, pageSize)
	if err != nil {
		serverError(c, err)
		return
	}
	paged(c, tasks, page, pageSize, total)
}
func (h *TaskHandlers) Get(c *gin.Context) {
	taskID, err := paramUUID(c, "task_id")
	if err != nil {
		badRequest(c, "invalid task_id")
		return
	}
	task, err := h.C.Get(userCtx(c), taskID)
	if err != nil {
		serverError(c, err)
		return
	}
	ok(c, task)
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
	task, err := h.C.Update(userCtx(c), taskID, req)
	if err != nil {
		serverError(c, err)
		return
	}
	ok(c, task)
}
func (h *TaskHandlers) SetAssignees(c *gin.Context) {
	taskID, err := paramUUID(c, "task_id")
	if err != nil {
		badRequest(c, "invalid task_id")
		return
	}
	req, err := bindJSON[types.SetAssigneesRequest](c)
	if err != nil {
		badRequest(c, err.Error())
		return
	}
	assignees, err := h.C.SetAssignees(userCtx(c), taskID, req)
	if err != nil {
		serverError(c, err)
		return
	}
	ok(c, assignees)
}
func (h *TaskHandlers) AddDependency(c *gin.Context) {
	taskID, err := paramUUID(c, "task_id")
	if err != nil {
		badRequest(c, "invalid task_id")
		return
	}
	req, err := bindJSON[types.CreateDependencyRequest](c)
	if err != nil {
		badRequest(c, err.Error())
		return
	}
	dep, err := h.C.AddDependency(userCtx(c), taskID, req)
	if err != nil {
		serverError(c, err)
		return
	}
	created(c, dep)
}
func (h *TaskHandlers) RemoveDependency(c *gin.Context) {
	taskID, err := paramUUID(c, "task_id")
	if err != nil {
		badRequest(c, "invalid task_id")
		return
	}
	depID, err := paramUUID(c, "dep_id")
	if err != nil {
		badRequest(c, "invalid dep_id")
		return
	}
	if err := h.C.RemoveDependency(userCtx(c), taskID, depID); err != nil {
		serverError(c, err)
		return
	}
	noContent(c)
}

func (h *TaskHandlers) ListScheduleTypes(c *gin.Context) {
	ok(c, scheduler.ListHandlerDefs())
}
