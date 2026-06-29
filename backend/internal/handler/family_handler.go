package handler

import (
	"github.com/dezhishen/now-and-again/backend/pkg/types"
	"github.com/gin-gonic/gin"
)

// ─── Family ───────────────────────────────────────────────────────

func (h *FamilyHandlers) Create(c *gin.Context) {
	req, err := bindJSON[types.CreateFamilyRequest](c)
	if err != nil {
		badRequest(c, err.Error())
		return
	}
	f, err := h.C.Create(userCtx(c), req)
	if err != nil {
		serverError(c, err)
		return
	}
	created(c, f)
}

func (h *FamilyHandlers) Join(c *gin.Context) {
	req, err := bindJSON[types.JoinFamilyRequest](c)
	if err != nil {
		badRequest(c, err.Error())
		return
	}
	m, err := h.C.Join(userCtx(c), req)
	if err != nil {
		serverError(c, err)
		return
	}
	created(c, m)
}

func (h *FamilyHandlers) Get(c *gin.Context) {
	familyID, err := paramUUID(c, "family_id")
	if err != nil {
		badRequest(c, "invalid family_id")
		return
	}
	f, err := h.C.Get(userCtx(c), familyID)
	if err != nil {
		serverError(c, err)
		return
	}
	ok(c, f)
}

func (h *FamilyHandlers) Update(c *gin.Context) {
	familyID, err := paramUUID(c, "family_id")
	if err != nil {
		badRequest(c, "invalid family_id")
		return
	}
	req, err := bindJSON[types.UpdateFamilyRequest](c)
	if err != nil {
		badRequest(c, err.Error())
		return
	}
	f, err := h.C.Update(userCtx(c), familyID, req)
	if err != nil {
		serverError(c, err)
		return
	}
	ok(c, f)
}

func (h *FamilyHandlers) Delete(c *gin.Context) {
	familyID, err := paramUUID(c, "family_id")
	if err != nil {
		badRequest(c, "invalid family_id")
		return
	}
	if err := h.C.Delete(userCtx(c), familyID); err != nil {
		serverError(c, err)
		return
	}
	ok(c, gin.H{"message": "family archived"})
}

func (h *FamilyHandlers) Restore(c *gin.Context) {
	familyID, err := paramUUID(c, "family_id")
	if err != nil {
		badRequest(c, "invalid family_id")
		return
	}
	if err := h.C.Restore(userCtx(c), familyID); err != nil {
		serverError(c, err)
		return
	}
	ok(c, gin.H{"message": "family restored"})
}

func (h *FamilyHandlers) ListMyFamilies(c *gin.Context) {
	families, err := h.C.ListMyFamilies(userCtx(c))
	if err != nil {
		serverError(c, err)
		return
	}
	ok(c, families)
}

func (h *FamilyHandlers) ListMembers(c *gin.Context) {
	familyID, err := paramUUID(c, "family_id")
	if err != nil {
		badRequest(c, "invalid family_id")
		return
	}
	members, err := h.C.ListMembers(userCtx(c), familyID)
	if err != nil {
		serverError(c, err)
		return
	}
	ok(c, members)
}

func (h *FamilyHandlers) UpdateMemberRole(c *gin.Context) {
	familyID, err := paramUUID(c, "family_id")
	if err != nil {
		badRequest(c, "invalid family_id")
		return
	}
	userID, err := paramUUID(c, "user_id")
	if err != nil {
		badRequest(c, "invalid user_id")
		return
	}
	req, err := bindJSON[types.UpdateMemberRoleRequest](c)
	if err != nil {
		badRequest(c, err.Error())
		return
	}
	if err := h.C.UpdateMemberRole(userCtx(c), familyID, userID, req.Role); err != nil {
		serverError(c, err)
		return
	}
	ok(c, gin.H{"message": "role updated"})
}

func (h *FamilyHandlers) RemoveMember(c *gin.Context) {
	familyID, err := paramUUID(c, "family_id")
	if err != nil {
		badRequest(c, "invalid family_id")
		return
	}
	userID, err := paramUUID(c, "user_id")
	if err != nil {
		badRequest(c, "invalid user_id")
		return
	}
	if err := h.C.RemoveMember(userCtx(c), familyID, userID); err != nil {
		serverError(c, err)
		return
	}
	ok(c, gin.H{"message": "member removed"})
}

func (h *FamilyHandlers) LeaveFamily(c *gin.Context) {
	familyID, err := paramUUID(c, "family_id")
	if err != nil {
		badRequest(c, "invalid family_id")
		return
	}
	if err := h.C.LeaveFamily(userCtx(c), familyID); err != nil {
		serverError(c, err)
		return
	}
	ok(c, gin.H{"message": "left family"})
}

