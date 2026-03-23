package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/mairuu/mp-api/internal/app/paging"
	"github.com/mairuu/mp-api/internal/features/history/model"
)

type Repository interface {
	Save(ctx context.Context, h *model.History) error
	SaveMany(ctx context.Context, h []model.History) error
	DeleteByChapters(ctx context.Context, userID uuid.UUID, chapterIDs []uuid.UUID) error

	ListRecent(ctx context.Context, userID uuid.UUID, p paging.Paging) (*Page[RecentReadItem], error)
	ListByManga(ctx context.Context, userID uuid.UUID, mangaID uuid.UUID, p paging.Paging) (*Page[MangaReadItem], error)
}

type Page[T any] struct {
	Items  []T
	Total  int
	Limit  int
	Offset int
}
