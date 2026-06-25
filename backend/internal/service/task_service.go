package service

import (
	"context"
	"fmt"
	"time"

	"github.com/dezhishen/now-and-again/backend/internal/repository"
	"github.com/dezhishen/now-and-again/shared/types"
	"github.com/google/uuid"
)

// ─── Helpers ──────────────────────────────────────────────────────

func userIDFromCtx(ctx context.Context) string {
	if v, ok := ctx.Value("user_id").(string); ok {
		return v
	}
	return ""
}

// ─── TaskService ──────────────────────────────────────────────────

func (s *TaskService) Create(ctx context.Context, req *types.CreateTaskRequest) (*types.Task, error) {
	userID := userIDFromCtx(ctx)

	t := &repository.TaskModel{
		FamilyID:         req.FamilyID.String(),
		SubGroupID:       uuidPtrToString(req.SubGroupID),
		TaskCode:         req.TaskCode,
		Title:            req.Title,
		Description:      req.Description,
		Status:           "todo",
		Priority:         string(req.Priority),
		DueDate:          req.DueDate,
		RecurrenceConfig: string(req.RecurrenceConfig),
		TypeSpecificData: string(req.TypeSpecificData),
		CreatedBy:        userID,
	}
	if t.Priority == "" {
		t.Priority = "medium"
	}

	if err := s.repo.Create(t); err != nil {
		return nil, fmt.Errorf("create task: %w", err)
	}

	// Assign users if specified
	if len(req.AssigneeIDs) > 0 {
		ids := make([]string, len(req.AssigneeIDs))
		for i, id := range req.AssigneeIDs {
			ids[i] = id.String()
		}
		_ = s.repo.SetAssignees(t.ID, ids)
	}

	return taskModelToType(t), nil
}

func (s *TaskService) List(ctx context.Context, familyID uuid.UUID, status *types.TaskStatus, assigneeID *uuid.UUID, page, pageSize int) ([]types.Task, int, error) {
	var st *string
	if status != nil {
		s := string(*status)
		st = &s
	}
	var aid *string
	if assigneeID != nil {
		s := assigneeID.String()
		aid = &s
	}
	offset := (page - 1) * pageSize

	models, total, err := s.repo.List(familyID.String(), st, aid, offset, pageSize)
	if err != nil {
		return nil, 0, err
	}

	tasks := make([]types.Task, len(models))
	for i, m := range models {
		tasks[i] = *taskModelToType(&m)
	}
	return tasks, int(total), nil
}

func (s *TaskService) Get(ctx context.Context, taskID uuid.UUID) (*types.Task, error) {
	t, err := s.repo.FindByID(taskID.String())
	if err != nil || t == nil {
		return nil, fmt.Errorf("task not found")
	}
	return taskModelToType(t), nil
}

func (s *TaskService) Update(ctx context.Context, taskID uuid.UUID, req *types.UpdateTaskRequest) (*types.Task, error) {
	t, err := s.repo.FindByID(taskID.String())
	if err != nil || t == nil {
		return nil, fmt.Errorf("task not found")
	}

	if req.Title != nil {
		t.Title = *req.Title
	}
	if req.Description != nil {
		t.Description = *req.Description
	}
	if req.Status != nil {
		t.Status = string(*req.Status)
		if *req.Status == types.TaskStatusDone {
			now := time.Now()
			t.CompletedAt = &now
		}
	}
	if req.Priority != nil {
		t.Priority = string(*req.Priority)
	}

	if err := s.repo.Update(t); err != nil {
		return nil, fmt.Errorf("update task: %w", err)
	}
	return taskModelToType(t), nil
}

func (s *TaskService) SetAssignees(ctx context.Context, taskID uuid.UUID, req *types.SetAssigneesRequest) ([]types.TaskAssignee, error) {
	ids := make([]string, len(req.UserIDs))
	for i, id := range req.UserIDs {
		ids[i] = id.String()
	}
	if err := s.repo.SetAssignees(taskID.String(), ids); err != nil {
		return nil, err
	}
	models, err := s.repo.ListAssignees(taskID.String())
	if err != nil {
		return nil, err
	}
	assignees := make([]types.TaskAssignee, len(models))
	for i, m := range models {
		assignees[i] = types.TaskAssignee{
			ID:         uuid.MustParse(m.ID),
			TaskID:     uuid.MustParse(m.TaskID),
			UserID:     uuid.MustParse(m.UserID),
			AssignedAt: m.AssignedAt.Format(time.RFC3339),
			User:       modelToUser(&m.User),
		}
	}
	return assignees, nil
}

func (s *TaskService) AddDependency(ctx context.Context, taskID uuid.UUID, req *types.CreateDependencyRequest) (*types.TaskDependency, error) {
	dep := &repository.TaskDependencyModel{
		BlockedTaskID:  req.BlockedTaskID.String(),
		BlockerTaskID:  req.BlockerTaskID.String(),
		DependencyType: string(req.DependencyType),
	}
	if err := s.repo.CreateDependency(dep); err != nil {
		return nil, err
	}
	return &types.TaskDependency{
		ID:             uuid.MustParse(dep.ID),
		BlockedTaskID:  uuid.MustParse(dep.BlockedTaskID),
		BlockerTaskID:  uuid.MustParse(dep.BlockerTaskID),
		DependencyType: types.DependencyType(dep.DependencyType),
	}, nil
}

func (s *TaskService) RemoveDependency(ctx context.Context, taskID, depID uuid.UUID) error {
	return s.repo.RemoveDependency(depID.String())
}

// ─── Model conversion ─────────────────────────────────────────────

func taskModelToType(m *repository.TaskModel) *types.Task {
	t := &types.Task{
		ID:          uuid.MustParse(m.ID),
		FamilyID:    uuid.MustParse(m.FamilyID),
		TaskCode:    m.TaskCode,
		Title:       m.Title,
		Description: m.Description,
		Status:      types.TaskStatus(m.Status),
		Priority:    types.Priority(m.Priority),
		DueDate:     m.DueDate,
		CreatedBy:   uuid.MustParse(m.CreatedBy),
		CompletedAt: m.CompletedAt,
		Timestamps:  types.Timestamps{CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt},
	}
	if m.SubGroupID != nil {
		id := uuid.MustParse(*m.SubGroupID)
		t.SubGroupID = &id
	}
	for _, a := range m.Assignees {
		t.Assignees = append(t.Assignees, types.TaskAssignee{
			ID: uuid.MustParse(a.ID), TaskID: uuid.MustParse(a.TaskID),
			UserID: uuid.MustParse(a.UserID), AssignedAt: a.AssignedAt.Format(time.RFC3339),
			User: modelToUser(&a.User),
		})
	}
	return t
}

func uuidPtrToString(id *uuid.UUID) *string {
	if id == nil {
		return nil
	}
	s := id.String()
	return &s
}
