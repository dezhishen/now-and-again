package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/dezhishen/now-and-again/shared/types"
)

func (h *SubGroupHandlers) Create(c *gin.Context) {
	familyID, err := paramUUID(c, "family_id")
	if err != nil {
		badRequest(c, "invalid family_id")
		return
	}
	req, err := bindJSON[types.CreateSubGroupRequest](c)
	if err != nil {
		badRequest(c, err.Error())
		return
	}
	sg, err := h.C.Create(userCtx(c), familyID, req)
	if err != nil {
		serverError(c, err)
		return
	}
	created(c, sg)
}
func (h *SubGroupHandlers) List(c *gin.Context) {
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
func (h *SubGroupHandlers) AddMember(c *gin.Context) {
	subGroupID, err := paramUUID(c, "subgroup_id")
	if err != nil {
		badRequest(c, "invalid subgroup_id")
		return
	}
	req, err := bindJSON[types.AddSubGroupMemberRequest](c)
	if err != nil {
		badRequest(c, err.Error())
		return
	}
	member, err := h.C.AddMember(userCtx(c), subGroupID, req)
	if err != nil {
		serverError(c, err)
		return
	}
	created(c, member)
}
func (h *SubGroupHandlers) RemoveMember(c *gin.Context) {
	subGroupID, err := paramUUID(c, "subgroup_id")
	if err != nil {
		badRequest(c, "invalid subgroup_id")
		return
	}
	userID, err := paramUUID(c, "user_id")
	if err != nil {
		badRequest(c, "invalid user_id")
		return
	}
	if err := h.C.RemoveMember(userCtx(c), subGroupID, userID); err != nil {
		serverError(c, err)
		return
	}
	noContent(c)
}
