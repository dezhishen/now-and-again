package repository

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ─── API Key ──────────────────────────────────────────────────────

func (r *ApiKeyRepo) CreateApiKey(userID, name string, expiresAt *time.Time) (*ApiKeyModel, string, error) {
	raw := "na_" + uuid.New().String()
	prefix := raw[:12]
	keyHash := hashToken(raw)

	ak := &ApiKeyModel{
		UserID:    userID,
		Name:      name,
		KeyPrefix: prefix,
		KeyHash:   keyHash,
		ExpiresAt: expiresAt,
	}
	if err := r.db.Create(ak).Error; err != nil {
		return nil, "", fmt.Errorf("create api key: %w", err)
	}
	return ak, raw, nil
}

func (r *ApiKeyRepo) ValidateApiKey(raw string) (userID string, err error) {
	var ak ApiKeyModel

	prefix := ""
	if len(raw) >= 12 {
		prefix = raw[:12]
		err = r.db.Where("key_prefix = ? AND revoked = ?", prefix, false).First(&ak).Error
	} else {
		err = r.db.Where("key_hash = ? AND revoked = ?", hashToken(raw), false).First(&ak).Error
	}
	if err != nil {
		return "", fmt.Errorf("invalid api key")
	}

	if hashToken(raw) != ak.KeyHash {
		return "", fmt.Errorf("invalid api key")
	}

	if ak.ExpiresAt != nil && ak.ExpiresAt.Before(time.Now()) {
		return "", fmt.Errorf("api key expired")
	}

	now := time.Now()
	r.db.Model(&ak).Update("last_used_at", now)

	return ak.UserID, nil
}

func (r *ApiKeyRepo) ListByUser(userID string) ([]ApiKeyModel, error) {
	var keys []ApiKeyModel
	err := r.db.Where("user_id = ? AND revoked = ?", userID, false).
		Order("created_at DESC").Find(&keys).Error
	return keys, err
}

func (r *ApiKeyRepo) Revoke(keyID, userID string) error {
	return r.db.Model(&ApiKeyModel{}).
		Where("id = ? AND user_id = ?", keyID, userID).
		Update("revoked", true).Error
}
