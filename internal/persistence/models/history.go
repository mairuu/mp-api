package models

import (
	"time"

	"github.com/google/uuid"
)

type HistoryDB struct {
	UserID    uuid.UUID `gorm:"primaryKey;type:uuid;index:idx_user_chapter_read"`
	User      UserDB    `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	ChapterID uuid.UUID `gorm:"primaryKey;type:uuid;index:idx_user_chapter_read"`
	Chapter   ChapterDB `gorm:"foreignKey:ChapterID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Progress  float32   `gorm:"type:float;not null;default:0"`
	ReadAt    time.Time `gorm:"index:idx_user_chapter_read,sort:desc"`
}

func (HistoryDB) TableName() string {
	return "histories"
}
