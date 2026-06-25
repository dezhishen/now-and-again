package repository

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"gorm.io/gorm"
)

// ─── ApiKeyRepo ──────────────────────────────────────────────────

// CreateApiKey generates a new API key. Returns the raw key (shown once).
func (r *ApiKeyRepo) CreateApiKey(userID, name string, scopesJSON string, expiresAt *time.Time) (string, *ApiKeyModel, error) {
	raw := "na_" + randomHex(32) // na_ prefix, 64 hex chars
	hash := hashToken(raw)

	m := &ApiKeyModel{
		UserID:    userID,
		Name:      name,
		KeyHash:   hash,
		KeyPrefix: raw[:11], // "na_" + first 8 chars
		Scopes:    scopesJSON,
		ExpiresAt: expiresAt,
	}
	if err := r.db.Create(m).Error; err != nil {
		return "", nil, err
	}
	return raw, m, nil
}

// ValidateApiKeyRaw is the middleware-compatible version returning userID as string.
func (r *ApiKeyRepo) ValidateApiKey(raw string) (userID string, err error) {
	m, err := r.ValidateApiKeyModel(raw)
	if err != nil || m == nil {
		return "", err
	}
	return m.UserID, nil
}

// ValidateApiKeyModel looks up a raw API key and returns the full model.
func (r *ApiKeyRepo) ValidateApiKeyModel(raw string) (*ApiKeyModel, error) {
	hash := hashToken(raw)
	var m ApiKeyModel
	err := r.db.Where("key_hash = ? AND revoked = false", hash).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if m.ExpiresAt != nil && time.Now().After(*m.ExpiresAt) {
		return nil, nil
	}
	// Update last used
	r.db.Model(&m).Update("last_used_at", time.Now())
	return &m, nil
}

// ListApiKeys returns all API keys for a user (without the secret).
func (r *ApiKeyRepo) ListApiKeys(userID string) ([]ApiKeyModel, error) {
	var keys []ApiKeyModel
	err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&keys).Error
	return keys, err
}

// RevokeApiKey marks a key as revoked.
func (r *ApiKeyRepo) RevokeApiKey(userID, keyID string) error {
	return r.db.Model(&ApiKeyModel{}).
		Where("id = ? AND user_id = ?", keyID, userID).
		Update("revoked", true).Error
}

func randomHex(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}
