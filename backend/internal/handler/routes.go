package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/dezhishen/now-and-again/backend/pkg/contracts"
)

func RegisterRoutes(public *gin.Engine, auth *gin.RouterGroup, c *contracts.AllContracts, imgHandler *ImageHandlers, settingsHandler *SettingsHandlers, taskHandler *TaskHandlers, icsHandler *IcsHandlers) {
	h := NewHandlers(c)

	// ── Public ──────────────────────────────────────────────────
	public.POST("/api/auth/register", h.User.Register)
	public.POST("/api/auth/login", h.User.Login)
	public.POST("/api/auth/refresh", h.User.Refresh)
	public.POST("/api/auth/logout", h.User.Logout)

	// Image serving (public)
	public.GET("/api/images/:id", imgHandler.Serve)

	// ── Protected ───────────────────────────────────────────────
	// Admin
	auth.GET("/api/admin/users", h.User.ListUsers)
	auth.GET("/api/admin/settings", settingsHandler.GetAll)
	auth.PUT("/api/admin/settings", settingsHandler.Update)

	// User
	auth.GET("/api/users/me", h.User.GetMe)
	auth.PUT("/api/users/me", h.User.UpdateMe)
	auth.GET("/api/users/me/families", h.Family.ListMyFamilies)

	// Family
	auth.POST("/api/families", h.Family.Create)
	auth.POST("/api/families/join", h.Family.Join)
	auth.GET("/api/families/:family_id", h.Family.Get)
	auth.PATCH("/api/families/:family_id", h.Family.Update)
	auth.DELETE("/api/families/:family_id", h.Family.Delete)
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

	// Floor Plans
	auth.POST("/api/families/:family_id/floor-plans", h.FloorPlan.Upload)
	auth.GET("/api/families/:family_id/floor-plans", h.FloorPlan.ListByFamily)
	auth.GET("/api/floor-plans/:plan_id", h.FloorPlan.GetByID)
	auth.PUT("/api/floor-plans/:plan_id/cover", h.FloorPlan.SetCover)
	auth.DELETE("/api/floor-plans/:plan_id", h.FloorPlan.Delete)

	// Locations (first-class entity)
	auth.GET("/api/families/:family_id/locations", h.FloorPlan.ListFamilyLocations)
	auth.POST("/api/families/:family_id/locations", h.FloorPlan.CreateLocation)
	auth.GET("/api/floor-plans/:plan_id/locations", h.FloorPlan.ListFloorPlanLocations)
	auth.PUT("/api/locations/:location_id", h.FloorPlan.UpdateLocation)
	auth.DELETE("/api/locations/:location_id", h.FloorPlan.DeleteLocation)

	// Tasks
	auth.POST("/api/families/:family_id/tasks", taskHandler.Create)
	auth.GET("/api/families/:family_id/tasks", taskHandler.List)
	auth.GET("/api/tasks/:task_id", taskHandler.Get)
	auth.PUT("/api/tasks/:task_id", taskHandler.Update)
	auth.DELETE("/api/tasks/:task_id", taskHandler.Delete)
	auth.GET("/api/tasks/:task_id/logs", taskHandler.ListLogs)
	auth.POST("/api/tasks/:task_id/trigger", taskHandler.Trigger)

	// Todos
	auth.GET("/api/families/:family_id/todos", taskHandler.ListTodos)
	auth.GET("/api/todos/:todo_id", taskHandler.GetTodo)
	auth.PUT("/api/todos/:todo_id", taskHandler.CompleteTodo)

	// Calendar
	auth.GET("/api/families/:family_id/calendar", taskHandler.GetCalendar)

	// ICS Feeds (authenticated management)
	auth.POST("/api/families/:family_id/ics-feeds", icsHandler.Create)
	auth.GET("/api/families/:family_id/ics-feeds", icsHandler.List)
	auth.GET("/api/ics-feeds/:feed_id", icsHandler.Get)
	auth.PUT("/api/ics-feeds/:feed_id", icsHandler.Update)
	auth.DELETE("/api/ics-feeds/:feed_id", icsHandler.Delete)

	// ICS public endpoint (no JWT - custom auth)
	public.GET("/api/ics/:token", icsHandler.ServeICS)
}
