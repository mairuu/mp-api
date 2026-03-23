package repository

import (
	"time"

	"github.com/google/uuid"
)

type RecentReadItem struct {
	MangaID         uuid.UUID
	MangaTitle      string
	CoverObjectName *string
	ChapterID       uuid.UUID
	ChapterTitle    string
	Progress        float32
	ReadAt          time.Time
}

type MangaReadItem struct {
	ChapterID uuid.UUID
	Progress  float32
	ReadAt    time.Time
}
