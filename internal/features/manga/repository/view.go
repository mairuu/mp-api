package repository

import (
	"time"

	"github.com/google/uuid"
)

type MangaSummary struct {
	ID               uuid.UUID
	Title            string
	CoverVolume      *string
	CoverObjecrtName *string
}

type ChapterSummary struct {
	ID        uuid.UUID
	MangaID   uuid.UUID
	Number    string
	Title     *string
	Volume    *string
	CreatedAt time.Time
}
