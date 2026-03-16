package mappers

import (
	"github.com/google/uuid"
	"github.com/mairuu/mp-api/internal/features/library/model"
	"github.com/mairuu/mp-api/internal/persistence/models"
)

func ToLibraryMangaModel(db *models.LibraryMangaDB) model.LibraryManga {
	return model.LibraryManga{
		MangaID: db.MangaID,
		Tags:    db.Tags,
		AddedAt: db.AddedAt,
	}
}

func ToLibraryMangaDB(m *model.LibraryManga, ownerID uuid.UUID) models.LibraryMangaDB {
	return models.LibraryMangaDB{
		OwnerID: ownerID,
		MangaID: m.MangaID,
		Tags:    m.Tags,
		AddedAt: m.AddedAt,
	}
}
