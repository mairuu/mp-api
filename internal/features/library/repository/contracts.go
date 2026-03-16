package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/mairuu/mp-api/internal/features/library/model"
)

type Repository interface {
	SaveLibrary(ctx context.Context, library *model.Library) error

	GetLibrary(ctx context.Context, ownerID uuid.UUID) (*model.Library, error)
	GetLibrarySummary(ctx context.Context, ownerID uuid.UUID) (*LibrarySummary, error) // todo: consider if this should be a separate query for performance reasons
}
