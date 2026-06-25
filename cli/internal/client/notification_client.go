package client

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/dezhishen/now-and-again/shared/types"
)

// ─── Notification ────────────────────────────────────────────────

func (c *NotificationClient) List(ctx context.Context, page, pageSize int) ([]types.Notification, int, error) {
	return nil, 0, fmt.Errorf("not implemented")
}
func (c *NotificationClient) UpsertChannelConfig(ctx context.Context, req *types.UpsertUserChannelRequest) (*types.UserChannelConfig, error) {
	return nil, fmt.Errorf("not implemented")
}
func (c *NotificationClient) ListTemplates(ctx context.Context, familyID uuid.UUID) ([]types.NotificationTemplate, error) {
	return nil, fmt.Errorf("not implemented")
}
func (c *NotificationClient) UpsertTemplate(ctx context.Context, familyID uuid.UUID, req *types.UpsertTemplateRequest) (*types.NotificationTemplate, error) {
	return nil, fmt.Errorf("not implemented")
}
