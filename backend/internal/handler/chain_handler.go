package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/dezhishen/now-and-again/shared/types"
)

func (h *ChainHandlers) Create(c *gin.Context) {
	familyID, err := paramUUID(c, "family_id")
	if err != nil {
		badRequest(c, "invalid family_id"); return
	}
	req, err := bindJSON[types.CreateChainRequest](c)
	if err != nil {
		badRequest(c, err.Error()); return
	}
	chain, err := h.C.Create(userCtx(c), familyID, req)
	if err != nil {
		serverError(c, err); return
	}
	created(c, chain)
}
func (h *ChainHandlers) List(c *gin.Context) {
	familyID, err := paramUUID(c, "family_id")
	if err != nil {
		badRequest(c, "invalid family_id"); return
	}
	list, err := h.C.List(userCtx(c), familyID)
	if err != nil {
		serverError(c, err); return
	}
	ok(c, list)
}
func (h *ChainHandlers) Get(c *gin.Context) {
	chainID, err := paramUUID(c, "chain_id")
	if err != nil {
		badRequest(c, "invalid chain_id"); return
	}
	chain, err := h.C.Get(userCtx(c), chainID)
	if err != nil {
		serverError(c, err); return
	}
	ok(c, chain)
}
func (h *ChainHandlers) AddStep(c *gin.Context) {
	chainID, err := paramUUID(c, "chain_id")
	if err != nil {
		badRequest(c, "invalid chain_id"); return
	}
	req, err := bindJSON[types.AddStepRequest](c)
	if err != nil {
		badRequest(c, err.Error()); return
	}
	step, err := h.C.AddStep(userCtx(c), chainID, req)
	if err != nil {
		serverError(c, err); return
	}
	created(c, step)
}
func (h *ChainHandlers) ReorderSteps(c *gin.Context) {
	chainID, err := paramUUID(c, "chain_id")
	if err != nil {
		badRequest(c, "invalid chain_id"); return
	}
	req, err := bindJSON[types.ReorderStepsRequest](c)
	if err != nil {
		badRequest(c, err.Error()); return
	}
	if err := h.C.ReorderSteps(userCtx(c), chainID, req); err != nil {
		serverError(c, err); return
	}
	noContent(c)
}
func (h *ChainHandlers) RemoveStep(c *gin.Context) {
	chainID, err := paramUUID(c, "chain_id")
	if err != nil {
		badRequest(c, "invalid chain_id"); return
	}
	stepID, err := paramUUID(c, "step_id")
	if err != nil {
		badRequest(c, "invalid step_id"); return
	}
	if err := h.C.RemoveStep(userCtx(c), chainID, stepID); err != nil {
		serverError(c, err); return
	}
	noContent(c)
}
func (h *ChainHandlers) Start(c *gin.Context) {
	chainID, err := paramUUID(c, "chain_id")
	if err != nil {
		badRequest(c, "invalid chain_id"); return
	}
	resp, err := h.C.Start(userCtx(c), chainID)
	if err != nil {
		serverError(c, err); return
	}
	created(c, resp)
}
