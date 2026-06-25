package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/dezhishen/now-and-again/shared/contracts"
)

// RegisterRoutes wires all API endpoints using contract interfaces.
func RegisterRoutes(public *gin.Engine, auth *gin.RouterGroup, c *contracts.AllContracts) {
	h := NewHandlers(c)

	// ── Public ──────────────────────────────────────────────────
	public.GET("/api/system/status", h.User.CheckInit)
	public.GET("/api/schedule-types", h.Task.ListScheduleTypes)
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

	// SubGroup
	auth.POST("/api/families/:family_id/subgroups", h.SubGroup.Create)
	auth.GET("/api/families/:family_id/subgroups", h.SubGroup.List)
	auth.POST("/api/subgroups/:subgroup_id/members", h.SubGroup.AddMember)
	auth.DELETE("/api/subgroups/:subgroup_id/members/:user_id", h.SubGroup.RemoveMember)

	// Task
	auth.POST("/api/families/:family_id/tasks", h.Task.Create)
	auth.GET("/api/families/:family_id/tasks", h.Task.List)
	auth.GET("/api/tasks/:task_id", h.Task.Get)
	auth.PATCH("/api/tasks/:task_id", h.Task.Update)
	auth.POST("/api/tasks/:task_id/assignees", h.Task.SetAssignees)
	auth.POST("/api/tasks/:task_id/dependencies", h.Task.AddDependency)
	auth.DELETE("/api/tasks/:task_id/dependencies/:dep_id", h.Task.RemoveDependency)

	// Task Chain
	auth.POST("/api/families/:family_id/chains", h.Chain.Create)
	auth.GET("/api/families/:family_id/chains", h.Chain.List)
	auth.GET("/api/chains/:chain_id", h.Chain.Get)
	auth.POST("/api/chains/:chain_id/steps", h.Chain.AddStep)
	auth.PUT("/api/chains/:chain_id/steps/reorder", h.Chain.ReorderSteps)
	auth.DELETE("/api/chains/:chain_id/steps/:step_id", h.Chain.RemoveStep)
	auth.POST("/api/chains/:chain_id/start", h.Chain.Start)

	// Inspection
	auth.POST("/api/families/:family_id/inspections", h.Inspection.Create)
	auth.GET("/api/families/:family_id/inspections", h.Inspection.List)
	auth.GET("/api/inspections/:inspection_id", h.Inspection.Get)
	auth.POST("/api/inspections/:inspection_id/items", h.Inspection.AddItem)
	auth.PATCH("/api/inspections/:inspection_id/items/:item_id", h.Inspection.UpdateItem)
	auth.POST("/api/inspections/:inspection_id/complete", h.Inspection.Complete)

	// Task Log
	auth.GET("/api/tasks/:task_id/logs", h.Log.ListByTask)
	auth.POST("/api/tasks/:task_id/comments", h.Log.AddComment)

	// Notification
	auth.GET("/api/users/me/notifications", h.Notification.List)
	auth.PUT("/api/users/me/channel-configs", h.Notification.UpsertChannelConfig)
	auth.GET("/api/families/:family_id/notification-templates", h.Notification.ListTemplates)
	auth.PUT("/api/families/:family_id/notification-templates", h.Notification.UpsertTemplate)

	// API Keys
	auth.POST("/api/users/me/api-keys", h.ApiKey.Create)
	auth.GET("/api/users/me/api-keys", h.ApiKey.List)
	auth.DELETE("/api/users/me/api-keys/:key_id", h.ApiKey.Revoke)
}
