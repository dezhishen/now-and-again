package repository

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/dezhishen/now-and-again/backend/pkg/timeutil"
	"github.com/google/uuid"
)

func hashToken(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}

// ─── Refresh Token ────────────────────────────────────────────────

func (r *UserRepo) CreateRefreshToken(userID string, ttl time.Duration) (raw string, err error) {
	raw = uuid.New().String() + uuid.New().String()
	rt := &RefreshTokenModel{
		ID:        uuid.New().String(),
		UserID:    userID,
		TokenHash: hashToken(raw),
		ExpiresAt: timeutil.Now().Add(ttl),
		CreatedAt: timeutil.Now(),
	}
	if err := r.db.Create(rt).Error; err != nil {
		return "", err
	}
	return raw, nil
}

func (r *UserRepo) ValidateRefreshToken(raw string) (userID string, err error) {
	var rt RefreshTokenModel
	err = r.db.Where("token_hash = ? AND revoked = ? AND expires_at > ?",
		hashToken(raw), false, timeutil.Now()).First(&rt).Error
	if err != nil {
		return "", err
	}
	return rt.UserID, nil
}

func (r *UserRepo) RevokeRefreshToken(raw string) error {
	return r.db.Model(&RefreshTokenModel{}).
		Where("token_hash = ?", hashToken(raw)).
		Update("revoked", true).Error
}

func (r *UserRepo) RevokeAllUserTokens(userID string) error {
	return r.db.Model(&RefreshTokenModel{}).
		Where("user_id = ? AND revoked = ?", userID, false).
		Update("revoked", true).Error
}
