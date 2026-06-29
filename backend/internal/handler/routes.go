package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/dezhishen/now-and-again/backend/pkg/contracts"
)

func RegisterRoutes(public *gin.Engine, auth *gin.RouterGroup, familyAuth *gin.RouterGroup, c *contracts.AllContracts, imgHandler *ImageHandlers, settingsHandler *SettingsHandlers, taskHandler *TaskHandlers, todoHandler *TodoHandlers, logHandler *LogHandlers, icsHandler *IcsHandlers, calendarHandler *CalendarHandlers, locationHandler *LocationHandlers) {
	h := NewHandlers(c)

	// ── Public ──────────────────────────────────────────────────
	public.POST("/api/auth/register", h.User.Register)
	public.POST("/api/auth/login", h.User.Login)
	public.POST("/api/auth/refresh", h.User.Refresh)
	public.POST("/api/auth/logout", h.User.Logout)

	// Image serving (public)
	public.GET("/api/images/:id", imgHandler.Serve)

	// ── Protected (no family required) ──────────────────────────
	auth.GET("/api/admin/users", h.User.ListUsers)
	auth.GET("/api/admin/settings", settingsHandler.GetAll)
	auth.PUT("/api/admin/settings", settingsHandler.Update)

	// User
	auth.GET("/api/users/me", h.User.GetMe)
	auth.PUT("/api/users/me", h.User.UpdateMe)
	auth.GET("/api/users/me/families", h.Family.ListMyFamilies)

	// Family CRUD (no active family needed)
	auth.POST("/api/families", h.Family.Create)
	auth.POST("/api/families/join", h.Family.Join)
	auth.GET("/api/families/:family_id", h.Family.Get)
	auth.PATCH("/api/families/:family_id", h.Family.Update)
	auth.DELETE("/api/families/:family_id", h.Family.Delete)
	auth.POST("/api/families/:family_id/restore", h.Family.Restore)
	auth.POST("/api/families/:family_id/leave", h.Family.LeaveFamily)

	// API Keys
	auth.POST("/api/users/me/api-keys", h.ApiKey.Create)
	auth.GET("/api/users/me/api-keys", h.ApiKey.List)
	auth.DELETE("/api/users/me/api-keys/:key_id", h.ApiKey.Revoke)

	// ── Family-scoped (requires X-Family-Id header) ─────────────

	// Members
	familyAuth.GET("/api/members", h.Family.ListMembers)
	familyAuth.PUT("/api/members/:user_id/role", h.Family.UpdateMemberRole)
	familyAuth.DELETE("/api/members/:user_id", h.Family.RemoveMember)
	familyAuth.POST("/api/family/leave", h.Family.LeaveFamily)
	familyAuth.GET("/api/family/join-requests", h.Family.ListJoinRequests)
	familyAuth.PUT("/api/family/join-requests", h.Family.ReviewJoinRequest)

	// Family Groups
	familyAuth.POST("/api/groups", h.Family.CreateGroup)
	familyAuth.GET("/api/groups", h.Family.ListGroups)
	familyAuth.POST("/api/groups/:group_id/join", h.Family.JoinGroup)
	familyAuth.POST("/api/groups/:group_id/leave", h.Family.LeaveGroup)
	familyAuth.GET("/api/groups/:group_id/members", h.Family.ListGroupMembers)
	familyAuth.DELETE("/api/groups/:group_id/members/:user_id", h.Family.RemoveGroupMember)
	familyAuth.GET("/api/groups/:group_id/join-requests", h.Family.ListGroupJoinRequests)
	familyAuth.PUT("/api/groups/:group_id/join-requests", h.Family.ReviewGroupJoinRequest)

	// Floor Plans
	familyAuth.POST("/api/floor-plans", h.FloorPlan.Upload)
	familyAuth.GET("/api/floor-plans", h.FloorPlan.ListByFamily)
	familyAuth.GET("/api/floor-plans/:plan_id", h.FloorPlan.GetByID)
	familyAuth.PUT("/api/floor-plans/:plan_id/cover", h.FloorPlan.SetCover)
	familyAuth.DELETE("/api/floor-plans/:plan_id", h.FloorPlan.Delete)

	// Locations
	familyAuth.GET("/api/locations", locationHandler.ListFamilyLocations)
	familyAuth.POST("/api/locations", locationHandler.CreateLocation)
	familyAuth.GET("/api/floor-plans/:plan_id/locations", locationHandler.ListFloorPlanLocations)
	familyAuth.PUT("/api/locations/:location_id", locationHandler.UpdateLocation)
	familyAuth.DELETE("/api/locations/:location_id", locationHandler.DeleteLocation)

	// Tasks
	familyAuth.POST("/api/tasks", taskHandler.Create)
	familyAuth.GET("/api/tasks", taskHandler.List)
	familyAuth.GET("/api/tasks/:task_id", taskHandler.Get)
	familyAuth.PUT("/api/tasks/:task_id", taskHandler.Update)
	familyAuth.PUT("/api/tasks/:task_id/enabled", taskHandler.SetEnabled)
	familyAuth.DELETE("/api/tasks/:task_id", taskHandler.Delete)
	familyAuth.POST("/api/tasks/:task_id/trigger", taskHandler.Trigger)

	// Task Logs
	familyAuth.GET("/api/tasks/:task_id/logs", logHandler.ListLogs)

	// Todos
	familyAuth.GET("/api/todos", todoHandler.ListTodos)
	familyAuth.GET("/api/todos/:todo_id", todoHandler.GetTodo)
	familyAuth.PUT("/api/todos/:todo_id", todoHandler.CompleteTodo)

	// Calendar
	familyAuth.GET("/api/calendar", calendarHandler.GetCalendar)

	// ICS Feeds
	familyAuth.POST("/api/ics-feeds", icsHandler.Create)
	familyAuth.GET("/api/ics-feeds", icsHandler.List)
	familyAuth.GET("/api/ics-feeds/:feed_id", icsHandler.Get)
	familyAuth.PUT("/api/ics-feeds/:feed_id", icsHandler.Update)
	familyAuth.DELETE("/api/ics-feeds/:feed_id", icsHandler.Delete)

	// ICS public endpoint (no JWT - custom auth)
	public.GET("/api/ics/:token", icsHandler.ServeICS)
}
