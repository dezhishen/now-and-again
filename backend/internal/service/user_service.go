package service

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/dezhishen/now-and-again/backend/internal/repository"
	"github.com/dezhishen/now-and-again/shared/types"
)

// ─── Setup ────────────────────────────────────────────────────────

func (s *UserService) Setup(ctx context.Context, req *types.SetupRequest) (*types.User, error) {
	count, err := s.repo.Count()
	if err != nil {
		return nil, fmt.Errorf("check users: %w", err)
	}
	if count > 0 {
		return nil, fmt.Errorf("system already initialized")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &repository.UserModel{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hash),
		DisplayName:  req.DisplayName,
		IsAdmin:      true,
	}

	if err := s.repo.Create(user); err != nil {
		return nil, fmt.Errorf("create admin: %w", err)
	}
	return modelToUser(user), nil
}

// ─── CheckInit ────────────────────────────────────────────────────

func (s *UserService) CheckInit(ctx context.Context) (*types.SystemStatus, error) {
	count, err := s.repo.Count()
	if err != nil {
		return nil, fmt.Errorf("check init: %w", err)
	}
	return &types.SystemStatus{Initialized: count > 0}, nil
}

// ─── Register ─────────────────────────────────────────────────────

func (s *UserService) Register(ctx context.Context, req *types.CreateUserRequest) (*types.User, error) {
	count, err := s.repo.Count()
	if err != nil {
		return nil, fmt.Errorf("check init: %w", err)
	}
	if count == 0 {
		return nil, fmt.Errorf("system not initialized, please run setup first")
	}

	if existing, _ := s.repo.FindByUsername(req.Username); existing != nil {
		return nil, fmt.Errorf("username already taken")
	}
	if existing, _ := s.repo.FindByEmail(req.Email); existing != nil {
		return nil, fmt.Errorf("email already registered")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &repository.UserModel{
		Username:     req.Username,
		Email:        req.Email,
		Phone:        req.Phone,
		PasswordHash: string(hash),
		DisplayName:  req.DisplayName,
		IsAdmin:      false,
	}

	if err := s.repo.Create(user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return modelToUser(user), nil
}

// ─── Login ────────────────────────────────────────────────────────

const (
	accessTokenTTL  = 15 * time.Minute
	refreshTokenTTL = 7 * 24 * time.Hour
)

func (s *UserService) Login(ctx context.Context, req *types.LoginRequest) (*types.TokenPair, error) {
	user, err := s.repo.FindByUsername(req.Username)
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("invalid username or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("invalid username or password")
	}

	return s.generateTokenPair(user)
}

// ─── Refresh ──────────────────────────────────────────────────────

func (s *UserService) Refresh(ctx context.Context, refreshToken string) (*types.TokenPair, error) {
	rt, err := s.repo.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired refresh token")
	}

	// Rotate: revoke old, issue new
	_ = s.repo.RevokeRefreshToken(refreshToken)

	user, err := s.repo.FindByID(rt.UserID)
	if err != nil || user == nil {
		return nil, fmt.Errorf("user not found")
	}

	return s.generateTokenPair(user)
}

// ─── Logout ───────────────────────────────────────────────────────

func (s *UserService) Logout(ctx context.Context) error {
	if rt, ok := ctx.Value("refresh_token").(string); ok && rt != "" {
		_ = s.repo.RevokeRefreshToken(rt)
	}
	return nil
}

// ─── Token generation ─────────────────────────────────────────────

func (s *UserService) generateTokenPair(user *repository.UserModel) (*types.TokenPair, error) {
	// Access token (short-lived JWT)
	now := time.Now()
	accessExp := now.Add(accessTokenTTL)
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"iat": now.Unix(),
		"exp": accessExp.Unix(),
	})
	accessStr, err := accessToken.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, fmt.Errorf("sign access token: %w", err)
	}

	// Refresh token (opaque, stored in DB)
	refreshStr, err := s.repo.CreateRefreshToken(user.ID, refreshTokenTTL)
	if err != nil {
		return nil, fmt.Errorf("create refresh token: %w", err)
	}

	return &types.TokenPair{
		AccessToken:  accessStr,
		RefreshToken: refreshStr,
		ExpiresIn:    int(accessTokenTTL.Seconds()),
		User:         *modelToUser(user),
	}, nil
}

// ─── GetMe ────────────────────────────────────────────────────────

func (s *UserService) GetMe(ctx context.Context) (*types.User, error) {
	return nil, fmt.Errorf("not implemented")
}

// ─── UpdateMe ─────────────────────────────────────────────────────

func (s *UserService) UpdateMe(ctx context.Context, req *types.UpdateUserRequest) (*types.User, error) {
	return nil, fmt.Errorf("not implemented")
}

// ─── Helpers ──────────────────────────────────────────────────────

func modelToUser(m *repository.UserModel) *types.User {
	return &types.User{
		ID:          uuid.MustParse(m.ID),
		Username:    m.Username,
		Email:       m.Email,
		Phone:       m.Phone,
		DisplayName: m.DisplayName,
		AvatarURL:   m.AvatarURL,
		IsAdmin:     m.IsAdmin,
	}
}

// ─── ListUsers (admin) ────────────────────────────────────────────

func (s *UserService) ListUsers(ctx context.Context) ([]types.User, error) {
	models, err := s.repo.ListAll()
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	users := make([]types.User, len(models))
	for i, m := range models {
		u := modelToUser(&m)
		users[i] = *u
	}
	return users, nil
}
