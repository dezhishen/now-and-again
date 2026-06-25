package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/dezhishen/now-and-again/shared/types"
)

func (h *FamilyHandlers) Create(c *gin.Context) {
	req, err := bindJSON[types.CreateFamilyRequest](c)
	if err != nil {
		badRequest(c, err.Error()); return
	}
	family, err := h.C.Create(userCtx(c), req)
	if err != nil {
		serverError(c, err); return
	}
	created(c, family)
}
func (h *FamilyHandlers) Join(c *gin.Context) {
	req, err := bindJSON[types.JoinFamilyRequest](c)
	if err != nil {
		badRequest(c, err.Error()); return
	}
	member, err := h.C.Join(userCtx(c), req)
	if err != nil {
		serverError(c, err); return
	}
	created(c, member)
}
func (h *FamilyHandlers) Get(c *gin.Context) {
	id, err := paramUUID(c, "family_id")
	if err != nil {
		badRequest(c, "invalid family_id"); return
	}
	family, err := h.C.Get(userCtx(c), id)
	if err != nil {
		serverError(c, err); return
	}
	ok(c, family)
}
func (h *FamilyHandlers) ListMembers(c *gin.Context) {
	id, err := paramUUID(c, "family_id")
	if err != nil {
		badRequest(c, "invalid family_id"); return
	}
	members, err := h.C.ListMembers(userCtx(c), id)
	if err != nil {
		serverError(c, err); return
	}
	ok(c, members)
}
func (h *FamilyHandlers) UpdateMemberRole(c *gin.Context) {
	familyID, err := paramUUID(c, "family_id")
	if err != nil {
		badRequest(c, "invalid family_id"); return
	}
	userID, err := paramUUID(c, "user_id")
	if err != nil {
		badRequest(c, "invalid user_id"); return
	}
	var req types.UpdateMemberRoleRequest
	if err := bodyJSON(c, &req); err != nil {
		badRequest(c, err.Error()); return
	}
	if err := h.C.UpdateMemberRole(userCtx(c), familyID, userID, req.Role); err != nil {
		serverError(c, err); return
	}
	noContent(c)
}
func (h *FamilyHandlers) RemoveMember(c *gin.Context) {
	familyID, err := paramUUID(c, "family_id")
	if err != nil {
		badRequest(c, "invalid family_id"); return
	}
	userID, err := paramUUID(c, "user_id")
	if err != nil {
		badRequest(c, "invalid user_id"); return
	}
	if err := h.C.RemoveMember(userCtx(c), familyID, userID); err != nil {
		serverError(c, err); return
	}
	noContent(c)
}

func (h *FamilyHandlers) ListMyFamilies(c *gin.Context) {
	families, err := h.C.ListMyFamilies(userCtx(c))
	if err != nil {
		serverError(c, err); return
	}
	ok(c, families)
}

func (h *FamilyHandlers) LeaveFamily(c *gin.Context) {
	familyID, err := paramUUID(c, "family_id")
	if err != nil {
		badRequest(c, "invalid family_id"); return
	}
	if err := h.C.LeaveFamily(userCtx(c), familyID); err != nil {
		serverError(c, err); return
	}
	noContent(c)
}
