package repositories

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/mairuu/mp-api/internal/features/library/model"
	libraryrepo "github.com/mairuu/mp-api/internal/features/library/repository"
	"github.com/mairuu/mp-api/internal/persistence/mappers"
	"github.com/mairuu/mp-api/internal/persistence/models"
	"github.com/mairuu/mp-api/internal/platform/collections"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type LibraryRepository struct {
	db *gorm.DB
}

// verify it implements the interface
var _ libraryrepo.Repository = (*LibraryRepository)(nil)

func NewLibraryRepository(db *gorm.DB) *LibraryRepository {
	return &LibraryRepository{db: db}
}

func (r *LibraryRepository) SaveLibrary(ctx context.Context, lib *model.Library) error {
	if lib == nil {
		return fmt.Errorf("library is nil")
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		dbs, err := gorm.G[models.LibraryMangaDB](r.db).
			Where("owner_id = ?", lib.OwnerID).
			Find(ctx)
		if err != nil {
			return fmt.Errorf("fetch existing library mangas: %w", err)
		}

		mangas := make([]models.LibraryMangaDB, len(lib.Mangas))
		for i, m := range lib.Mangas {
			mangas[i] = mappers.ToLibraryMangaDB(&m, lib.OwnerID)
		}

		differ := collections.IdentifiableDiffer[uuid.UUID, models.LibraryMangaDB]{
			GetKey: func(l *models.LibraryMangaDB) uuid.UUID {
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
					Columns:   []clause.Column{{Name: "owner_id"}, {Name: "manga_id"}},
					DoUpdates: clause.AssignmentColumns([]string{"tags"}),
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

			err = tx.
				Where("owner_id = ? AND manga_id IN ?", lib.OwnerID, mangaIDs).
				Delete(&models.LibraryMangaDB{}).Error
			if err != nil {
				return fmt.Errorf("delete library mangas: %w", err)
			}
		}

		return nil
	})
}

func (r *LibraryRepository) GetLibrary(ctx context.Context, ownerID uuid.UUID) (*model.Library, error) {
	dbs, err := gorm.G[*models.LibraryMangaDB](r.db).
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
		lib.Mangas = append(lib.Mangas, mappers.ToLibraryMangaModel(db))
	}

	return lib, nil
}

func (r *LibraryRepository) GetLibrarySummary(ctx context.Context, ownerID uuid.UUID) (*libraryrepo.LibrarySummary, error) {
	count, err := gorm.G[*models.LibraryMangaDB](r.db).
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

	return &libraryrepo.LibrarySummary{
		Tags:        tags,
		TotalMangas: int(count),
	}, nil
}
