package mappers

import (
	"github.com/mairuu/mp-api/internal/features/user/model"
	"github.com/mairuu/mp-api/internal/persistence/models"
	"github.com/mairuu/mp-api/internal/platform/authorization"
)

func ToUserDB(u *model.User) models.UserDB {
	return models.UserDB{
		ID:           u.ID,
		Username:     u.Username,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		Role:         u.Role.String(),
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}
}

func UserDBToModel(udb *models.UserDB) model.User {
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

func ToRefreshTokenDB(rt *model.RefreshToken) models.RefreshTokenDB {
	return models.RefreshTokenDB{
		Token:     rt.Token,
		UserID:    rt.UserID,
		ExpiresAt: rt.ExpiresAt,
		CreatedAt: rt.CreatedAt,
		RevokedAt: rt.RevokedAt,
	}
}

func RefreshTokenDBToModel(rtdb *models.RefreshTokenDB) model.RefreshToken {
	return model.RefreshToken{
		Token:     rtdb.Token,
		UserID:    rtdb.UserID,
		ExpiresAt: rtdb.ExpiresAt,
		CreatedAt: rtdb.CreatedAt,
		RevokedAt: rtdb.RevokedAt,
	}
}
