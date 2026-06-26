package client

import (
	"context"
	"fmt"

	"github.com/dezhishen/now-and-again/backend/pkg/types"
	"github.com/google/uuid"
)

// Compile-time check: ApiKeyClient must implement ApiKeyContract.

// ApiKeyClient is the HTTP implementation of ApiKeyContract.
type ApiKeyClient struct {
	http *HTTPClient
}

func NewApiKeyClient(http *HTTPClient) *ApiKeyClient {
	return &ApiKeyClient{http: http}
}

func (c *ApiKeyClient) Create(ctx context.Context, req *types.CreateApiKeyRequest) (*types.CreateApiKeyResponse, error) {
	var resp types.CreateApiKeyResponse
	if err := c.http.do("POST", "/api/users/me/api-keys", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *ApiKeyClient) List(ctx context.Context) ([]types.ApiKey, error) {
	var keys []types.ApiKey
	if err := c.http.do("GET", "/api/users/me/api-keys", nil, &keys); err != nil {
		return nil, err
	}
	return keys, nil
}

func (c *ApiKeyClient) Revoke(ctx context.Context, keyID uuid.UUID) error {
	return c.http.do("DELETE", fmt.Sprintf("/api/users/me/api-keys/%s", keyID), nil, nil)
}
