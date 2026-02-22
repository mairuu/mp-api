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

	var coverURL *string
	if m.CoverObjecrtName != nil {
		coverURL = m.CoverObjecrtName
	}

	return MangaSummaryDTO{
		ID:       m.ID.String(),
		Title:    m.Title,
		CoverURL: coverURL,
	}
}

func (mp *mapper) ToMangaDTO(m *model.Manga) MangaDTO {
	if m == nil {
		return MangaDTO{}
	}

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
			IsPrimary:   sorted[i].IsPrimary,
			Description: sorted[i].Description,
			ObjectName:  sorted[i].ObjectName,
		})
	}

	return MangaDTO{
		ID:        m.ID.String(),
		Title:     m.Title,
		Synopsis:  m.Synopsis,
		Status:    string(m.Status),
		CoverArts: covers,
	}
}

func (mp *mapper) ToChapterSummaryDTO(c *repo.ChapterSummary) ChapterSummaryDTO {
	if c == nil {
		return ChapterSummaryDTO{}
	}

	return ChapterSummaryDTO{
		ID:      c.ID.String(),
		MangaID: c.MangaID.String(),
		Title:   c.Title,
		Volume:  c.Volume,
		Number:  c.Number,
	}
}

func (mp *mapper) ToChapterDTO(c *model.Chapter) ChapterDTO {
	if c == nil {
		return ChapterDTO{}
	}

	pages := make([]PageDTO, 0, len(c.Pages))
	for i := range c.Pages {
		pages = append(pages, PageDTO{
			Width:      c.Pages[i].Width,
			Height:     c.Pages[i].Height,
			ObjectName: c.Pages[i].ObjectName,
		})
	}

	return ChapterDTO{
		ID:      c.ID.String(),
		MangaID: c.MangaID.String(),
		Title:   c.Title,
		Volume:  c.Volume,
		Number:  c.Number,
		Pages:   pages,
	}
}
