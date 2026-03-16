package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/mairuu/mp-api/internal/features/library/model"
)

// todo: enforce foreign key constraints
type LibraryMangaDB struct {
	OwnerID uuid.UUID      `gorm:"primaryKey;type:uuid"`
	MangaID uuid.UUID      `gorm:"primaryKey;type:uuid"`
	Tags    pq.StringArray `gorm:"type:text[];default:'{}'"`
	AddedAt time.Time
}

func (LibraryMangaDB) TableName() string {
	return "library_mangas"
}

func toLibraryMangaModel(db *LibraryMangaDB) model.LibraryManga {
	return model.LibraryManga{
		MangaID: db.MangaID,
		Tags:    db.Tags,
		AddedAt: db.AddedAt,
	}
}

func toLibraryMangaDB(m *model.LibraryManga, ownerID uuid.UUID) LibraryMangaDB {
	return LibraryMangaDB{
		OwnerID: ownerID,
		MangaID: m.MangaID,
		Tags:    m.Tags,
		AddedAt: m.AddedAt,
	}
}
