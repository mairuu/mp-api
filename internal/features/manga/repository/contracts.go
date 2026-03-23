package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/mairuu/mp-api/internal/app/ordering"
	"github.com/mairuu/mp-api/internal/app/paging"
	model "github.com/mairuu/mp-api/internal/features/manga/model"
)

type Repository interface {
	SaveManga(ctx context.Context, m *model.Manga) error
	DeleteMangaByID(ctx context.Context, id uuid.UUID) error

	GetMangaByID(ctx context.Context, id uuid.UUID) (*model.Manga, error)
	ListMangas(
		ctx context.Context,
		filter MangaFilter,
		paging paging.Paging,
		ordering []ordering.Ordering,
	) (*Page[MangaSummary], error)

	SaveChapter(ctx context.Context, c *model.Chapter) error
	DeleteChapterByID(ctx context.Context, id uuid.UUID) error

	GetChapterByID(ctx context.Context, id uuid.UUID) (*model.Chapter, error)
	ListChapters(
		ctx context.Context,
		filter ChapterFilter,
		paging paging.Paging,
		ordering []ordering.Ordering,
	) (*Page[ChapterSummary], error)
}

type Page[T any] struct {
	Items  []T
	Total  int
	Limit  int
	Offset int
}

type MangaFilter struct {
	IDs      []string
	OwnerIDs []string
	Title    *string
	Status   *string
}

type ChapterFilter struct {
	IDs      []string
	MangaIDs []string
	Title    *string
	Number   *string
	Volume   *string
	State    *string
}

const (
	// shared
	OrderByTitle     ordering.Field = "title"
	OrderByCreatedAt ordering.Field = "created_at"
	OrderByUpdatedAt ordering.Field = "updated_at"

	// chapter-specific
	OrderByChapterNumber ordering.Field = "number"
	OrderByChapterVolume ordering.Field = "volume"
)
