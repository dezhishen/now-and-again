package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/dezhishen/now-and-again/shared/types"
)

func (h *NotificationHandlers) List(c *gin.Context) {
	page := queryInt(c, "page", 1)
	pageSize := queryInt(c, "page_size", 20)
	notifs, total, err := h.C.List(userCtx(c), page, pageSize)
	if err != nil {
		serverError(c, err)
		return
	}
	paged(c, notifs, page, pageSize, total)
}
func (h *NotificationHandlers) UpsertChannelConfig(c *gin.Context) {
	req, err := bindJSON[types.UpsertUserChannelRequest](c)
	if err != nil {
		badRequest(c, err.Error())
		return
	}
	cfg, err := h.C.UpsertChannelConfig(userCtx(c), req)
	if err != nil {
		serverError(c, err)
		return
	}
	ok(c, cfg)
}
func (h *NotificationHandlers) ListTemplates(c *gin.Context) {
	familyID, err := paramUUID(c, "family_id")
	if err != nil {
		badRequest(c, "invalid family_id")
		return
	}
	templates, err := h.C.ListTemplates(userCtx(c), familyID)
	if err != nil {
		serverError(c, err)
		return
	}
	ok(c, templates)
}
func (h *NotificationHandlers) UpsertTemplate(c *gin.Context) {
	familyID, err := paramUUID(c, "family_id")
	if err != nil {
		badRequest(c, "invalid family_id")
		return
	}
	req, err := bindJSON[types.UpsertTemplateRequest](c)
	if err != nil {
		badRequest(c, err.Error())
		return
	}
	tmpl, err := h.C.UpsertTemplate(userCtx(c), familyID, req)
	if err != nil {
		serverError(c, err)
		return
	}
	ok(c, tmpl)
}
