package models

import (
	"time"

	"github.com/google/uuid"
)

// todo: add indexes

type HistoryDB struct {
	UserID    uuid.UUID `gorm:"primaryKey;type:uuid"`
	User      UserDB    `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	ChapterID uuid.UUID `gorm:"primaryKey;type:uuid"`
	Chapter   ChapterDB `gorm:"foreignKey:ChapterID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	ReadAt    time.Time `gorm:"index_read_at"`
}

func (HistoryDB) TableName() string {
	return "histories"
}
