package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type LibraryMangaDB struct {
	OwnerID uuid.UUID      `gorm:"primaryKey;type:uuid"`
	Owner   *UserDB        `gorm:"foreignKey:OwnerID;constraint:OnDelete:CASCADE;"`
	MangaID uuid.UUID      `gorm:"primaryKey;type:uuid"`
	Manga   *MangaDB       `gorm:"foreignKey:MangaID;constraint:OnDelete:CASCADE;"`
	Tags    pq.StringArray `gorm:"type:text[];default:'{}'"`
	AddedAt time.Time
}

func (LibraryMangaDB) TableName() string {
	return "library_mangas"
}
