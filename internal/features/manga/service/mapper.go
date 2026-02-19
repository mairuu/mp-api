package service

import (
	"fmt"
	"slices"

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

	// map cover arts
	// sort by volume
	sorted := make([]*model.CoverArt, 0, len(m.Covers))
	for i := range m.Covers {
		sorted = append(sorted, &m.Covers[i])
	}

	slices.SortFunc(sorted, func(a, b *model.CoverArt) int {
		// pad zeros
		aFormat := fmt.Sprintf("%010s", a.Volume)
		bFormat := fmt.Sprintf("%010s", b.Volume)

		if aFormat < bFormat {
			return -1
		} else if aFormat > bFormat {
			return 1
		}
		return 0
	})

	covers := make([]CoverArtDTO, 0, len(m.Covers))
	for i := range sorted {
		covers = append(covers, CoverArtDTO{
			Volume:      sorted[i].Volume,
			Description: sorted[i].Description,
			ObjectName:  sorted[i].ObjectName,
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
