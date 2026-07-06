package service

import (
	"context"
	"encoding/json"
	"errors"
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
	return types.UserFromModel(m)
}

func (s *UserService) generateTokens(_ context.Context, userID string) (*types.TokenPair, error) {
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

// ─── Register ─────────────────────────────────────────────────────

func (s *UserService) Register(ctx context.Context, req *types.CreateUserRequest) (*types.User, error) {
	// Allow registration on first run (system auto-initializes via seedAdmin)
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
		return nil, err
	}
	if loaded == nil {
		return nil, fmt.Errorf("reload user: user not found")
	}
	return userModelToUser(loaded), nil
}

// ─── Login ────────────────────────────────────────────────────────

func (s *UserService) Login(ctx context.Context, req *types.LoginRequest) (*types.TokenPair, error) {
	acc, err := s.repo.FindAccountByUsername(req.Username)
	if err != nil {
		return nil, err
	}
	if acc == nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(acc.PasswordHash), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	user, err := s.repo.FindUserByID(acc.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
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
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.ErrRefreshTokenInvalid
		}
		return nil, fmt.Errorf("validate refresh token: %w", err)
	}

	_ = s.repo.RevokeRefreshToken(refreshToken)

	pair, err := s.generateTokens(ctx, userID)
	if err != nil {
		return nil, err
	}

	user, err := s.repo.FindUserByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
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
		return nil, err
	}
	if user == nil {
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
		return nil, err
	}
	if user == nil {
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
	if len(req.DefaultFamilyID) > 0 {
		if string(req.DefaultFamilyID) == "null" {
			// Explicitly cleared — set to nil
			user.DefaultFamilyID = nil
		} else {
			var id string
			if err := json.Unmarshal(req.DefaultFamilyID, &id); err == nil && id != "" {
				user.DefaultFamilyID = &id
			}
		}
	}

	if err := s.repo.UpdateUser(user); err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}
	return userModelToUser(user), nil
}

// ─── ListUsers (admin) ────────────────────────────────────────────

func (s *UserService) ListUsers(ctx context.Context, req *types.ListUsersRequest) (*types.ListUsersResponse, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}

	users, total, err := s.repo.SearchUsers(req.Query, req.Page, req.PageSize)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}

	result := make([]types.User, len(users))
	for i, u := range users {
		result[i] = *userModelToUser(&u)
	}

	totalPages := int(total) / req.PageSize
	if int(total)%req.PageSize > 0 {
		totalPages++
	}

	return &types.ListUsersResponse{
		Users:      result,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}, nil
}

// ─── ResetPassword (admin) ────────────────────────────────────────

func (s *UserService) ResetPassword(ctx context.Context, req *types.ResetPasswordRequest) (string, error) {
	acc, err := s.repo.FindAccountByUserID(req.UserID)
	if err != nil {
		return "", err
	}
	if acc == nil {
		return "", fmt.Errorf("account not found for user")
	}

	// Read default password from system settings
	setting, err := s.settingsRepo.Get("default_password")
	if err != nil || setting == nil || setting.Value == "" {
		return "", fmt.Errorf("default password not configured")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(setting.Value), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}

	acc.PasswordHash = string(hash)
	if err := s.repo.UpdateAccount(acc); err != nil {
		return "", fmt.Errorf("update account: %w", err)
	}
	return setting.Value, nil
}

// ─── ChangePassword ───────────────────────────────────────────────

func (s *UserService) ChangePassword(ctx context.Context, req *types.ChangePasswordRequest) error {
	userID := ctx.Value("user_id")
	if userID == nil {
		return fmt.Errorf("not authenticated")
	}

	acc, err := s.repo.FindAccountByUserID(userID.(string))
	if err != nil {
		return err
	}
	if acc == nil {
		return fmt.Errorf("account not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(acc.PasswordHash), []byte(req.OldPassword)); err != nil {
		return fmt.Errorf("current password is incorrect")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	acc.PasswordHash = string(hash)
	return s.repo.UpdateAccount(acc)
}

// ─── IsAdmin ──────────────────────────────────────────────────────

// IsAdmin returns true if the user has the "admin" role.
func (s *UserService) IsAdmin(userID string) bool {
	ok, err := s.repo.HasRole(userID, "admin")
	if err != nil {
		return false
	}
	return ok
}
