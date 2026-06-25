package client

import (
	"context"
	"fmt"

	"github.com/dezhishen/now-and-again/shared/types"
)

func (c *ApiKeyClient) Create(ctx context.Context, req *types.CreateApiKeyRequest) (*types.CreateApiKeyResponse, error) {
	return nil, fmt.Errorf("api key management is web-only")
}

func (c *ApiKeyClient) List(ctx context.Context) ([]types.ApiKey, error) {
	return nil, fmt.Errorf("api key management is web-only")
}

func (c *ApiKeyClient) Revoke(ctx context.Context, keyID string) error {
	return fmt.Errorf("api key management is web-only")
}
