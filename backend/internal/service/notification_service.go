package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/dezhishen/now-and-again/shared/types"
)

// ─── Notification ────────────────────────────────────────────────

func (s *NotificationService) List(ctx context.Context, page, pageSize int) ([]types.Notification, int, error) {
	return nil, 0, fmt.Errorf("not implemented")
}

func (s *NotificationService) UpsertChannelConfig(ctx context.Context, req *types.UpsertUserChannelRequest) (*types.UserChannelConfig, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *NotificationService) ListTemplates(ctx context.Context, familyID uuid.UUID) ([]types.NotificationTemplate, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *NotificationService) UpsertTemplate(ctx context.Context, familyID uuid.UUID, req *types.UpsertTemplateRequest) (*types.NotificationTemplate, error) {
	return nil, fmt.Errorf("not implemented")
}
