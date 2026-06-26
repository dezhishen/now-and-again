package handler

import (
	"github.com/dezhishen/now-and-again/backend/pkg/types"
	"github.com/gin-gonic/gin"
)

// ─── Floor Plan ──────────────────────────────────────────────────

func (h *FloorPlanHandlers) Upload(c *gin.Context) {
	familyID, err := paramUUID(c, "family_id")
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
	familyID, err := paramUUID(c, "family_id")
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

// ─── Rooms ───────────────────────────────────────────────────────

func (h *FloorPlanHandlers) CreateLocation(c *gin.Context) {
	planID, err := paramUUID(c, "plan_id")
	if err != nil {
		badRequest(c, "invalid plan_id")
		return
	}
	req, err := bindJSON[types.CreateLocationRequest](c)
	if err != nil {
		badRequest(c, err.Error())
		return
	}
	loc, err := h.C.CreateLocation(userCtx(c), planID, req)
	if err != nil {
		serverError(c, err)
		return
	}
	created(c, loc)
}

func (h *FloorPlanHandlers) ListLocations(c *gin.Context) {
	planID, err := paramUUID(c, "plan_id")
	if err != nil {
		badRequest(c, "invalid plan_id")
		return
	}
	locs, err := h.C.ListLocations(userCtx(c), planID)
	if err != nil {
		serverError(c, err)
		return
	}
	ok(c, locs)
}

func (h *FloorPlanHandlers) UpdateLocation(c *gin.Context) {
	locationID, err := paramUUID(c, "location_id")
	if err != nil {
		badRequest(c, "invalid location_id")
		return
	}
	req, err := bindJSON[types.UpdateLocationRequest](c)
	if err != nil {
		badRequest(c, err.Error())
		return
	}
	loc, err := h.C.UpdateLocation(userCtx(c), locationID, req)
	if err != nil {
		serverError(c, err)
		return
	}
	ok(c, loc)
}

func (h *FloorPlanHandlers) DeleteLocation(c *gin.Context) {
	locationID, err := paramUUID(c, "location_id")
	if err != nil {
		badRequest(c, "invalid location_id")
		return
	}
	if err := h.C.DeleteLocation(userCtx(c), locationID); err != nil {
		serverError(c, err)
		return
	}
	ok(c, gin.H{"message": "location deleted"})
}
