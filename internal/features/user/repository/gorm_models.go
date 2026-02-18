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
