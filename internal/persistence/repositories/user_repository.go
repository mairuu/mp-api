package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mairuu/mp-api/internal/features/user/model"
	userrepo "github.com/mairuu/mp-api/internal/features/user/repository"
	"github.com/mairuu/mp-api/internal/persistence/mappers"
	"github.com/mairuu/mp-api/internal/persistence/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UserRepository struct {
	db *gorm.DB
}

// verify it implements the interface
var _ userrepo.Repository = (*UserRepository)(nil)

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) SaveUser(ctx context.Context, u *model.User) error {
	if u == nil {
		return fmt.Errorf("user is nil")
	}

	udb := mappers.ToUserDB(u)
	err := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"username",
				"email",
				"password_hash",
				"role",
				"updated_at",
			}),
		}).
		Create(&udb).Error

	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return model.ErrUserAlreadyExists.
				WithArg("username", u.Username).
				WithArg("email", u.Email)
		}
		return fmt.Errorf("upsert user: %w", err)
	}

	return nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	udb, err := gorm.G[models.UserDB](r.db).Where("id = ?", id).First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, model.ErrUserNotFound.WithArg("id", id.String())
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	u := mappers.UserDBToModel(&udb)
	return &u, nil
}

func (r *UserRepository) GetUserByEmailOrUsername(ctx context.Context, emailOrUsername string) (*model.User, error) {
	udb, err := gorm.G[models.UserDB](r.db).Where("email = ? OR username = ?", emailOrUsername, emailOrUsername).First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, model.ErrUserNotFound.WithArg("email_or_username", emailOrUsername)
		}
		return nil, fmt.Errorf("get user by email or username: %w", err)
	}
	u := mappers.UserDBToModel(&udb)
	return &u, nil
}

func (r *UserRepository) DeleteUser(ctx context.Context, id uuid.UUID) error {
	affected, err := gorm.G[models.UserDB](r.db).Where("id = ?", id).Delete(ctx)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	if affected == 0 {
		return model.ErrUserNotFound.WithArg("id", id.String())
	}
	return nil
}

func (r *UserRepository) SaveRefreshToken(ctx context.Context, rt *model.RefreshToken) error {
	rtdb := mappers.ToRefreshTokenDB(rt)
	if err := r.db.WithContext(ctx).Create(&rtdb).Error; err != nil {
		return fmt.Errorf("save refresh token: %w", err)
	}
	return nil
}

func (r *UserRepository) GetRefreshToken(ctx context.Context, token string) (*model.RefreshToken, error) {
	rtdb, err := gorm.G[models.RefreshTokenDB](r.db).Where("token = ?", token).First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, model.ErrRefreshTokenNotFound
		}
		return nil, fmt.Errorf("get refresh token: %w", err)
	}
	rt := mappers.RefreshTokenDBToModel(&rtdb)
	return &rt, nil
}

func (r *UserRepository) RevokeRefreshToken(ctx context.Context, token string) error {
	now := time.Now()
	affected, err := gorm.G[models.RefreshTokenDB](r.db).
		Where("token = ? AND revoked_at IS NULL", token).
		Update(ctx, "revoked_at", now)
	if err != nil {
		return fmt.Errorf("revoke refresh token: %w", err)
	}
	if affected == 0 {
		return model.ErrRefreshTokenNotFound
	}
	return nil
}

func (r *UserRepository) RevokeAllUserRefreshTokens(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()
	_, err := gorm.G[models.RefreshTokenDB](r.db).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update(ctx, "revoked_at", now)
	if err != nil {
		return fmt.Errorf("revoke all user refresh tokens: %w", err)
	}
	return nil
}

func (r *UserRepository) DeleteExpiredRefreshTokens(ctx context.Context) error {
	_, err := gorm.G[models.RefreshTokenDB](r.db).
		Where("expires_at < ?", time.Now()).
		Delete(ctx)
	if err != nil {
		return fmt.Errorf("delete expired refresh tokens: %w", err)
	}
	return nil
}
