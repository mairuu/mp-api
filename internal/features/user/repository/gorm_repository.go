package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/mairuu/mp-api/internal/features/user/model"
)

type GormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

func (r *GormRepository) SaveUser(ctx context.Context, u *model.User) error {
	if u == nil {
		return fmt.Errorf("user is nil")
	}

	udb := toUserDB(u)
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

func (r *GormRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	udb, err := gorm.G[UserDB](r.db).Where("id = ?", id).First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, model.ErrUserNotFound.WithArg("id", id.String())
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	u := toUserModel(&udb)
	return &u, nil
}

func (r *GormRepository) GetUserByEmailOrUsername(ctx context.Context, emailOrUsername string) (*model.User, error) {
	udb, err := gorm.G[UserDB](r.db).Where("email = ? OR username = ?", emailOrUsername, emailOrUsername).First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, model.ErrUserNotFound.WithArg("email_or_username", emailOrUsername)
		}
		return nil, fmt.Errorf("get user by email or username: %w", err)
	}
	u := toUserModel(&udb)
	return &u, nil
}

func (r *GormRepository) DeleteUser(ctx context.Context, id uuid.UUID) error {
	affected, err := gorm.G[UserDB](r.db).Where("id = ?", id).Delete(ctx)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	if affected == 0 {
		return model.ErrUserNotFound.WithArg("id", id.String())
	}
	return nil
}
