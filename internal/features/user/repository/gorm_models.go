package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/mairuu/mp-api/internal/features/user/model"
	"github.com/mairuu/mp-api/internal/platform/authorization"
)

type UserDB struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey"`
	Username     string    `gorm:"type:varchar(30);uniqueIndex;not null"`
	Email        string    `gorm:"type:varchar(255);uniqueIndex;not null"`
	PasswordHash string    `gorm:"type:varchar(255);not null"`
	Role         string    `gorm:"type:varchar(10);not null;default:'user'"`
	CreatedAt    time.Time `gorm:"not null"`
	UpdatedAt    time.Time `gorm:"not null"`
}

func (UserDB) TableName() string {
	return "users"
}

func toUserDB(u *model.User) UserDB {
	return UserDB{
		ID:           u.ID,
		Username:     u.Username,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		Role:         u.Role.String(),
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}
}

func toUserModel(udb *UserDB) model.User {
	return model.User{
		ID:           udb.ID,
		Username:     udb.Username,
		Email:        udb.Email,
		PasswordHash: udb.PasswordHash,
		Role:         authorization.Role(udb.Role),
		CreatedAt:    udb.CreatedAt,
		UpdatedAt:    udb.UpdatedAt,
	}
}

type RefreshTokenDB struct {
	Token     string     `gorm:"type:varchar(64);primaryKey"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null;index"`
	ExpiresAt time.Time  `gorm:"not null"`
	CreatedAt time.Time  `gorm:"not null"`
	RevokedAt *time.Time `gorm:"default:null"`
}

func (RefreshTokenDB) TableName() string {
	return "refresh_tokens"
}

func toRefreshTokenDB(rt *model.RefreshToken) RefreshTokenDB {
	return RefreshTokenDB{
		Token:     rt.Token,
		UserID:    rt.UserID,
		ExpiresAt: rt.ExpiresAt,
		CreatedAt: rt.CreatedAt,
		RevokedAt: rt.RevokedAt,
	}
}

func toRefreshTokenModel(rtdb *RefreshTokenDB) model.RefreshToken {
	return model.RefreshToken{
		Token:     rtdb.Token,
		UserID:    rtdb.UserID,
		ExpiresAt: rtdb.ExpiresAt,
		CreatedAt: rtdb.CreatedAt,
		RevokedAt: rtdb.RevokedAt,
	}
}
