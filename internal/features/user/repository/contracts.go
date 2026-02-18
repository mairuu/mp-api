package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/mairuu/mp-api/internal/features/user/model"
)

type Repository interface {
	SaveUser(ctx context.Context, u *model.User) error
	GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	GetUserByEmailOrUsername(ctx context.Context, emailOrUsername string) (*model.User, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
}
