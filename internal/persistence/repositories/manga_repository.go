package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/mairuu/mp-api/internal/app/ordering"
	"github.com/mairuu/mp-api/internal/app/paging"
	"github.com/mairuu/mp-api/internal/features/manga/model"
	mangarepo "github.com/mairuu/mp-api/internal/features/manga/repository"
	"github.com/mairuu/mp-api/internal/persistence/mappers"
	"github.com/mairuu/mp-api/internal/persistence/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type MangaRepository struct {
	db *gorm.DB
}

// verify it implements the interface
var _ mangarepo.Repository = (*MangaRepository)(nil)

func NewMangaRepository(db *gorm.DB) *MangaRepository {
	return &MangaRepository{db: db}
}

func (r *MangaRepository) SaveManga(ctx context.Context, m *model.Manga) error {
	if m == nil {
		return fmt.Errorf("manga is nil")
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		mdb := mappers.ToMangaDB(m)
		err := tx.
			Clauses(clause.OnConflict{
				Columns: []clause.Column{{Name: "id"}},
				DoUpdates: clause.AssignmentColumns([]string{
					"title",
					"synopsis",
					"status",
					"updated_at",
				}),
			}).
			Create(&mdb).Error

		if err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return model.ErrMangaAlreadyExists.WithArg("id", m.ID.String())
			}
			return fmt.Errorf("upsert manga: %w", err)
		}

		// sync cover arts
		// delete and re-insert
		err = tx.Where("manga_id = ?", m.ID).Delete(&models.CoverArtDB{}).Error
		if err != nil {
			return fmt.Errorf("delete existing cover arts: %w", err)
		}

		if len(mdb.Covers) > 0 {
			err = tx.CreateInBatches(&mdb.Covers, 100).Error
			if err != nil {
				return fmt.Errorf("insert cover arts: %w", err)
			}
		}

		return nil
	})
}

func (r *MangaRepository) DeleteMangaByID(ctx context.Context, id uuid.UUID) error {
	affected, err := gorm.G[models.MangaDB](r.db).Where("id = ?", id).Delete(ctx)
	if err != nil {
		return fmt.Errorf("delete manga: %w", err)
	}
	if affected == 0 {
		return model.ErrMangaNotFound.WithArg("id", id.String())
	}
	return nil
}

func (r *MangaRepository) GetMangaByID(ctx context.Context, id uuid.UUID) (*model.Manga, error) {
	mdb, err := gorm.G[models.MangaDB](r.db).
		Preload("Covers", func(db gorm.PreloadBuilder) error {
			db.Order("volume")
			return nil
		}).
		Where("id = ?", id).
		First(ctx)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, model.ErrMangaNotFound.WithArg("id", id.String())
		}
		return nil, fmt.Errorf("get manga by id: %w", err)
	}

	mm := mappers.MangaDBToModel(&mdb)
	return &mm, nil
}

func (r *MangaRepository) countMangas(ctx context.Context, filter mangarepo.MangaFilter) (int, error) {
	q := r.db.WithContext(ctx).
		Model(&models.MangaDB{})
	q = applyMangaFilter(q, filter)

	var count int64
	if err := q.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count mangas: %w", err)
	}
	return int(count), nil
}

func (r *MangaRepository) ListMangas(
	ctx context.Context,
	filter mangarepo.MangaFilter,
	paging paging.Paging,
	ordering []ordering.Ordering,
) (*mangarepo.Page[mangarepo.MangaSummary], error) {
	total, err := r.countMangas(ctx, filter)
	if err != nil {
		return nil, err
	}

	ms := make([]struct {
		ID    uuid.UUID
		Title string
	}, 0)

	q := r.db.WithContext(ctx).
		Model(&models.MangaDB{}).
		Select("id", "title")
	q = applyMangaFilter(q, filter)
	q = applyPagging(q, paging)
	q = applyOrderings(q, ordering)
	if err := q.Scan(&ms).Error; err != nil {
		return nil, fmt.Errorf("list mangas: %w", err)
	}

	ids := make([]uuid.UUID, 0, len(ms))
	for i := range ms {
		ids = append(ids, ms[i].ID)
	}

	covers, err := gorm.G[models.CoverArtDB](r.db).
		Select("DISTINCT ON (manga_id) manga_id, object_name, volume").
		Where("manga_id in ?", ids).
		Order(`manga_id, is_primary DESC, "order" DESC`).
		Find(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch cover arts for manga summaries: %w", err)
	}

	coverMap := make(map[uuid.UUID]*models.CoverArtDB)
	for i := range covers {
		c := &covers[i]
		if _, exists := coverMap[c.MangaID]; !exists {
			coverMap[c.MangaID] = c
		}
	}

	mangas := make([]mangarepo.MangaSummary, 0, len(ms))
	for i, m := range ms {
		mangas = append(mangas, mangarepo.MangaSummary{
			ID:    m.ID,
			Title: m.Title,
		})
		if cover, exists := coverMap[m.ID]; exists {
			mangas[i].CoverVolume = cover.Volume
			mangas[i].CoverObjectName = &cover.ObjectName
		}
	}

	return &mangarepo.Page[mangarepo.MangaSummary]{
		Items:  mangas,
		Total:  total,
		Limit:  paging.Limit,
		Offset: paging.Offset,
	}, nil
}

