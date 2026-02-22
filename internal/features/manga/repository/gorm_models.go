package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/mairuu/mp-api/internal/features/manga/model"
)

// mangas

type MangaDB struct {
	ID        uuid.UUID    `gorm:"type:uuid;primaryKey"`
	OwnerID   uuid.UUID    `gorm:"type:uuid;not null;index:idx_user_id"`
	Title     string       `gorm:"type:varchar(255);not null;unique;index:idx_title_search"`
	Synopsis  string       `gorm:"type:text"`
	Status    string       `gorm:"type:varchar(10);not null;index:idx_status"`
	Covers    []CoverArtDB `gorm:"foreignKey:MangaID;constraint:OnDelete:CASCADE;"`
	CreatedAt time.Time    `gorm:"index:idx_created_at"`
	UpdatedAt time.Time
}

func (m *MangaDB) TableName() string {
	return "mangas"
}

func toMangaDB(m *model.Manga) MangaDB {
	covers := make([]CoverArtDB, 0, len(m.Covers))
	for i, c := range m.Covers {
		cdb := toCoverArtDB(&c, m.ID)
		cdb.Order = i
		covers = append(covers, cdb)
	}

	return MangaDB{
		ID:        m.ID,
		OwnerID:   m.OwnerID,
		Title:     m.Title,
		Synopsis:  m.Synopsis,
		Status:    string(m.Status),
		Covers:    covers,
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
		Covers:    covers,
		CreatedAt: mdb.CreatedAt,
		UpdatedAt: mdb.UpdatedAt,
	}
}

type CoverArtDB struct {
	MangaID     uuid.UUID `gorm:"type:uuid;primaryKey"`
	Order       int       `gorm:"type:int;primaryKey"`
	IsPrimary   bool      `gorm:"type:boolean;not null;default:false"`
	Volume      string    `gorm:"type:varchar(10);not null;primaryKey"`
	ObjectName  string    `gorm:"type:varchar(255);not null"`
	Description string    `gorm:"type:text"`
}

func (c *CoverArtDB) TableName() string {
	return "cover_arts"
}

func toCoverArtDB(cover *model.CoverArt, mangaID uuid.UUID) CoverArtDB {
	return CoverArtDB{
		MangaID:     mangaID,
		IsPrimary:   cover.IsPrimary,
		Volume:      cover.Volume,
		ObjectName:  cover.ObjectName,
		Description: cover.Description,
	}
}

func (cdb *CoverArtDB) toCoverArtModel() model.CoverArt {
	return model.CoverArt{
		IsPrimary:   cdb.IsPrimary,
		Volume:      cdb.Volume,
		ObjectName:  cdb.ObjectName,
		Description: cdb.Description,
	}
}

// chapters

type ChapterDB struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	MangaID   uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_manga_number"`
	Manga     *MangaDB  `gorm:"foreignKey:MangaID;constraint:OnDelete:CASCADE;"`
	Title     string    `gorm:"type:varchar(255)"`
	Volume    *string   `gorm:"type:varchar(10)"`
	Number    string    `gorm:"type:varchar(10);not null;uniqueIndex:idx_manga_number"`
	State     string    `gorm:"type:varchar(10);not null;index:idx_state"`
	Pages     []PageDB  `gorm:"foreignKey:ChapterID;constraint:OnDelete:CASCADE;"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (c *ChapterDB) TableName() string {
	return "chapters"
}

func toChapterDB(c *model.Chapter) ChapterDB {
	pages := make([]PageDB, 0, len(c.Pages))
	for i := range c.Pages {
		pages = append(pages, toPageDB(&c.Pages[i], c.ID, i+1))
	}
	return ChapterDB{
		ID:        c.ID,
		MangaID:   c.MangaID,
		Title:     c.Title,
		Volume:    c.Volume,
		Number:    c.Number,
		State:     string(c.State),
		Pages:     pages,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

func (cdb *ChapterDB) toChapterModel() model.Chapter {
	pages := make([]model.Page, 0, len(cdb.Pages))
	for i := range cdb.Pages {
		pages = append(pages, cdb.Pages[i].toPageModel())
	}
	return model.Chapter{
		ID:        cdb.ID,
		MangaID:   cdb.MangaID,
		Title:     cdb.Title,
		Volume:    cdb.Volume,
		Number:    cdb.Number,
		State:     model.ChapterState(cdb.State),
		Pages:     pages,
		CreatedAt: cdb.CreatedAt,
		UpdatedAt: cdb.UpdatedAt,
	}
}

type PageDB struct {
	ChapterID  uuid.UUID `gorm:"type:uuid;primaryKey"`
	Number     int       `gorm:"type:int;primaryKey"`
	Width      int       `gorm:"type:int;not null"`
	Height     int       `gorm:"type:int;not null"`
	ObjectName string    `gorm:"type:varchar(255);not null"`
}

func (p *PageDB) TableName() string {
	return "pages"
}

func toPageDB(page *model.Page, chapterID uuid.UUID, number int) PageDB {
	return PageDB{
		ChapterID:  chapterID,
		Number:     number,
		Width:      page.Width,
		Height:     page.Height,
		ObjectName: page.ObjectName,
	}
}

func (pdb *PageDB) toPageModel() model.Page {
	return model.Page{
		Width:      pdb.Width,
		Height:     pdb.Height,
		ObjectName: pdb.ObjectName,
	}
}
