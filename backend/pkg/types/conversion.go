package types

import (
	"encoding/json"

	"github.com/dezhishen/now-and-again/backend/pkg/model"
)

// ─── Model → DTO conversions ─────────────────────────────────────
// These live in types because "types" defines the usage/API layer,
// while "model" only describes DB structure.

// TaskFromModel converts a model.TaskModel to the API DTO.
func TaskFromModel(m *model.TaskModel) *Task {
	var data any
	json.Unmarshal([]byte(m.ScheduleData), &data)
	return &Task{
		ID: m.ID, FamilyID: m.FamilyID, GroupID: m.GroupID,
		ParentTaskID: m.ParentTaskID, IsRoot: m.IsRoot,
		LocationID: m.LocationID,
		Name:       m.Name, ScheduleType: m.ScheduleType, ScheduleData: data,
		Enabled: m.Enabled, Kind: m.Kind, DisplaySummary: m.DisplaySummary,
		Archived:   m.Archived,
		LastTodoAt: m.LastTodoAt,
		CreatedBy:  m.CreatedBy, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt,
	}
}

// TodoFromModel converts a model.TodoModel to the API DTO.
func TodoFromModel(m *model.TodoModel) *Todo {
	var task *Task
	if m.Task.ID != "" {
		task = TaskFromModel(&m.Task)
	}
	var user *User
	if m.User.ID != "" {
		user = UserFromModel(&m.User)
	}
	return &Todo{
		ID: m.ID, TaskID: m.TaskID, FamilyID: m.FamilyID,
		LocationID: m.LocationID,
		AssignedTo: m.AssignedTo, Status: m.Status, Remark: m.Remark,
		DueStart:    m.DueStart,
		DueDate:     m.DueDate,
		CompletedAt: m.CompletedAt, CompletedBy: m.CompletedBy,
		Task: task, User: user,
		CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt,
	}
}

// UserFromModel converts a model.UserModel to the API DTO.
func UserFromModel(m *model.UserModel) *User {
	roles := make([]string, 0, len(m.Roles))
	for _, ur := range m.Roles {
		roles = append(roles, ur.Role.Name)
	}
	return &User{
		ID:              m.ID,
		DisplayName:     m.DisplayName,
		Email:           m.Email,
		Phone:           m.Phone,
		AvatarURL:       m.AvatarURL,
		DefaultFamilyID: m.DefaultFamilyID,
		Roles:           roles,
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
	}
}
