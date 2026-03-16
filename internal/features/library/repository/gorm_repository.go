package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/mairuu/mp-api/internal/features/library/model"
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

func (r *GormRepository) SaveLibrary(ctx context.Context, lib *model.Library) error {
	if lib == nil {
		return fmt.Errorf("library is nil")
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		dbs, err := gorm.G[LibraryMangaDB](r.db).
			Where("owner_id = ?", lib.OwnerID).
			Find(ctx)
		if err != nil {
			return fmt.Errorf("fetch existing library mangas: %w", err)
		}

		mangas := make([]LibraryMangaDB, len(lib.Mangas))
		for i, m := range lib.Mangas {
			mangas[i] = toLibraryMangaDB(&m, lib.OwnerID)
		}

		differ := collections.IdentifiableDiffer[uuid.UUID, LibraryMangaDB]{
			GetKey: func(l *LibraryMangaDB) uuid.UUID {
				return l.MangaID
			},
		}

		dr, err := differ.Diff(dbs, mangas)
		if err != nil {
			return fmt.Errorf("diff library mangas: %w", err)
		}

		// todo: only upsert changed
		toUpserts := dr.Merged()
		toDeletes := dr.Deleted

		if len(toUpserts) > 0 {
			err = tx.
				Clauses(clause.OnConflict{
					Columns: []clause.Column{{Name: "owner_id"}, {Name: "manga_id"}},
					DoUpdates: clause.AssignmentColumns([]string{
						"tags",
					}),
				}).
				CreateInBatches(&toUpserts, 100).Error
			if err != nil {
				return fmt.Errorf("upsert library mangas: %w", err)
			}
		}

		if len(toDeletes) > 0 {
			var mangaIDs []uuid.UUID
			for _, m := range toDeletes {
				mangaIDs = append(mangaIDs, m.MangaID)
			}

			err = tx.Where("owner_id = ? AND manga_id IN ?", lib.OwnerID, mangaIDs).Delete(&LibraryMangaDB{}).Error
			if err != nil {
				return fmt.Errorf("delete library mangas: %w", err)
			}
		}

		return nil
	})
}

func (r *GormRepository) GetLibrary(ctx context.Context, ownerID uuid.UUID) (*model.Library, error) {
	dbs, err := gorm.G[*LibraryMangaDB](r.db).
		Where("owner_id = ?", ownerID).
		Find(ctx)
	if err != nil {
		return nil, fmt.Errorf("get library: %w", err)
	}

	lib := &model.Library{
		OwnerID: ownerID,
		Mangas:  make([]model.LibraryManga, 0, len(dbs)),
	}
	for _, db := range dbs {
		lib.Mangas = append(lib.Mangas, toLibraryMangaModel(db))
	}

	return lib, nil
}

func (r *GormRepository) GetLibrarySummary(ctx context.Context, ownerID uuid.UUID) (*LibrarySummary, error) {
	count, err := gorm.G[*LibraryMangaDB](r.db).
		Where("owner_id = ?", ownerID).
		Count(ctx, "*")
	if err != nil {
		return nil, fmt.Errorf("get library summary: %w", err)
	}

	tags, err := gorm.G[string](r.db).
		Raw("SELECT DISTINCT unnest(tags) AS tag FROM library_mangas WHERE owner_id = ?", ownerID).
		Find(ctx)
	if err != nil {
		return nil, fmt.Errorf("get library summary tags: %w", err)
	}

	// wip:
	// in case ^above is not working
	// type tagRow struct{ Tag string }
	// rows, err := gorm.G[tagRow](r.db).Raw(...).Find(ctx)
	// tags := make([]string, len(rows))
	// for i, r := range rows { tags[i] = r.Tag }

	return &LibrarySummary{
		Tags:        tags,
		TotalMangas: int(count),
	}, nil
}
