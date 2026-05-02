package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type RecentReadItem struct {
	MangaID         uuid.UUID
	MangaTitle      string
	CoverObjectName *string
	ChapterID       uuid.UUID
	ChapterTitle    string
	ChapterNumber   decimal.Decimal
	Progress        float32
	ReadAt          time.Time
}

type MangaReadItem struct {
	ChapterID uuid.UUID
	Progress  float32
	ReadAt    time.Time
}
