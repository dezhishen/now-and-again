package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/dezhishen/now-and-again/shared/contracts"
)

func RegisterRoutes(public *gin.Engine, auth *gin.RouterGroup, c *contracts.AllContracts) {
	h := NewHandlers(c)

	// ── Public ──────────────────────────────────────────────────
	public.GET("/api/system/status", h.User.CheckInit)
	public.POST("/api/setup", h.User.Setup)
	public.POST("/api/auth/register", h.User.Register)
	public.POST("/api/auth/login", h.User.Login)
	public.POST("/api/auth/refresh", h.User.Refresh)
	public.POST("/api/auth/logout", h.User.Logout)

	// ── Protected ───────────────────────────────────────────────
	// Admin
	auth.GET("/api/admin/users", h.User.ListUsers)

	// User
	auth.GET("/api/users/me", h.User.GetMe)
	auth.PUT("/api/users/me", h.User.UpdateMe)
	auth.GET("/api/users/me/families", h.Family.ListMyFamilies)

	// Family
	auth.POST("/api/families", h.Family.Create)
	auth.POST("/api/families/join", h.Family.Join)
	auth.GET("/api/families/:family_id", h.Family.Get)
	auth.GET("/api/families/:family_id/members", h.Family.ListMembers)
	auth.PUT("/api/families/:family_id/members/:user_id/role", h.Family.UpdateMemberRole)
	auth.DELETE("/api/families/:family_id/members/:user_id", h.Family.RemoveMember)
	auth.POST("/api/families/:family_id/leave", h.Family.LeaveFamily)

	// Family Join Requests
	auth.GET("/api/families/:family_id/join-requests", h.Family.ListJoinRequests)
	auth.PUT("/api/families/:family_id/join-requests", h.Family.ReviewJoinRequest)

	// Family Groups
	auth.POST("/api/families/:family_id/groups", h.Family.CreateGroup)
	auth.GET("/api/families/:family_id/groups", h.Family.ListGroups)
	auth.POST("/api/groups/:group_id/join", h.Family.JoinGroup)
	auth.POST("/api/groups/:group_id/leave", h.Family.LeaveGroup)
	auth.GET("/api/groups/:group_id/members", h.Family.ListGroupMembers)
	auth.DELETE("/api/groups/:group_id/members/:user_id", h.Family.RemoveGroupMember)
	auth.GET("/api/groups/:group_id/join-requests", h.Family.ListGroupJoinRequests)
	auth.PUT("/api/groups/:group_id/join-requests", h.Family.ReviewGroupJoinRequest)

	// API Keys
	auth.POST("/api/users/me/api-keys", h.ApiKey.Create)
	auth.GET("/api/users/me/api-keys", h.ApiKey.List)
	auth.DELETE("/api/users/me/api-keys/:key_id", h.ApiKey.Revoke)
}
