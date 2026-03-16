package models

import (
	"time"

	"github.com/google/uuid"
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
