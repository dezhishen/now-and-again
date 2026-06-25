package handler

import (
	"github.com/dezhishen/now-and-again/shared/types"
	"github.com/gin-gonic/gin"
)

func (h *ApiKeyHandlers) Create(c *gin.Context) {
	req, err := bindJSON[types.CreateApiKeyRequest](c)
	if err != nil {
		badRequest(c, err.Error())
		return
	}
	resp, err := h.C.Create(userCtx(c), req)
	if err != nil {
		serverError(c, err)
		return
	}
	created(c, resp)
}

func (h *ApiKeyHandlers) List(c *gin.Context) {
	keys, err := h.C.List(userCtx(c))
	if err != nil {
		serverError(c, err)
		return
	}
	ok(c, keys)
}

func (h *ApiKeyHandlers) Revoke(c *gin.Context) {
	keyID := c.Param("key_id")
	if err := h.C.Revoke(userCtx(c), keyID); err != nil {
		serverError(c, err)
		return
	}
	noContent(c)
}
