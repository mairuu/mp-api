package mappers

import (
	"github.com/mairuu/mp-api/internal/features/history/model"
	"github.com/mairuu/mp-api/internal/persistence/models"
)

func ToHistoryDB(m *model.History) models.HistoryDB {
	return models.HistoryDB{
		UserID:    m.UserID,
		ChapterID: m.ChapterID,
		ReadAt:    m.ReadAt,
	}
}
