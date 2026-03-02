package model

import (
	"time"

	"github.com/google/uuid"
)

type RefreshToken struct {
	Token     string
	UserID    uuid.UUID
	ExpiresAt time.Time
	CreatedAt time.Time
	RevokedAt *time.Time
}

func NewRefreshToken(token string, userID uuid.UUID, ttl time.Duration) *RefreshToken {
	now := time.Now()
	return &RefreshToken{
		Token:     token,
		UserID:    userID,
		ExpiresAt: now.Add(ttl),
		CreatedAt: now,
	}
}

func (rt *RefreshToken) IsExpired() bool {
	return time.Now().After(rt.ExpiresAt)
}

func (rt *RefreshToken) IsRevoked() bool {
	return rt.RevokedAt != nil
}
