package model

import (
	"time"

	"github.com/google/uuid"
)

type History struct {
	UserID    uuid.UUID
	ChapterID uuid.UUID
	Progress  float32 // 0.0 to 1.0
	ReadAt    time.Time
}

func NewHistory(userID, chapterID uuid.UUID, progress float32) History {
	if progress < 0.0 {
		progress = 0.0
	} else if progress > 1.0 {
		progress = 1.0
	}
	return History{
		UserID:    userID,
		ChapterID: chapterID,
		Progress:  progress,
		ReadAt:    time.Now(),
	}
}
