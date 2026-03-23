package repositories

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/mairuu/mp-api/internal/app/paging"
	"github.com/mairuu/mp-api/internal/features/history/model"
	"github.com/mairuu/mp-api/internal/features/history/repository"
	"github.com/mairuu/mp-api/internal/persistence/mappers"
	"github.com/mairuu/mp-api/internal/persistence/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type HistoryRepository struct {
	db *gorm.DB
}

func NewHistoryRepository(db *gorm.DB) *HistoryRepository {
	return &HistoryRepository{db: db}
}

var _ repository.Repository = (*HistoryRepository)(nil)

func (r *HistoryRepository) Save(ctx context.Context, h *model.History) error {
	if h == nil {
		return nil
	}
	return r.SaveMany(ctx, []model.History{*h})
}

func (r *HistoryRepository) SaveMany(ctx context.Context, h []model.History) error {
	if len(h) == 0 {
		return nil
	}

	dbs := make([]models.HistoryDB, len(h))
	for i, v := range h {
		dbs[i] = mappers.ToHistoryDB(&v)
	}

	err := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}, {Name: "chapter_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"progress", "read_at"}),
		}).
		Create(&dbs).Error
	if err != nil {
		return fmt.Errorf("save many histories: %w", err)
	}

	return nil
}

func (r *HistoryRepository) DeleteByChapters(ctx context.Context, userID uuid.UUID, chapterIDs []uuid.UUID) error {
	if len(chapterIDs) == 0 {
		return nil
	}

	err := r.db.
		WithContext(ctx).
		Where("user_id = ? AND chapter_id IN ?", userID, chapterIDs).
		Delete(&models.HistoryDB{}).Error
	if err != nil {
		return fmt.Errorf("delete by chapter ids: %w", err)
	}
	return nil
}

func (r *HistoryRepository) ListRecent(
	ctx context.Context,
	userID uuid.UUID,
	p paging.Paging,
) (*repository.Page[repository.RecentReadItem], error) {
	var total int64
	err := gorm.G[models.HistoryDB](r.db).
		Raw(`
SELECT COUNT(DISTINCT c.manga_id)
FROM histories h
JOIN chapters c ON h.chapter_id = c.id
WHERE h.user_id = ?;
		`, userID).
		Scan(ctx, &total)
	if err != nil {
		return nil, fmt.Errorf("count recent history: %w", err)
	}

	rows, err := gorm.G[repository.RecentReadItem](r.db).
		Raw(`
WITH
recent_chapters AS (
	SELECT 
		c.manga_id,
		c.id as chapter_id,
		c.title as chapter_title,
		h.progress,
		h.read_at,
		ROW_NUMBER() OVER (
			PARTITION BY c.manga_id
			ORDER BY h.read_at DESC
		) AS rn
	FROM histories h
	JOIN chapters c ON h.chapter_id = c.id
	WHERE h.user_id = ?
),
best_cover AS (
	SELECT DISTINCT ON (manga_id)
		manga_id,
		object_name
	FROM cover_arts
	ORDER BY manga_id, is_primary DESC, "order" DESC
)
SELECT 
	m.id as manga_id,
	m.title as manga_title,
	bc.object_name as cover_object_name,
	rc.chapter_id,
	rc.chapter_title,
	rc.progress,
	rc.read_at
FROM recent_chapters rc
JOIN mangas m ON rc.manga_id = m.id
LEFT JOIN best_cover bc ON bc.manga_id = m.id
WHERE rc.rn = 1
ORDER BY rc.read_at DESC
LIMIT ? OFFSET ?;
		`, userID, p.Limit, p.Offset).
		Find(ctx)

	if err != nil {
		return nil, fmt.Errorf("list recent histories: %w", err)
	}

	return &repository.Page[repository.RecentReadItem]{
		Items:  rows,
		Total:  int(total),
		Limit:  p.Limit,
		Offset: p.Offset,
	}, nil
}

func (r *HistoryRepository) ListByManga(
	ctx context.Context,
	userID uuid.UUID,
	mangaID uuid.UUID,
	p paging.Paging,
) (*repository.Page[repository.MangaReadItem], error) {
	var total int64
	err := gorm.G[models.HistoryDB](r.db).
		Raw(`
SELECT COUNT(*)
FROM histories h
JOIN chapters c ON h.chapter_id = c.id
WHERE h.user_id = ? AND c.manga_id = ?
		`, userID, mangaID).
		Scan(ctx, &total)
	if err != nil {
		return nil, fmt.Errorf("count manga history: %w", err)
	}

	rows, err := gorm.G[repository.MangaReadItem](r.db).
		Raw(`
SELECT 
	h.chapter_id,
	h.progress,
	h.read_at
FROM histories h
JOIN chapters c ON h.chapter_id = c.id
WHERE h.user_id = ? AND c.manga_id = ?
ORDER BY h.read_at DESC
LIMIT ? OFFSET ?;
		`, userID, mangaID, p.Limit, p.Offset).
		Find(ctx)

	if err != nil {
		return nil, fmt.Errorf("list manga histories: %w", err)
	}

	return &repository.Page[repository.MangaReadItem]{
		Items:  rows,
		Total:  int(total),
		Limit:  p.Limit,
		Offset: p.Offset,
	}, nil
}
