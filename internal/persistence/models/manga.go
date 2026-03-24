package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type MangaDB struct {
	ID        uuid.UUID    `gorm:"type:uuid;primaryKey"`
	OwnerID   uuid.UUID    `gorm:"type:uuid;not null;index:idx_user_id"`
	Owner     *UserDB      `gorm:"foreignKey:OwnerID;constraint:OnDelete:CASCADE;"`
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

type CoverArtDB struct {
	MangaID     uuid.UUID        `gorm:"type:uuid;primaryKey;index:idx_manga_primary_order;index:idx_manga_volume"`
	Manga       *MangaDB         `gorm:"foreignKey:MangaID;constraint:OnDelete:CASCADE;"`
	Order       int              `gorm:"type:int;primaryKey;index:idx_manga_primary_order,sort:desc"`
	ObjectName  string           `gorm:"type:varchar(255);not null"`
	IsPrimary   bool             `gorm:"type:boolean;not null;default:false;index:idx_manga_primary_order,sort:desc"`
	Volume      *decimal.Decimal `gorm:"type:decimal(10,4);index:idx_manga_volume"`
	Description *string          `gorm:"type:text"`
}

func (c *CoverArtDB) TableName() string {
	return "cover_arts"
}

type ChapterDB struct {
	ID        uuid.UUID        `gorm:"type:uuid;primaryKey"`
	MangaID   uuid.UUID        `gorm:"type:uuid;not null;uniqueIndex:idx_manga_number"`
	Manga     *MangaDB         `gorm:"foreignKey:MangaID;constraint:OnDelete:CASCADE;"`
	Title     *string          `gorm:"type:varchar(255)"`
	Volume    *decimal.Decimal `gorm:"type:decimal(10, 4)"`
	Number    decimal.Decimal  `gorm:"type:decimal(10, 4);not null;uniqueIndex:idx_manga_number"`
	State     string           `gorm:"type:varchar(10);not null;index:idx_state"`
	Pages     []ChapterPageDB  `gorm:"foreignKey:ChapterID;constraint:OnDelete:CASCADE;"`
	CreatedAt time.Time        `gorm:"index:idx_created_at"`
	UpdatedAt time.Time
}

func (c *ChapterDB) TableName() string {
	return "chapters"
}

type ChapterPageDB struct {
	ChapterID  uuid.UUID `gorm:"type:uuid;primaryKey"`
	Number     int       `gorm:"type:int;primaryKey"`
	Width      int       `gorm:"type:int;not null"`
	Height     int       `gorm:"type:int;not null"`
	ObjectName string    `gorm:"type:varchar(255);not null"`
}

func (p *ChapterPageDB) TableName() string {
	return "chapter_pages"
}
