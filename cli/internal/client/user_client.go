package client

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dezhishen/now-and-again/shared/types"
)

// ─── User ────────────────────────────────────────────────────────

func (c *UserClient) Setup(ctx context.Context, req *types.SetupRequest) (*types.User, error) {
	return nil, fmt.Errorf("setup is a web-only operation, use the web UI")
}

func (c *UserClient) CheckInit(ctx context.Context) (*types.SystemStatus, error) {
	return nil, fmt.Errorf("not implemented")
}

func (c *UserClient) Register(ctx context.Context, req *types.CreateUserRequest) (*types.User, error) {
	data, err := c.http.Post("/api/auth/register", req)
	if err != nil {
		return nil, err
	}
	var resp struct{ Data types.User }
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

func (c *UserClient) Login(ctx context.Context, req *types.LoginRequest) (*types.TokenPair, error) {
	data, err := c.http.Post("/api/auth/login", req)
	if err != nil {
		return nil, err
	}
	var resp struct{ Data types.TokenPair }
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

func (c *UserClient) Refresh(ctx context.Context, refreshToken string) (*types.TokenPair, error) {
	return nil, fmt.Errorf("not implemented")
}

func (c *UserClient) Logout(ctx context.Context) error {
	return fmt.Errorf("not implemented")
}

func (c *UserClient) GetMe(ctx context.Context) (*types.User, error) {
	return nil, fmt.Errorf("not implemented")
}

func (c *UserClient) UpdateMe(ctx context.Context, req *types.UpdateUserRequest) (*types.User, error) {
	return nil, fmt.Errorf("not implemented")
}

func (c *UserClient) ListUsers(ctx context.Context) ([]types.User, error) {
	return nil, fmt.Errorf("not implemented")
}

// ─── API Key (CLI client) ────────────────────────────────────────

func (c *Client) SetApiKey(key string) {
	c.Token = ""
	c.customHeader = map[string]string{"X-API-Key": key}
}
