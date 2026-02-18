package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/mairuu/mp-api/internal/features/manga/model"
)

type MangaDB struct {
	ID        uuid.UUID    `gorm:"type:uuid;primaryKey"`
	OwnerID   uuid.UUID    `gorm:"type:uuid;not null;index:idx_user_id"`
	Title     string       `gorm:"type:varchar(255);not null;unique;index:idx_title_search"`
	Synopsis  string       `gorm:"type:text"`
	Status    string       `gorm:"type:varchar(10);not null;index:idx_status"`
	State     string       `gorm:"type:varchar(10);not null;index:idx_state"`
	Covers    []CoverArtDB `gorm:"foreignKey:MangaID;constraint:OnDelete:CASCADE;"`
	CreatedAt time.Time    `gorm:"index:idx_created_at"`
	UpdatedAt time.Time
}

func (m *MangaDB) TableName() string {
	return "mangas"
}

func toMangaDB(m *model.Manga) MangaDB {
	return MangaDB{
		ID:        m.ID,
		OwnerID:   m.OwnerID,
		Title:     m.Title,
		Synopsis:  m.Synopsis,
		Status:    string(m.Status),
		State:     string(m.State),
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func (mdb *MangaDB) toMangaModel() model.Manga {
	covers := make([]model.CoverArt, 0, len(mdb.Covers))
	for _, cdb := range mdb.Covers {
		covers = append(covers, cdb.toCoverArtModel())
	}

	return model.Manga{
		ID:        mdb.ID,
		OwnerID:   mdb.OwnerID,
		Title:     mdb.Title,
		Synopsis:  mdb.Synopsis,
		Status:    model.MangaStatus(mdb.Status),
		State:     model.MangaState(mdb.State),
		Covers:    covers,
		CreatedAt: mdb.CreatedAt,
		UpdatedAt: mdb.UpdatedAt,
	}
}

type CoverArtDB struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	MangaID     uuid.UUID `gorm:"type:uuid;not null;index:idx_manga_id_volume"`
	Volume      string    `gorm:"type:varchar(10);not null;index:idx_manga_id_volume"`
	ObjectName  string    `gorm:"type:varchar(255);not null"`
	Description string    `gorm:"type:text"`
}

func (c *CoverArtDB) TableName() string {
	return "cover_arts"
}

func toCoverArtDB(cover *model.CoverArt, mangaID uuid.UUID) CoverArtDB {
	return CoverArtDB{
		ID:          cover.ID,
		MangaID:     mangaID,
		Volume:      cover.Volume,
		ObjectName:  cover.ObjectName,
		Description: cover.Description,
	}
}

func (cdb *CoverArtDB) toCoverArtModel() model.CoverArt {
	return model.CoverArt{
		ID:          cdb.ID,
		Volume:      cdb.Volume,
		ObjectName:  cdb.ObjectName,
		Description: cdb.Description,
	}
}
