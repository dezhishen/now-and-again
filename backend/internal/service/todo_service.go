package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/dezhishen/now-and-again/backend/internal/repository"
	"github.com/dezhishen/now-and-again/backend/pkg/scheduler"
	"github.com/dezhishen/now-and-again/backend/pkg/taskkind"
	"github.com/dezhishen/now-and-again/backend/pkg/types"
)

type TodoService struct {
	repo      *repository.TaskRepo
	scheduler *scheduler.Scheduler
	Ops       *taskkind.Ops
}

func NewTodoService(repo *repository.TaskRepo, sched *scheduler.Scheduler) *TodoService {
	return &TodoService{
		repo:      repo,
		scheduler: sched,
		Ops:       &taskkind.Ops{Repo: repo, Scheduler: sched},
	}
}

// ─── Todo ────────────────────────────────────────────────────────

func (s *TodoService) ListTodos(ctx context.Context, familyID uuid.UUID, groupID, status string) ([]types.Todo, error) {
	todos, err := s.repo.ListTodosByFamily(familyID.String(), status)
	if err != nil {
		return nil, err
	}
	result := make([]types.Todo, 0, len(todos))
	for _, t := range todos {
		if groupID != "" && t.Task.GroupID != "" && t.Task.GroupID != groupID {
			continue
		}
		result = append(result, *todoModelToType(&t))
	}
	return result, nil
}

func (s *TodoService) GetTodo(ctx context.Context, todoID uuid.UUID) (*types.Todo, error) {
	t, err := s.repo.FindTodoByID(todoID.String())
	if err != nil {
		return nil, err
	}
	return todoModelToType(t), nil
}

func (s *TodoService) GetTodoWithExtra(ctx context.Context, todoID uuid.UUID) (map[string]any, error) {
	t, err := s.repo.FindTodoFull(todoID.String())
	if err != nil {
		return nil, err
	}
	result := map[string]any{
		"todo": todoModelToType(t),
	}
	if h := taskkind.Get(t.Task.Kind); h != nil {
		extra, _ := h.GetExtra(s.Ops, &t.Task)
		if extra != nil {
			result["extra"] = extra
		}
	}
	return result, nil
}

func (s *TodoService) CompleteTodo(ctx context.Context, todoID uuid.UUID, req *types.CompleteTodoRequest) (*types.Todo, error) {
	userID := ctx.Value("user_id").(string)
	todo, err := s.repo.FindTodoFull(todoID.String())
	if err != nil {
		return nil, err
	}

	status := req.Status
	branchName := req.BranchName

	if err := s.repo.CompleteTodo(todoID.String(), userID, status, req.Remark); err != nil {
		return nil, err
	}

	msg := fmt.Sprintf("完成待办: %s", todo.Task.Name)
	if req.Remark != "" {
		msg += fmt.Sprintf(" | 备注: %s", req.Remark)
	}
	s.repo.CreateUserLog(todo.TaskID, todoID.String(), userID, status, msg)

	if h := taskkind.Get(todo.Task.Kind); h != nil {
		h.OnComplete(s.Ops, todo, req.Selections, branchName, userID)
	}

	s.disableCompletedOnceTask(todo)

	t, err := s.repo.FindTodoByID(todoID.String())
	if err != nil {
		return nil, err
	}
	return todoModelToType(t), nil
}

func (s *TodoService) disableCompletedOnceTask(todo *repository.TodoModel) {
	if todo.Task.ScheduleType != "once" {
		return
	}
	s.repo.DisableTask(todo.TaskID)
	s.scheduler.RemoveJob(todo.TaskID)
}

func todoModelToType(t *repository.TodoModel) *types.Todo {
	var task *types.TaskTemplate
	if t.Task.ID != "" {
		task = taskModelToType(&t.Task)
	}
	var user *types.User
	if t.User.ID != "" {
		user = userModelToUser(&t.User)
	}
	return &types.Todo{
		ID: t.ID, TaskID: t.TaskID, FamilyID: t.FamilyID,
		LocationID: t.LocationID,
		AssignedTo: t.AssignedTo, Status: t.Status, BranchName: t.BranchName, Remark: t.Remark,
		DueStart:    t.DueStart,
		DueDate:     t.DueDate,
		CompletedAt: t.CompletedAt, CompletedBy: t.CompletedBy,
		Task: task, User: user,
		CreatedAt: t.CreatedAt, UpdatedAt: t.UpdatedAt,
	}
}
