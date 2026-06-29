package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ─── Floor Plan ──────────────────────────────────────────────────

func (h *FloorPlanHandlers) Upload(c *gin.Context) {
	fid := familyID(c)
	familyID, err := uuid.Parse(fid)
	if err != nil {
		badRequest(c, "invalid family_id")
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		badRequest(c, "file is required")
		return
	}
	defer file.Close()

	label := c.PostForm("label")
	if label == "" {
		label = "1F"
	}
	isCover := c.PostForm("is_cover") == "true"

	fp, err := h.C.Upload(userCtx(c), familyID, label, isCover, file, header)
	if err != nil {
		serverError(c, err)
		return
	}
	created(c, fp)
}

func (h *FloorPlanHandlers) ListByFamily(c *gin.Context) {
	fid := familyID(c)
	familyID, err := uuid.Parse(fid)
	if err != nil {
		badRequest(c, "invalid family_id")
		return
	}
	plans, err := h.C.ListByFamily(userCtx(c), familyID)
	if err != nil {
		serverError(c, err)
		return
	}
	ok(c, plans)
}

func (h *FloorPlanHandlers) GetByID(c *gin.Context) {
	planID, err := paramUUID(c, "plan_id")
	if err != nil {
		badRequest(c, "invalid plan_id")
		return
	}
	fp, err := h.C.GetByID(userCtx(c), planID)
	if err != nil {
		serverError(c, err)
		return
	}
	ok(c, fp)
}

func (h *FloorPlanHandlers) SetCover(c *gin.Context) {
	planID, err := paramUUID(c, "plan_id")
	if err != nil {
		badRequest(c, "invalid plan_id")
		return
	}
	if err := h.C.SetCover(userCtx(c), planID); err != nil {
		serverError(c, err)
		return
	}
	ok(c, gin.H{"message": "cover updated"})
}

func (h *FloorPlanHandlers) Delete(c *gin.Context) {
	planID, err := paramUUID(c, "plan_id")
	if err != nil {
		badRequest(c, "invalid plan_id")
		return
	}
	if err := h.C.Delete(userCtx(c), planID); err != nil {
		serverError(c, err)
		return
	}
	ok(c, gin.H{"message": "floor plan deleted"})
}