func (r *MangaRepository) SaveChapter(ctx context.Context, c *model.Chapter) error {
	if c == nil {
		return fmt.Errorf("chapter is nil")
	}

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		cdb := mappers.ToChapterDB(c)
		err := tx.
			Clauses(clause.OnConflict{
				Columns: []clause.Column{{Name: "id"}},
				DoUpdates: clause.AssignmentColumns([]string{
					"title",
					"volume",
					"number",
					"state",
					"updated_at",
				}),
			}).
			Create(&cdb).Error
		if err != nil {
			if errors.Is(err, gorm.ErrForeignKeyViolated) {
				return model.ErrMangaNotFound.WithArg("manga_id", c.MangaID.String())
			}
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return model.ErrChapterAlreadyExists.
					WithArg("manga_id", c.MangaID.String()).
					WithArg("number", c.Number)
			}
			return fmt.Errorf("upsert chapter: %w", err)
		}

		// sync pages
		// delete and re-insert
		err = tx.Where("chapter_id = ?", c.ID).Delete(&models.ChapterPageDB{}).Error
		if err != nil {
			return fmt.Errorf("delete existing pages: %w", err)
		}

		if len(cdb.Pages) > 0 {
			err = tx.CreateInBatches(&cdb.Pages, 100).Error
			if err != nil {
				return fmt.Errorf("insert pages: %w", err)
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *MangaRepository) countChapters(ctx context.Context, filter mangarepo.ChapterFilter) (int, error) {
	q := r.db.WithContext(ctx).
		Model(&models.ChapterDB{})
	q = applyChapterFilter(q, filter)

	var count int64
	if err := q.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count chapters: %w", err)
	}
	return int(count), nil
}

func (r *MangaRepository) ListChapters(
	ctx context.Context,
	filter mangarepo.ChapterFilter,
	paging paging.Paging,
	ordering []ordering.Ordering,
) (*mangarepo.Page[mangarepo.ChapterSummary], error) {
	total, err := r.countChapters(ctx, filter)
	if err != nil {
		return nil, err
	}

	q := r.db.WithContext(ctx).
		Model(&models.ChapterDB{}).
		Select("id", "manga_id", "title", "number", "volume", "created_at")
	q = applyChapterFilter(q, filter)
	q = applyPagging(q, paging)
	q = applyOrderings(q, ordering)

	var chapters []mangarepo.ChapterSummary
	if err := q.Scan(&chapters).Error; err != nil {
		return nil, fmt.Errorf("list chapters: %w", err)
	}

	if chapters == nil {
		chapters = []mangarepo.ChapterSummary{}
	}

	return &mangarepo.Page[mangarepo.ChapterSummary]{
		Items:  chapters,
		Total:  total,
		Limit:  paging.Limit,
		Offset: paging.Offset,
	}, nil
}

func (r *MangaRepository) GetChapterByID(ctx context.Context, id uuid.UUID) (*model.Chapter, error) {
	cdb, err := gorm.G[models.ChapterDB](r.db).
		Where("id = ?", id).
		Preload("Pages", func(db gorm.PreloadBuilder) error {
			db.Order("number")
			return nil
		}).
		First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, model.ErrChapterNotFound.WithArg("id", id.String())
		}
		return nil, fmt.Errorf("get chapter by id: %w", err)
	}

	cm := mappers.ChapterDBToModel(&cdb)
	return &cm, nil
}

func (r *MangaRepository) DeleteChapterByID(ctx context.Context, id uuid.UUID) error {
	affected, err := gorm.G[models.ChapterDB](r.db).Where("id = ?", id).Delete(ctx)
	if err != nil {
		return fmt.Errorf("delete chapter: %w", err)
	}
	if affected == 0 {
		return model.ErrChapterNotFound.WithArg("id", id.String())
	}
	return nil
}

func applyMangaFilter(q *gorm.DB, filter mangarepo.MangaFilter) *gorm.DB {
	if len(filter.IDs) > 0 {
		if len(filter.IDs) == 1 {
			q = q.Where("id = ?", filter.IDs[0])
		} else {
			q = q.Where("id IN ?", filter.IDs)
		}
	}
	if len(filter.OwnerIDs) > 0 {
		if len(filter.OwnerIDs) == 1 {
			q = q.Where("owner_id = ?", filter.OwnerIDs[0])
		} else {
			q = q.Where("owner_id IN ?", filter.OwnerIDs)
		}
	}
	if filter.Title != nil {
		q = q.Where("title ILIKE ?", "%"+*filter.Title+"%")
	}
	if filter.Status != nil {
		q = q.Where("status = ?", *filter.Status)
	}
	return q
}

func applyChapterFilter(q *gorm.DB, filter mangarepo.ChapterFilter) *gorm.DB {
	if len(filter.IDs) > 0 {
		if len(filter.IDs) == 1 {
			q = q.Where("id = ?", filter.IDs[0])
		} else {
			q = q.Where("id IN ?", filter.IDs)
		}
	}
	if len(filter.MangaIDs) > 0 {
		if len(filter.MangaIDs) == 1 {
			q = q.Where("manga_id = ?", filter.MangaIDs[0])
		} else {
			q = q.Where("manga_id IN ?", filter.MangaIDs)
		}
	}
	if filter.Title != nil {
		q = q.Where("title ILIKE ?", "%"+*filter.Title+"%")
	}
	if filter.Number != nil {
		q = q.Where("number = ?", *filter.Number)
	}
	if filter.Volume != nil {
		q = q.Where("volume = ?", *filter.Volume)
	}
	if filter.State != nil {
		q = q.Where("state = ?", *filter.State)
	}
	return q
}
