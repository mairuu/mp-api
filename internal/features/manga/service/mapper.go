package service

import (
	"github.com/mairuu/mp-api/internal/features/manga/model"
	repo "github.com/mairuu/mp-api/internal/features/manga/repository"
)

// helper struct for mapping between repository and service layer
type mapper struct{}

func (_ *mapper) ToMangaSummaryDTO(m *repo.MangaSummary) MangaSummaryDTO {
	if m == nil {
		return MangaSummaryDTO{}
	}

	return MangaSummaryDTO{
		ID:    m.ID.String(),
		Title: m.Title,
	}
}

func (mp *mapper) ToMangaDTO(m *model.Manga) MangaDTO {
	if m == nil {
		return MangaDTO{}
	}

	covers := make([]CoverArtDTO, 0, len(m.Covers))
	for i := range m.Covers {
		covers = append(covers, CoverArtDTO{
			ID:          m.Covers[i].ID.String(),
			Volume:      m.Covers[i].Volume,
			Description: m.Covers[i].Description,
			ObjectName:  m.Covers[i].ObjectName,
		})
	}

	return MangaDTO{
		ID:        m.ID.String(),
		Title:     m.Title,
		Synopsis:  m.Synopsis,
		Status:    string(m.Status),
		State:     string(m.State),
		CoverArts: covers,
	}
}
