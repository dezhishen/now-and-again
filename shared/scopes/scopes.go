package scopes

import "slices"

// ─── Individual Scopes ────────────────────────────────────────────

const (
	FamilyRead     = "family:read"
	FamilyWrite    = "family:write"
	FamilyAdmin    = "family:admin"
	FloorPlanRead  = "floorplan:read"
	FloorPlanWrite = "floorplan:write"
	TaskRead       = "task:read"
	TaskWrite      = "task:write"
	IcsRead        = "ics:read"
	UserRead       = "user:read"
	AdminRead      = "admin:read"
	AdminWrite     = "admin:write"
)

// ─── Scope Groups ─────────────────────────────────────────────────

var ReadScopes = []string{FamilyRead, FloorPlanRead, TaskRead, IcsRead, UserRead}
var WriteScopes = []string{FamilyRead, FamilyWrite, FloorPlanRead, FloorPlanWrite, TaskRead, TaskWrite, IcsRead, UserRead}
var AdminScopes = []string{FamilyRead, FamilyWrite, FamilyAdmin, FloorPlanRead, FloorPlanWrite, TaskRead, TaskWrite, IcsRead, UserRead, AdminRead, AdminWrite}

// All returns every defined scope.
func All() []string {
	return []string{
		FamilyRead, FamilyWrite, FamilyAdmin,
		FloorPlanRead, FloorPlanWrite,
		TaskRead, TaskWrite,
		IcsRead,
		UserRead,
		AdminRead, AdminWrite,
	}
}

// Valid checks if a scope string is known.
func Valid(scope string) bool {
	return slices.Contains(All(), scope)
}

// Has checks if the granted scopes contain the required scope.
// A scope group prefix match is not used — exact match only.
func Has(granted []string, required string) bool {
	return slices.Contains(granted, required)
}

// ExpandGroups expands group keywords into individual scopes.
// "read" → ReadScopes, "write" → WriteScopes, "admin" → AdminScopes.
func ExpandGroups(input []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range input {
		var list []string
		switch s {
		case "read":
			list = ReadScopes
		case "write":
			list = WriteScopes
		case "admin":
			list = AdminScopes
		default:
			if Valid(s) {
				list = []string{s}
			}
		}
		for _, v := range list {
			if !seen[v] {
				seen[v] = true
				out = append(out, v)
			}
		}
	}
	return out
}

// RouteScope maps API routes to their required scope.
// Key = HTTP method + path (e.g. "POST /api/families/:family_id/tasks")
func RouteScope(method, path string) string {
	// Strip path params for matching
	key := method + " " + path
	if s, ok := routeScopes[key]; ok {
		return s
	}
	return ""
}

var routeScopes = map[string]string{

	// ── Family ─────────────────────────────────────────────────
	"POST /api/families":                 FamilyWrite,
	"POST /api/families/join":            FamilyWrite,
	"GET  /api/families/:family_id":      FamilyRead,
	"PATCH /api/families/:family_id":     FamilyWrite,
	"DELETE /api/families/:family_id":    FamilyAdmin,
	"GET  /api/families/:family_id/members":    FamilyRead,
	"PUT  /api/families/:family_id/members/:user_id/role": FamilyAdmin,
	"DELETE /api/families/:family_id/members/:user_id":    FamilyAdmin,
	"POST /api/families/:family_id/leave": FamilyWrite,

	// Join requests
	"GET /api/families/:family_id/join-requests":  FamilyRead,
	"PUT /api/families/:family_id/join-requests":  FamilyAdmin,

	// Groups
	"POST /api/families/:family_id/groups":        FamilyWrite,
	"GET  /api/families/:family_id/groups":        FamilyRead,
	"POST /api/groups/:group_id/join":             FamilyWrite,
	"POST /api/groups/:group_id/leave":            FamilyWrite,
	"GET  /api/groups/:group_id/members":          FamilyRead,
	"DELETE /api/groups/:group_id/members/:user_id": FamilyAdmin,
	"GET  /api/groups/:group_id/join-requests":    FamilyRead,
	"PUT  /api/groups/:group_id/join-requests":    FamilyAdmin,

	// ── Floor Plans ────────────────────────────────────────────
	"POST /api/families/:family_id/floor-plans":  FloorPlanWrite,
	"GET  /api/families/:family_id/floor-plans":  FloorPlanRead,
	"GET  /api/floor-plans/:plan_id":             FloorPlanRead,
	"PUT  /api/floor-plans/:plan_id/cover":       FloorPlanWrite,
	"DELETE /api/floor-plans/:plan_id":           FloorPlanWrite,
	"POST /api/floor-plans/:plan_id/locations":   FloorPlanWrite,
	"GET  /api/floor-plans/:plan_id/locations":   FloorPlanRead,
	"PUT  /api/locations/:location_id":           FloorPlanWrite,
	"DELETE /api/locations/:location_id":         FloorPlanWrite,

	// ── Tasks ──────────────────────────────────────────────────
	"POST /api/families/:family_id/tasks":   TaskWrite,
	"GET  /api/families/:family_id/tasks":   TaskRead,
	"PUT  /api/tasks/:task_id":              TaskWrite,
	"DELETE /api/tasks/:task_id":            TaskWrite,
	"GET  /api/tasks/:task_id/logs":         TaskRead,
	"GET  /api/families/:family_id/todos":   TaskRead,
	"PUT  /api/todos/:todo_id":              TaskWrite,

	// ── ICS ────────────────────────────────────────────────────
	"POST /api/families/:family_id/ics-feeds":  TaskWrite,
	"GET  /api/families/:family_id/ics-feeds":  IcsRead,
	"GET  /api/ics-feeds/:feed_id":             IcsRead,
	"PUT  /api/ics-feeds/:feed_id":             TaskWrite,
	"DELETE /api/ics-feeds/:feed_id":           TaskWrite,

	// ── User ───────────────────────────────────────────────────
	"GET /api/users/me":          UserRead,
	"PUT /api/users/me":          UserRead,
	"GET /api/users/me/families": FamilyRead,
	"POST /api/users/me/api-keys": UserRead,
	"GET  /api/users/me/api-keys": UserRead,
	"DELETE /api/users/me/api-keys/:key_id": UserRead,

	// ── Admin ──────────────────────────────────────────────────
	"GET /api/admin/users":    AdminRead,
	"GET /api/admin/settings": AdminRead,
	"PUT /api/admin/settings": AdminWrite,
}

// Descriptions returns human-readable labels for UI.
func Descriptions() map[string]string {
	return map[string]string{
		FamilyRead:     "查看家庭信息",
		FamilyWrite:    "修改家庭资源",
		FamilyAdmin:    "管理家庭成员与角色",
		FloorPlanRead:  "查看户型图",
		FloorPlanWrite: "上传/编辑户型图",
		TaskRead:       "查看任务与待办",
		TaskWrite:      "创建/编辑任务",
		IcsRead:        "读取日历订阅",
		UserRead:       "查看个人信息",
		AdminRead:      "查看管理面板",
		AdminWrite:     "修改系统设置",
	}
}
