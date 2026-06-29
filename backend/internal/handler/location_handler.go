package handler

import (
	"github.com/dezhishen/now-and-again/backend/pkg/contracts"
	"github.com/dezhishen/now-and-again/backend/pkg/types"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type LocationHandlers struct {
	C contracts.LocationContract
}

func (h *LocationHandlers) CreateLocation(c *gin.Context) {
	fid := familyID(c)
	familyID, err := uuid.Parse(fid)
	if err != nil {
		badRequest(c, "invalid family_id")
		return
	}
	req, err := bindJSON[types.CreateLocationRequest](c)
	if err != nil {
		validationError(c, err)
		return
	}
	loc, err := h.C.CreateLocation(userCtx(c), familyID, req)
	if err != nil {
		serverError(c, err)
		return
	}
	created(c, loc)
}

func (h *LocationHandlers) ListFamilyLocations(c *gin.Context) {
	fid := familyID(c)
	familyID, err := uuid.Parse(fid)
	if err != nil {
		badRequest(c, "invalid family_id")
		return
	}
	locs, err := h.C.ListFamilyLocations(userCtx(c), familyID)
	if err != nil {
		serverError(c, err)
		return
	}
	ok(c, locs)
}

func (h *LocationHandlers) ListFloorPlanLocations(c *gin.Context) {
	planID, err := paramUUID(c, "plan_id")
	if err != nil {
		badRequest(c, "invalid plan_id")
		return
	}
	locs, err := h.C.ListFloorPlanLocations(userCtx(c), planID)
	if err != nil {
		serverError(c, err)
		return
	}
	ok(c, locs)
}

func (h *LocationHandlers) UpdateLocation(c *gin.Context) {
	locationID, err := paramUUID(c, "location_id")
	if err != nil {
		badRequest(c, "invalid location_id")
		return
	}
	req, err := bindJSON[types.UpdateLocationRequest](c)
	if err != nil {
		validationError(c, err)
		return
	}
	loc, err := h.C.UpdateLocation(userCtx(c), locationID, req)
	if err != nil {
		serverError(c, err)
		return
	}
	ok(c, loc)
}

func (h *LocationHandlers) DeleteLocation(c *gin.Context) {
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
