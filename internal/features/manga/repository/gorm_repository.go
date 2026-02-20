package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/mairuu/mp-api/internal/features/manga/model"
	"github.com/mairuu/mp-api/internal/platform/collections"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

func (r *GormRepository) SaveManga(ctx context.Context, m *model.Manga) error {
	if m == nil {
		return fmt.Errorf("manga is nil")
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		mdb := toMangaDB(m)
		err := tx.
			Clauses(clause.OnConflict{
				Columns: []clause.Column{{Name: "id"}},
				DoUpdates: clause.AssignmentColumns([]string{
					"title",
					"synopsis",
					"status",
					"state",
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
		// could've been done by delete and re-insert; but we diff just for fun

		var existingCovers []CoverArtDB
		err = tx.
			Model(&CoverArtDB{}).
			Where("manga_id = ?", m.ID).
			Find(&existingCovers).Error
		if err != nil {
			return fmt.Errorf("fetch existing cover arts: %w", err)
		}

		differ := collections.IdentifiableDiffer[string, CoverArtDB]{
			GetKey: func(ca *CoverArtDB) string {
				return ca.Volume
			},
		}

		result, err := differ.Diff(existingCovers, mdb.Covers)
		if err != nil {
			return fmt.Errorf("diff cover arts: %w", err)
		}

		toUpserts := result.Merged()
		if len(toUpserts) > 0 {
			err = tx.
				Clauses(clause.OnConflict{
					Columns: []clause.Column{{Name: "manga_id"}, {Name: "volume"}},
					DoUpdates: clause.AssignmentColumns([]string{
						"object_name",
						"description",
					}),
				}).
				Create(&toUpserts).Error
			if err != nil {
				return fmt.Errorf("upsert cover arts: %w", err)
			}
		}

		if len(result.Deleted) > 0 {
			err := tx.Delete(result.Deleted).Error
			if err != nil {
				return fmt.Errorf("delete removed cover arts: %w", err)
			}
		}

		return nil
	})
}

func (r *GormRepository) DeleteMangaByID(ctx context.Context, id uuid.UUID) error {
	affected, err := gorm.G[MangaDB](r.db).Where("id = ?", id).Delete(ctx)
	if err != nil {
		return fmt.Errorf("delete manga: %w", err)
	}
	if affected == 0 {
		return model.ErrMangaNotFound.WithArg("id", id.String())
	}
	return nil
}

func (r *GormRepository) GetMangaByID(ctx context.Context, id uuid.UUID) (*model.Manga, error) {
	mdb, err := gorm.G[MangaDB](r.db).
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

	mm := mdb.toMangaModel()
	return &mm, nil
}

func (r *GormRepository) CountMangas(ctx context.Context, filter MangaFilter) (int, error) {
	q := r.db.WithContext(ctx).
		Model(&MangaDB{})
	q = applyMangaFilter(q, filter)

	var count int64
	if err := q.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count mangas: %w", err)
	}
	return int(count), nil
}

func (r *GormRepository) ListMangas(ctx context.Context, filter MangaFilter, pagging Pagging, ordering []Ordering) ([]MangaSummary, error) {
	q := r.db.WithContext(ctx).
		Model(&MangaDB{}).
		Select("id", "title")
	q = applyMangaFilter(q, filter)
	q = applyPagging(q, pagging)
	q = applyOrderings(q, ordering)

	var mangas []MangaSummary
	if err := q.Scan(&mangas).Error; err != nil {
		return nil, fmt.Errorf("list mangas: %w", err)
	}

	if mangas == nil {
		mangas = []MangaSummary{}
	}

	return mangas, nil
}

func (r *GormRepository) SaveChapter(ctx context.Context, c *model.Chapter) error {
	if c == nil {
		return fmt.Errorf("chapter is nil")
	}

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		cdb := toChapterDB(c)
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
		err = tx.Where("chapter_id = ?", c.ID).Delete(&PageDB{}).Error
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

func (r *GormRepository) CountChapters(ctx context.Context, filter ChapterFilter) (int, error) {
	q := r.db.WithContext(ctx).
		Model(&ChapterDB{})
	q = applyChapterFilter(q, filter)

	var count int64
	if err := q.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count chapters: %w", err)
	}
	return int(count), nil
}

func (r *GormRepository) ListChapters(ctx context.Context, filter ChapterFilter, pagging Pagging, ordering []Ordering) ([]ChapterSummary, error) {
	q := r.db.WithContext(ctx).
		Model(&ChapterDB{}).
		Select("id", "manga_id", "title", "number", "volume")
	q = applyChapterFilter(q, filter)
	q = applyPagging(q, pagging)
	q = applyOrderings(q, ordering)

	var chapters []ChapterSummary
	if err := q.Scan(&chapters).Error; err != nil {
		return nil, fmt.Errorf("list chapters: %w", err)
	}

	if chapters == nil {
		chapters = []ChapterSummary{}
	}

	return chapters, nil
}

func (r *GormRepository) GetChapterByID(ctx context.Context, id uuid.UUID) (*model.Chapter, error) {
	cdb, err := gorm.G[ChapterDB](r.db).
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

	cm := cdb.toChapterModel()
	return &cm, nil
}

func (r *GormRepository) DeleteChapterByID(ctx context.Context, id uuid.UUID) error {
	affected, err := gorm.G[ChapterDB](r.db).Where("id = ?", id).Delete(ctx)
	if err != nil {
		return fmt.Errorf("delete chapter: %w", err)
	}
	if affected == 0 {
		return model.ErrChapterNotFound.WithArg("id", id.String())
	}
	return nil
}

func applyMangaFilter(q *gorm.DB, filter MangaFilter) *gorm.DB {
	if len(filter.IDs) > 0 {
		if len(filter.IDs) == 1 {
			q = q.Where("id = ?", filter.IDs[0])
		} else {
			q = q.Where("id IN ?", filter.IDs)
		}
	}
	if filter.Title != nil {
		q = q.Where("title ILIKE ?", "%"+*filter.Title+"%")
	}
	if filter.Status != nil {
		q = q.Where("status = ?", *filter.Status)
	}
	if filter.State != nil {
		q = q.Where("state = ?", *filter.State)
	}
	return q
}

func applyChapterFilter(q *gorm.DB, filter ChapterFilter) *gorm.DB {
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

func applyPagging(q *gorm.DB, pagging Pagging) *gorm.DB {
	q = q.Limit(pagging.Limit).Offset(pagging.Offset)
	return q
}

func applyOrderings(q *gorm.DB, ordering []Ordering) *gorm.DB {
	for _, o := range ordering {
		if o.Field == "" && !o.Direction.IsValid() {
			continue
		}
		q = q.Order(clause.OrderByColumn{
			Column: clause.Column{Name: o.Field},
			Desc:   o.Direction == Desc,
		})
	}

	return q
}
