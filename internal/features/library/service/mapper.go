package service

import (
	"time"

	"github.com/mairuu/mp-api/internal/features/library/model"
	"github.com/mairuu/mp-api/internal/features/library/repository"
)

type mapper struct{}

func (m *mapper) ToLibraryDTO(lib *model.Library) *LibraryDTO {
	dto := &LibraryDTO{
		Tags:   lib.AllTags(),
		Mangas: make([]LibraryMangaDTO, len(lib.Mangas)),
	}

	for i, manga := range lib.Mangas {
		dto.Mangas[i] = *m.ToLibraryMangaDTO(&manga)
	}

	return dto
}

func (m *mapper) ToLibraryMangaDTO(manga *model.LibraryManga) *LibraryMangaDTO {
	return &LibraryMangaDTO{
		MangaID: manga.MangaID.String(),
		Tags:    manga.Tags,
		AddedAt: manga.AddedAt.Format(time.RFC3339),
	}
}

func (m *mapper) ToLibrarySummaryDTO(lib *repository.LibrarySummary) *LibrarySummaryDTO {
	return &LibrarySummaryDTO{
		Tags:        lib.Tags,
		TotalMangas: lib.TotalMangas,
	}
}