// ─── Join Requests ────────────────────────────────────────────────

func (h *FamilyHandlers) ListJoinRequests(c *gin.Context) {
	familyID, err := paramUUID(c, "family_id")
	if err != nil {
		badRequest(c, "invalid family_id")
		return
	}
	reqs, err := h.C.ListJoinRequests(userCtx(c), familyID)
	if err != nil {
		serverError(c, err)
		return
	}
	ok(c, reqs)
}

func (h *FamilyHandlers) ReviewJoinRequest(c *gin.Context) {
	familyID, err := paramUUID(c, "family_id")
	if err != nil {
		badRequest(c, "invalid family_id")
		return
	}
	req, err := bindJSON[types.ReviewJoinRequest](c)
	if err != nil {
		badRequest(c, err.Error())
		return
	}
	if err := h.C.ReviewJoinRequest(userCtx(c), familyID, req); err != nil {
		serverError(c, err)
		return
	}
	ok(c, gin.H{"message": "request reviewed"})
}

// ─── Family Group ─────────────────────────────────────────────────

func (h *FamilyHandlers) CreateGroup(c *gin.Context) {
	familyID, err := paramUUID(c, "family_id")
	if err != nil {
		badRequest(c, "invalid family_id")
		return
	}
	req, err := bindJSON[types.CreateFamilyGroupRequest](c)
	if err != nil {
		badRequest(c, err.Error())
		return
	}
	g, err := h.C.CreateGroup(userCtx(c), familyID, req)
	if err != nil {
		serverError(c, err)
		return
	}
	created(c, g)
}

func (h *FamilyHandlers) ListGroups(c *gin.Context) {
	familyID, err := paramUUID(c, "family_id")
	if err != nil {
		badRequest(c, "invalid family_id")
		return
	}
	groups, err := h.C.ListGroups(userCtx(c), familyID)
	if err != nil {
		serverError(c, err)
		return
	}
	ok(c, groups)
}

func (h *FamilyHandlers) JoinGroup(c *gin.Context) {
	groupID, err := paramUUID(c, "group_id")
	if err != nil {
		badRequest(c, "invalid group_id")
		return
	}
	m, err := h.C.JoinGroup(userCtx(c), groupID)
	if err != nil {
		serverError(c, err)
		return
	}
	created(c, m)
}

func (h *FamilyHandlers) LeaveGroup(c *gin.Context) {
	groupID, err := paramUUID(c, "group_id")
	if err != nil {
		badRequest(c, "invalid group_id")
		return
	}
	if err := h.C.LeaveGroup(userCtx(c), groupID); err != nil {
		serverError(c, err)
		return
	}
	ok(c, gin.H{"message": "left group"})
}

func (h *FamilyHandlers) ListGroupMembers(c *gin.Context) {
	groupID, err := paramUUID(c, "group_id")
	if err != nil {
		badRequest(c, "invalid group_id")
		return
	}
	members, err := h.C.ListGroupMembers(userCtx(c), groupID)
	if err != nil {
		serverError(c, err)
		return
	}
	ok(c, members)
}

func (h *FamilyHandlers) RemoveGroupMember(c *gin.Context) {
	groupID, err := paramUUID(c, "group_id")
	if err != nil {
		badRequest(c, "invalid group_id")
		return
	}
	userID, err := paramUUID(c, "user_id")
	if err != nil {
		badRequest(c, "invalid user_id")
		return
	}
	if err := h.C.RemoveGroupMember(userCtx(c), groupID, userID); err != nil {
		serverError(c, err)
		return
	}
	ok(c, gin.H{"message": "member removed"})
}

func (h *FamilyHandlers) ListGroupJoinRequests(c *gin.Context) {
	groupID, err := paramUUID(c, "group_id")
	if err != nil {
		badRequest(c, "invalid group_id")
		return
	}
	reqs, err := h.C.ListGroupJoinRequests(userCtx(c), groupID)
	if err != nil {
		serverError(c, err)
		return
	}
	ok(c, reqs)
}

func (h *FamilyHandlers) ReviewGroupJoinRequest(c *gin.Context) {
	groupID, err := paramUUID(c, "group_id")
	if err != nil {
		badRequest(c, "invalid group_id")
		return
	}
	req, err := bindJSON[types.ReviewGroupJoinRequest](c)
	if err != nil {
		badRequest(c, err.Error())
		return
	}
	if err := h.C.ReviewGroupJoinRequest(userCtx(c), groupID, req); err != nil {
		serverError(c, err)
		return
	}
	ok(c, gin.H{"message": "request reviewed"})
}
