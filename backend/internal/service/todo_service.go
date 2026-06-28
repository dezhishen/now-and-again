package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/dezhishen/now-and-again/backend/internal/repository"
	"github.com/dezhishen/now-and-again/backend/pkg/scheduler"
	"github.com/dezhishen/now-and-again/backend/pkg/types"
)

type TodoService struct {
	*taskOrchestrator
}

func NewTodoService(repo *repository.TaskRepo, sched *scheduler.Scheduler) *TodoService {
	return &TodoService{taskOrchestrator: newTaskOrchestrator(repo, sched)}
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

func (s *TodoService) GetTodoWithExtra(ctx context.Context, todoID uuid.UUID) (*types.TodoWithExtra, error) {
	t, err := s.repo.FindTodoFull(todoID.String())
	if err != nil {
		return nil, err
	}
	result := &types.TodoWithExtra{Todo: todoModelToType(t)}
	if h := s.taskManager.Get(t.Task.Kind); h != nil {
		result.Extra, _ = h.GetExtra(s.taskStorage, &t.Task)
	}
	return result, nil
}

func (s *TodoService) CompleteTodo(ctx context.Context, todoID uuid.UUID, req *types.CompleteTodoRequest) (*types.Todo, error) {
	userID := ctx.Value("user_id").(string)
	todo, err := s.repo.FindTodoFull(todoID.String())
	if err != nil {
		return nil, err
	}
	todoFields := req.Todo
	status := todoFields.Status

	updated, err := s.repo.CompleteTodo(todoID.String(), userID, status, todoFields.Remark)
	if err != nil {
		return nil, err
	}

	// Only create log and trigger plugins if the todo was actually pending.
	// Duplicate completions are silently ignored (idempotent).
	if updated {
		msg := fmt.Sprintf("完成待办: %s", todo.Task.Name)
		if todoFields.Remark != "" {
			msg += fmt.Sprintf(" | 备注: %s", todoFields.Remark)
		}
		s.repo.CreateUserLog(todo.TaskID, todoID.String(), userID, status, msg)
		todo.CompletedBy = userID
		if h := s.taskManager.Get(todo.Task.Kind); h != nil {
			h.OnComplete(s.taskStorage, todo, req.Extra)
		}

		s.disableCompletedOnceTask(todo)
	}

	t, err := s.repo.FindTodoByID(todoID.String())
	if err != nil {
		return nil, err
	}
	return todoModelToType(t), nil
}

func (s *TodoService) disableCompletedOnceTask(todo *repository.TodoModel) {
	s.scheduler.MarkCompleted(todo.TaskID)
}

func todoModelToType(t *repository.TodoModel) *types.Todo {
	return types.TodoFromModel(t)
}
