package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/dezhishen/now-and-again/backend/internal/repository"
	"github.com/dezhishen/now-and-again/backend/pkg/types"
)

type LogService struct {
	repo *repository.TaskRepo
}

func NewLogService(repo *repository.TaskRepo) *LogService {
	return &LogService{repo: repo}
}

// ─── Logs ────────────────────────────────────────────────────────

func (s *LogService) ListLogs(ctx context.Context, taskID uuid.UUID, limit, offset int, userOnly bool) ([]types.TaskLog, error) {
	return s.ListTaskLogs(ctx, taskID, limit, offset, userOnly)
}

func (s *LogService) ListTaskLogs(ctx context.Context, taskID uuid.UUID, limit, offset int, userOnly bool) ([]types.TaskLog, error) {
	if limit <= 0 {
		limit = 50
	}
	var logs []repository.TaskLogModel
	var err error
	if userOnly {
		logs, err = s.repo.ListUserLogs(taskID.String(), limit, offset)
	} else {
		logs, err = s.repo.ListLogs(taskID.String(), limit, offset)
	}
	if err != nil {
		return nil, err
	}
	result := make([]types.TaskLog, len(logs))
	for i, l := range logs {
		result[i] = types.TaskLog{
			ID: l.ID, TaskID: l.TaskID, TaskName: l.Task.Name, TodoID: l.TodoID,
			Status: l.Status, Message: l.Message, LogType: l.LogType,
			OperatorID: l.OperatorID, CreatedAt: l.CreatedAt,
		}
	}
	return result, nil
}
