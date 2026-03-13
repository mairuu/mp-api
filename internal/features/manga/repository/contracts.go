package repository

import (
	"context"

	"github.com/google/uuid"
	model "github.com/mairuu/mp-api/internal/features/manga/model"
)

type Repository interface {
	SaveManga(ctx context.Context, m *model.Manga) error
	DeleteMangaByID(ctx context.Context, id uuid.UUID) error

	GetMangaByID(ctx context.Context, id uuid.UUID) (*model.Manga, error)
	CountMangas(ctx context.Context, filter MangaFilter) (int, error)
	ListMangas(ctx context.Context, filter MangaFilter, paging Pagging, ordering []Ordering) ([]MangaSummary, error)

	SaveChapter(ctx context.Context, c *model.Chapter) error
	DeleteChapterByID(ctx context.Context, id uuid.UUID) error

	GetChapterByID(ctx context.Context, id uuid.UUID) (*model.Chapter, error)
	CountChapters(ctx context.Context, filter ChapterFilter) (int, error)
	ListChapters(ctx context.Context, filter ChapterFilter, paging Pagging, ordering []Ordering) ([]ChapterSummary, error)
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

type Pagging struct {
	Limit  int
	Offset int
}

type OrderingField string

const (
	// shared
	OrderByTitle     OrderingField = "title"
	OrderByCreatedAt OrderingField = "created_at"
	OrderByUpdatedAt OrderingField = "updated_at"

	// chapter-specific
	OrderByChapterNumber OrderingField = "number"
	OrderByChapterVolume OrderingField = "volume"
)

type Ordering struct {
	Field     OrderingField
	Direction OrderingDirection
}

type OrderingDirection string

const (
	Asc  OrderingDirection = "ASC"
	Desc OrderingDirection = "DESC"
)

func (d OrderingDirection) IsValid() bool {
	return d == Asc || d == Desc
}
