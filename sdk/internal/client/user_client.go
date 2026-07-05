package client

import (
	"context"
	"fmt"

	"github.com/dezhishen/now-and-again/backend/pkg/types"
)

// Compile-time check: UserClient must implement UserContract.

// UserClient is the HTTP implementation of UserContract.
type UserClient struct {
	http *HTTPClient
}

func NewUserClient(http *HTTPClient) *UserClient {
	return &UserClient{http: http}
}

func (c *UserClient) Register(ctx context.Context, req *types.CreateUserRequest) (*types.User, error) {
	var user types.User
	if err := c.http.do("POST", "/api/auth/register", req, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (c *UserClient) Login(ctx context.Context, req *types.LoginRequest) (*types.TokenPair, error) {
	var pair types.TokenPair
	if err := c.http.do("POST", "/api/auth/login", req, &pair); err != nil {
		return nil, err
	}
	return &pair, nil
}

func (c *UserClient) Refresh(ctx context.Context, refreshToken string) (*types.TokenPair, error) {
	// Refresh is handled via cookie in browser context; CLI uses API Key or Bearer token directly.
	// For CLI, we just return an error indicating this isn't supported.
	return nil, fmt.Errorf("refresh token flow not available in CLI; use 'na login' to get a new token")
}

func (c *UserClient) Logout(ctx context.Context) error {
	return c.http.do("POST", "/api/auth/logout", nil, nil)
}

func (c *UserClient) GetMe(ctx context.Context) (*types.User, error) {
	var user types.User
	if err := c.http.do("GET", "/api/users/me", nil, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (c *UserClient) UpdateMe(ctx context.Context, req *types.UpdateUserRequest) (*types.User, error) {
	var user types.User
	if err := c.http.do("PUT", "/api/users/me", req, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (c *UserClient) ListUsers(ctx context.Context) ([]types.User, error) {
	var users []types.User
	if err := c.http.do("GET", "/api/admin/users", nil, &users); err != nil {
		return nil, err
	}
	return users, nil
}
