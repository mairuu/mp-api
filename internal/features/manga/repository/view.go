package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type MangaSummary struct {
	ID              uuid.UUID
	Title           string
	CoverVolume     *decimal.Decimal
	CoverObjectName *string
}

type ChapterSummary struct {
	ID        uuid.UUID
	MangaID   uuid.UUID
	Number    decimal.Decimal
	Title     *string
	Volume    *decimal.Decimal
	CreatedAt time.Time
}
