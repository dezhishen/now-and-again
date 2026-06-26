package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/dezhishen/now-and-again/backend/internal/repository"
	"github.com/dezhishen/now-and-again/shared/scopes"
	"github.com/dezhishen/now-and-again/shared/types"
)

func (s *ApiKeyService) Create(ctx context.Context, req *types.CreateApiKeyRequest) (*types.CreateApiKeyResponse, error) {
	userID := ctx.Value("user_id")
	if userID == nil {
		return nil, fmt.Errorf("not authenticated")
	}

	ak, raw, err := s.repo.CreateApiKey(userID.(string), req.Name, scopes.ExpandGroups(req.Scopes), req.ExpiresAt)
	if err != nil {
		return nil, fmt.Errorf("create api key: %w", err)
	}

	return &types.CreateApiKeyResponse{
		ApiKey: &types.ApiKey{
			ID:         ak.ID,
			Name:       ak.Name,
			KeyPrefix:  ak.KeyPrefix,
			RawKey:     raw,
			Scopes:     repository.UnmarshalScopes(ak.Scopes),
			LastUsedAt: ak.LastUsedAt,
			ExpiresAt:  ak.ExpiresAt,
			CreatedAt:  ak.CreatedAt,
		},
		Message: "API Key created. Store it safely — the full key won't be shown again.",
	}, nil
}

func (s *ApiKeyService) List(ctx context.Context) ([]types.ApiKey, error) {
	userID := ctx.Value("user_id")
	if userID == nil {
		return nil, fmt.Errorf("not authenticated")
	}

	keys, err := s.repo.ListByUser(userID.(string))
	if err != nil {
		return nil, fmt.Errorf("list api keys: %w", err)
	}

	result := make([]types.ApiKey, len(keys))
	for i, k := range keys {
		result[i] = types.ApiKey{
			ID:         k.ID,
			Name:       k.Name,
			KeyPrefix:  k.KeyPrefix,
			Scopes:     repository.UnmarshalScopes(k.Scopes),
			LastUsedAt: k.LastUsedAt,
			ExpiresAt:  k.ExpiresAt,
			CreatedAt:  k.CreatedAt,
		}
	}
	return result, nil
}

func (s *ApiKeyService) Revoke(ctx context.Context, keyID uuid.UUID) error {
	userID := ctx.Value("user_id")
	if userID == nil {
		return fmt.Errorf("not authenticated")
	}

	if err := s.repo.Revoke(keyID.String(), userID.(string)); err != nil {
		return fmt.Errorf("revoke api key: %w", err)
	}
	return nil
}
