package inspection

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/dezhishen/now-and-again/backend/internal/repository"
	"github.com/dezhishen/now-and-again/backend/pkg/taskkind"
	"github.com/dezhishen/now-and-again/backend/pkg/types"
)

// Service implements contracts.InspectionContract for the server side.
// It lives in the inspection package alongside the handler so all
// inspection logic stays cohesive.
type Service struct {
	ops *taskkind.Ops
}

// NewService creates a contract-compliant inspection service.
func NewService(ops *taskkind.Ops) *Service {
	return &Service{ops: ops}
}

// SubmitInspection implements contracts.InspectionContract.
func (s *Service) SubmitInspection(ctx context.Context, taskID uuid.UUID, req *types.SubmitInspectionRequest) (*types.Todo, error) {
	userID, _ := ctx.Value("user_id").(string)
	if userID == "" {
		return nil, fmt.Errorf("user not authenticated")
	}

	todo, err := s.ops.Repo.FindTodoByID(req.TodoID)
	if err != nil || todo.TaskID != taskID.String() {
		return nil, fmt.Errorf("todo not found")
	}

	if err := s.ops.Repo.CompleteTodo(req.TodoID, userID, "done", ""); err != nil {
		return nil, err
	}

	h := taskkind.Get("inspection")
	if h == nil {
		return nil, fmt.Errorf("inspection handler not registered")
	}
	insp, ok := h.(taskkind.Inspector)
	if !ok {
		return nil, fmt.Errorf("handler does not support inspection")
	}

	selections := make([]taskkind.Selection, len(req.Selections))
	for i, sel := range req.Selections {
		selections[i] = taskkind.Selection{Item: sel.Item, Branch: sel.Branch}
	}
	if err := insp.OnInspect(s.ops, todo, selections, userID); err != nil {
		return nil, err
	}

	if todo.Task.ScheduleType == "once" {
		s.ops.Repo.DisableTask(todo.TaskID)
		s.ops.Scheduler.RemoveJob(todo.TaskID)
	}

	todo, err = s.ops.Repo.FindTodoByID(req.TodoID)
	if err != nil {
		return nil, err
	}
	return todoToShared(todo), nil
}

func todoToShared(t *repository.TodoModel) *types.Todo {
	var task *types.TaskTemplate
	if t.Task.ID != "" {
		task = &types.TaskTemplate{
			ID: t.Task.ID, FamilyID: t.Task.FamilyID, GroupID: t.Task.GroupID,
			Name: t.Task.Name, ScheduleType: t.Task.ScheduleType,
			Enabled: t.Task.Enabled, Kind: t.Task.Kind,
			CreatedBy: t.Task.CreatedBy, CreatedAt: t.Task.CreatedAt, UpdatedAt: t.Task.UpdatedAt,
		}
	}
	return &types.Todo{
		ID: t.ID, TaskID: t.TaskID, FamilyID: t.FamilyID, LocationID: t.LocationID,
		Status: t.Status, BranchName: t.BranchName, Remark: t.Remark,
		DueStart: t.DueStart, DueDate: t.DueDate,
		CompletedAt: t.CompletedAt, CompletedBy: t.CompletedBy,
		Task: task, CreatedAt: t.CreatedAt, UpdatedAt: t.UpdatedAt,
	}
}
