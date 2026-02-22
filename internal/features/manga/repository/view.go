package repository

import "github.com/google/uuid"

type MangaSummary struct {
	ID               uuid.UUID
	Title            string
	CoverVolume      *string
	CoverObjecrtName *string
}

type ChapterSummary struct {
	ID      uuid.UUID
	MangaID uuid.UUID
	Title   string
	Number  string
	Volume  *string
}
