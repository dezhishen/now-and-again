package repository

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
)

// ─── RefreshTokenRepo ─────────────────────────────────────────────

// CreateRefreshToken generates a new refresh token for a user.
// Returns the raw token (to send to client) and stores the hash.
func (r *UserRepo) CreateRefreshToken(userID string, ttl time.Duration) (string, error) {
	raw := uuid.New().String() + uuid.New().String() // 72-char random
	hash := hashToken(raw)

	rt := &RefreshTokenModel{
		UserID:    userID,
		TokenHash: hash,
		ExpiresAt: time.Now().Add(ttl),
	}
	if err := r.db.Create(rt).Error; err != nil {
		return "", err
	}
	return raw, nil
}

// ValidateRefreshToken checks if a raw token is valid and not revoked.
// Returns the token model if valid, or nil if invalid/expired/revoked.
func (r *UserRepo) ValidateRefreshToken(raw string) (*RefreshTokenModel, error) {
	hash := hashToken(raw)
	var rt RefreshTokenModel
	err := r.db.Where("token_hash = ? AND expires_at > ? AND revoked = false", hash, time.Now()).First(&rt).Error
	if err != nil {
		return nil, err // gorm.ErrRecordNotFound = invalid
	}
	return &rt, nil
}

// RevokeRefreshToken marks a token as revoked (used during rotation).
func (r *UserRepo) RevokeRefreshToken(raw string) error {
	hash := hashToken(raw)
	return r.db.Model(&RefreshTokenModel{}).Where("token_hash = ?", hash).Update("revoked", true).Error
}

// RevokeAllUserTokens revokes all refresh tokens for a user (logout all devices).
func (r *UserRepo) RevokeAllUserTokens(userID string) error {
	return r.db.Model(&RefreshTokenModel{}).Where("user_id = ?", userID).Update("revoked", true).Error
}

func hashToken(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}
