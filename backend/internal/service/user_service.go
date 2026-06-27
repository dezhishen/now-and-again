package service

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/dezhishen/now-and-again/backend/internal/repository"
	"github.com/dezhishen/now-and-again/backend/pkg/timeutil"
	"github.com/dezhishen/now-and-again/backend/pkg/types"
)

const (
	accessTokenTTL  = 15 * time.Minute
	refreshTokenTTL = 7 * 24 * time.Hour
)

// ─── Helpers ──────────────────────────────────────────────────────

func userModelToUser(m *repository.UserModel) *types.User {
	roles := make([]string, 0, len(m.Roles))
	for _, ur := range m.Roles {
		roles = append(roles, ur.Role.Name)
	}
	return &types.User{
		ID:          m.ID,
		DisplayName: m.DisplayName,
		Email:       m.Email,
		Phone:       m.Phone,
		AvatarURL:   m.AvatarURL,
		Roles:       roles,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

func (s *UserService) generateTokens(ctx context.Context, userID string) (*types.TokenPair, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"iat": timeutil.Now().Unix(),
		"exp": timeutil.Now().Add(accessTokenTTL).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, fmt.Errorf("sign token: %w", err)
	}

	refreshToken, err := s.repo.CreateRefreshToken(userID, refreshTokenTTL)
	if err != nil {
		return nil, fmt.Errorf("create refresh token: %w", err)
	}

	return &types.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(accessTokenTTL.Seconds()),
	}, nil
}

// ─── Setup ────────────────────────────────────────────────────────

func (s *UserService) Setup(ctx context.Context, req *types.SetupRequest) (*types.User, error) {
	count, err := s.repo.CountUsers()
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

	var userID string
	err = s.repo.Tx(func(tx *gorm.DB) error {
		user := &repository.UserModel{
			DisplayName: req.DisplayName,
			Email:       req.Email,
		}
		if err := tx.Create(user).Error; err != nil {
			return fmt.Errorf("create user: %w", err)
		}
		userID = user.ID

		acc := &repository.AccountModel{
			UserID:       user.ID,
			Provider:     "local",
			Username:     req.Username,
			PasswordHash: string(hash),
		}
		if err := tx.Create(acc).Error; err != nil {
			return fmt.Errorf("create account: %w", err)
		}

		var role repository.RoleModel
		if err := tx.Where("name = ?", "admin").First(&role).Error; err != nil {
			return fmt.Errorf("find admin role: %w", err)
		}
		if err := tx.Create(&repository.UserRoleModel{UserID: user.ID, RoleID: role.ID}).Error; err != nil {
			return fmt.Errorf("assign role: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	loaded, err := s.repo.FindUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("reload user: %w", err)
	}
	return userModelToUser(loaded), nil
}

// ─── CheckInit ────────────────────────────────────────────────────

func (s *UserService) CheckInit(ctx context.Context) (*types.SystemStatus, error) {
	count, err := s.repo.CountUsers()
	if err != nil {
		return nil, fmt.Errorf("check init: %w", err)
	}
	return &types.SystemStatus{Initialized: count > 0}, nil
}

// ─── Register ─────────────────────────────────────────────────────

func (s *UserService) Register(ctx context.Context, req *types.CreateUserRequest) (*types.User, error) {
	count, err := s.repo.CountUsers()
	if err != nil {
		return nil, fmt.Errorf("check init: %w", err)
	}
	if count == 0 {
		return nil, fmt.Errorf("system not initialized, please run setup first")
	}

	if existing, _ := s.repo.FindAccountByUsername(req.Username); existing != nil {
		return nil, fmt.Errorf("username already taken")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	var userID string
	err = s.repo.Tx(func(tx *gorm.DB) error {
		user := &repository.UserModel{
			DisplayName: req.DisplayName,
			Email:       req.Email,
			Phone:       req.Phone,
		}
		if err := tx.Create(user).Error; err != nil {
			return fmt.Errorf("create user: %w", err)
		}
		userID = user.ID

		acc := &repository.AccountModel{
			UserID:       user.ID,
			Provider:     "local",
			Username:     req.Username,
			PasswordHash: string(hash),
		}
		if err := tx.Create(acc).Error; err != nil {
			return fmt.Errorf("create account: %w", err)
		}

		var role repository.RoleModel
		if err := tx.Where("name = ?", "user").First(&role).Error; err != nil {
			return fmt.Errorf("find user role: %w", err)
		}
		if err := tx.Create(&repository.UserRoleModel{UserID: user.ID, RoleID: role.ID}).Error; err != nil {
			return fmt.Errorf("assign role: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	loaded, err := s.repo.FindUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("reload user: %w", err)
	}
	return userModelToUser(loaded), nil
}

// ─── Login ────────────────────────────────────────────────────────

func (s *UserService) Login(ctx context.Context, req *types.LoginRequest) (*types.TokenPair, error) {
	acc, err := s.repo.FindAccountByUsername(req.Username)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(acc.PasswordHash), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	user, err := s.repo.FindUserByID(acc.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	pair, err := s.generateTokens(ctx, acc.UserID)
	if err != nil {
		return nil, err
	}
	pair.User = userModelToUser(user)
	return pair, nil
}

// ─── Refresh ──────────────────────────────────────────────────────

func (s *UserService) Refresh(ctx context.Context, refreshToken string) (*types.TokenPair, error) {
	userID, err := s.repo.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token")
	}

	_ = s.repo.RevokeRefreshToken(refreshToken)

	pair, err := s.generateTokens(ctx, userID)
	if err != nil {
		return nil, err
	}

	user, err := s.repo.FindUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}
	pair.User = userModelToUser(user)
	return pair, nil
}

// ─── Logout ───────────────────────────────────────────────────────

func (s *UserService) Logout(ctx context.Context) error {
	return nil
}

// ─── GetMe ────────────────────────────────────────────────────────

func (s *UserService) GetMe(ctx context.Context) (*types.User, error) {
	userID := ctx.Value("user_id")
	if userID == nil {
		return nil, fmt.Errorf("not authenticated")
	}

	user, err := s.repo.FindUserByID(userID.(string))
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}
	return userModelToUser(user), nil
}

// ─── UpdateMe ─────────────────────────────────────────────────────

func (s *UserService) UpdateMe(ctx context.Context, req *types.UpdateUserRequest) (*types.User, error) {
	userID := ctx.Value("user_id")
	if userID == nil {
		return nil, fmt.Errorf("not authenticated")
	}

	user, err := s.repo.FindUserByID(userID.(string))
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	if req.DisplayName != nil {
		user.DisplayName = *req.DisplayName
	}
	if req.Email != nil {
		user.Email = *req.Email
	}
	if req.Phone != nil {
		user.Phone = *req.Phone
	}
	if req.AvatarURL != nil {
		user.AvatarURL = *req.AvatarURL
	}

	if err := s.repo.UpdateUser(user); err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}
	return userModelToUser(user), nil
}

// ─── ListUsers ────────────────────────────────────────────────────

func (s *UserService) ListUsers(ctx context.Context) ([]types.User, error) {
	users, err := s.repo.ListUsers()
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}

	result := make([]types.User, len(users))
	for i, u := range users {
		result[i] = *userModelToUser(&u)
	}
	return result, nil
}
