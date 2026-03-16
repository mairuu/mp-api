package service

import (
	"github.com/google/uuid"
	"github.com/mairuu/mp-api/internal/features/library/model"
)

type LibraryDTO struct {
	Tags   []string          `json:"tags"`
	Mangas []LibraryMangaDTO `json:"mangas"`
}

type LibraryMangaDTO struct {
	MangaID string   `json:"manga_id"`
	Tags    []string `json:"tags"`
	AddedAt string   `json:"added_at"`
}

type LibrarySummaryDTO struct {
	Tags        []string `json:"tags"`
	TotalMangas int      `json:"total_mangas"`
}

type UpsertLibraryMangaDTO struct {
	MangaID string   `json:"manga_id"`
	Tags    []string `json:"tags"`
}

func (dto *UpsertLibraryMangaDTO) ToModel() (*model.LibraryManga, error) {
	id, err := uuid.Parse(dto.MangaID)
	if err != nil {
		return nil, model.ErrInvalidLibraryManga.WithArg("id", dto.MangaID)
	}
	return &model.LibraryManga{
		MangaID: id,
		Tags:    dto.Tags,
	}, nil
}
