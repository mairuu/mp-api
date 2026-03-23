package service

import (
	"time"

	"github.com/mairuu/mp-api/internal/features/history/repository"
)

type mapper struct{}

func (m *mapper) ToRecentReadDTO(h *repository.RecentReadItem) RecentReadDTO {
	return RecentReadDTO{
		MangaID:         h.MangaID.String(),
		MangaName:       h.MangaTitle,
		CoverObjectName: h.CoverObjectName,
		ChapterID:       h.ChapterID.String(),
		ChapterName:     h.ChapterTitle,
		ReadAt:          h.ReadAt.Format(time.RFC3339),
	}
}

func (m *mapper) ToMangaReadDTO(h *repository.MangaReadItem) MangaReadDTO {
	return MangaReadDTO{
		ChapterID: h.ChapterID.String(),
		Progress:  h.Progress,
		ReadAt:    h.ReadAt.Format(time.RFC3339),
	}
}
