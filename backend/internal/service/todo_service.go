package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"database/sql"

	"github.com/dezhishen/now-and-again/backend/internal/repository"
	"github.com/dezhishen/now-and-again/backend/pkg/scheduler"
	"github.com/dezhishen/now-and-again/backend/pkg/types"
)

type TodoService struct {
	*taskOrchestrator
}

func NewTodoService(repo *repository.TaskRepo, familyRepo *repository.FamilyRepo) *TodoService {
	return &TodoService{taskOrchestrator: newTaskOrchestrator(repo, familyRepo)}
}

// ─── Todo ────────────────────────────────────────────────────────

func (s *TodoService) ListTodos(ctx context.Context, familyID uuid.UUID, groupID, status string) ([]types.Todo, error) {
	// Explicit group filter: query only that group's todos.
	if groupID != "" {
		todos, err := s.repo.ListTodosByGroup(familyID.String(), groupID, status)
		if err != nil {
			return nil, err
		}
		result := make([]types.Todo, 0, len(todos))
		for i := range todos {
			result = append(result, *todoModelToType(&todos[i]))
		}
		return result, nil
	}

	// No group filter: query ungrouped todos + todos in groups the user joined.
	userID, _ := ctx.Value("user_id").(string)
	userGroupIDs := []string{}
	if s.familyRepo != nil {
		ids, err := s.familyRepo.ListUserGroupIDs(userID, familyID.String())
		if err == nil {
			userGroupIDs = ids
		}
	}
	todos, err := s.repo.ListTodosByFamily(familyID.String(), status, userGroupIDs)
	if err != nil {
		return nil, err
	}
	result := make([]types.Todo, 0, len(todos))
	for i := range todos {
		result = append(result, *todoModelToType(&todos[i]))
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
		// Sync the in-memory todo with the just-persisted fields so OnTodo
		// receives the remark the user typed.
		todo.Remark = sql.NullString{String: todoFields.Remark, Valid: todoFields.Remark != ""}
		todo.Status = status
		todo.CompletedBy = sql.NullString{String: userID, Valid: userID != ""}

		action := "完成待办"
		if status == string(types.TodoStatusSkipped) {
			action = "跳过待办"
		}
		msg := fmt.Sprintf("%s: %s", action, todo.Task.Name)
		if todoFields.Remark != "" {
			msg += fmt.Sprintf(" | 备注: %s", todoFields.Remark)
		}
		s.repo.CreateUserLog(todo.TaskID, todoID.String(), userID, status, msg)

		// Dispatch: prefer CreatedByKind (composite/creator), fall back to Kind (real type).
		kind := todo.Task.CreatedByKind
		if kind == "" {
			kind = todo.Task.Kind
		}
		if h := s.taskManager.Get(kind); h != nil {
			h.OnTodo(s.taskStorage, todo, req.Extra)
		}

		// Notify scheduler: handler decides if this is a terminal event
		// (one-shot handlers unschedule, recurring handlers ignore).
		scheduler.MarkCompleted(todo.TaskID)
	}

	t, err := s.repo.FindTodoByID(todoID.String())
	if err != nil {
		return nil, err
	}
	return todoModelToType(t), nil
}

func todoModelToType(t *repository.TodoModel) *types.Todo {
	return types.TodoFromModel(t)
}
