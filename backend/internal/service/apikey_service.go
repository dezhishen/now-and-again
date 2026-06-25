package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dezhishen/now-and-again/backend/internal/repository"
	"github.com/dezhishen/now-and-again/shared/types"
)

// ─── ApiKeyService ────────────────────────────────────────────────

type ApiKeyService struct {
	repo *repository.ApiKeyRepo
}

func NewApiKeyService(repo *repository.ApiKeyRepo) *ApiKeyService {
	return &ApiKeyService{repo: repo}
}

func (s *ApiKeyService) Create(ctx context.Context, req *types.CreateApiKeyRequest) (*types.CreateApiKeyResponse, error) {
	userID := userIDFromCtx(ctx)

	var expiresAt *time.Time
	if req.ExpiresAt != nil && *req.ExpiresAt != "" {
		t, err := time.Parse(time.RFC3339, *req.ExpiresAt)
		if err == nil {
			expiresAt = &t
		}
	}

	raw, model, err := s.repo.CreateApiKey(userID, req.Name, "", expiresAt)
	if err != nil {
		return nil, fmt.Errorf("create api key: %w", err)
	}

	var lastUsed *string
	if model.LastUsedAt != nil {
		s := model.LastUsedAt.Format(time.RFC3339)
		lastUsed = &s
	}
	var exp *string
	if model.ExpiresAt != nil {
		s := model.ExpiresAt.Format(time.RFC3339)
		exp = &s
	}

	return &types.CreateApiKeyResponse{
		ApiKey: types.ApiKey{
			ID:         model.ID,
			Name:       model.Name,
			KeyPrefix:  model.KeyPrefix,
			LastUsedAt: lastUsed,
			ExpiresAt:  exp,
			CreatedAt:  model.CreatedAt.Format(time.RFC3339),
		},
		RawKey: raw,
	}, nil
}

func (s *ApiKeyService) List(ctx context.Context) ([]types.ApiKey, error) {
	userID := userIDFromCtx(ctx)
	models, err := s.repo.ListApiKeys(userID)
	if err != nil {
		return nil, err
	}
	keys := make([]types.ApiKey, len(models))
	for i, m := range models {
		var lastUsed *string
		if m.LastUsedAt != nil {
			s := m.LastUsedAt.Format(time.RFC3339)
			lastUsed = &s
		}
		var exp *string
		if m.ExpiresAt != nil {
			s := m.ExpiresAt.Format(time.RFC3339)
			exp = &s
		}
		var scopes *string
		if m.Scopes != "" {
			scopes = &m.Scopes
		}
		keys[i] = types.ApiKey{
			ID:         m.ID,
			Name:       m.Name,
			KeyPrefix:  m.KeyPrefix,
			Scopes:     scopes,
			LastUsedAt: lastUsed,
			ExpiresAt:  exp,
			Revoked:    m.Revoked,
			CreatedAt:  m.CreatedAt.Format(time.RFC3339),
		}
	}
	return keys, nil
}

func (s *ApiKeyService) Revoke(ctx context.Context, keyID string) error {
	userID := userIDFromCtx(ctx)
	return s.repo.RevokeApiKey(userID, keyID)
}

// Marshal scopes to JSON string for storage.
func marshalScopes(scopes []string) string {
	if len(scopes) == 0 {
		return ""
	}
	data, _ := json.Marshal(scopes)
	return string(data)
}
