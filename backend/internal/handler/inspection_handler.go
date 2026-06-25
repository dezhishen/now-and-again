package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/dezhishen/now-and-again/shared/types"
)

func (h *InspectionHandlers) Create(c *gin.Context) {
	familyID, err := paramUUID(c, "family_id")
	if err != nil {
		badRequest(c, "invalid family_id")
		return
	}
	req, err := bindJSON[types.CreateInspectionRequest](c)
	if err != nil {
		badRequest(c, err.Error())
		return
	}
	insp, err := h.C.Create(userCtx(c), familyID, req)
	if err != nil {
		serverError(c, err)
		return
	}
	created(c, insp)
}
func (h *InspectionHandlers) List(c *gin.Context) {
	familyID, err := paramUUID(c, "family_id")
	if err != nil {
		badRequest(c, "invalid family_id")
		return
	}
	list, err := h.C.List(userCtx(c), familyID)
	if err != nil {
		serverError(c, err)
		return
	}
	ok(c, list)
}
func (h *InspectionHandlers) Get(c *gin.Context) {
	inspectionID, err := paramUUID(c, "inspection_id")
	if err != nil {
		badRequest(c, "invalid inspection_id")
		return
	}
	insp, err := h.C.Get(userCtx(c), inspectionID)
	if err != nil {
		serverError(c, err)
		return
	}
	ok(c, insp)
}
func (h *InspectionHandlers) AddItem(c *gin.Context) {
	inspectionID, err := paramUUID(c, "inspection_id")
	if err != nil {
		badRequest(c, "invalid inspection_id")
		return
	}
	req, err := bindJSON[types.AddInspectionItemRequest](c)
	if err != nil {
		badRequest(c, err.Error())
		return
	}
	item, err := h.C.AddItem(userCtx(c), inspectionID, req)
	if err != nil {
		serverError(c, err)
		return
	}
	created(c, item)
}
func (h *InspectionHandlers) UpdateItem(c *gin.Context) {
	inspectionID, err := paramUUID(c, "inspection_id")
	if err != nil {
		badRequest(c, "invalid inspection_id")
		return
	}
	itemID, err := paramUUID(c, "item_id")
	if err != nil {
		badRequest(c, "invalid item_id")
		return
	}
	req, err := bindJSON[types.UpdateInspectionItemRequest](c)
	if err != nil {
		badRequest(c, err.Error())
		return
	}
	item, err := h.C.UpdateItem(userCtx(c), inspectionID, itemID, req)
	if err != nil {
		serverError(c, err)
		return
	}
	ok(c, item)
}
func (h *InspectionHandlers) Complete(c *gin.Context) {
	inspectionID, err := paramUUID(c, "inspection_id")
	if err != nil {
		badRequest(c, "invalid inspection_id")
		return
	}
	insp, err := h.C.Complete(userCtx(c), inspectionID)
	if err != nil {
		serverError(c, err)
		return
	}
	ok(c, insp)
}
