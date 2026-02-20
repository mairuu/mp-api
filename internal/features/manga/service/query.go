package service

import (
	"slices"
	"strings"

	repo "github.com/mairuu/mp-api/internal/features/manga/repository"
)

type MangaListQuery struct {
	MangaFilterQuery
	PagingQuery
	OrderingQuery
}

type MangaFilterQuery struct {
	IDs    []string `form:"ids[]"`
	Title  *string  `form:"title"`
	Status *string  `form:"status"`
}

func (f *MangaFilterQuery) ToMangaFilter() repo.MangaFilter {
	return repo.MangaFilter{
		IDs:    f.IDs,
		Title:  f.Title,
		Status: f.Status,
	}
}

type ChapterListQuery struct {
	ChapterFilterQuery
	PagingQuery
	OrderingQuery
}

type ChapterFilterQuery struct {
	IDs      []string `form:"ids[]"`
	MangaIDs []string `form:"manga_ids[]"`
	Title    *string  `form:"title"`
	Number   *string  `form:"number"`
	Volume   *string  `form:"volume"`
}

func (f *ChapterFilterQuery) ToChapterFilter() repo.ChapterFilter {
	return repo.ChapterFilter{
		IDs:      f.IDs,
		MangaIDs: f.MangaIDs,
		Title:    f.Title,
		Number:   f.Number,
		Volume:   f.Volume,
	}
}

type OrderingQuery struct {
	// syntax: order=field1,asc&order=field2,desc
	Orders []string `form:"orders[]"`
}

func (o *OrderingQuery) ToOrdering(validFields []string) []repo.Ordering {
	var orderings []repo.Ordering
	for _, order := range o.Orders {
		parts := strings.Split(order, ",")
		if len(parts) != 2 {
			continue
		}
		field := parts[0]
		direction := parts[1]

		if !slices.Contains(validFields, field) {
			continue
		}

		dir := repo.Asc
		switch direction {
		case "desc", "DESC":
			dir = repo.Desc
		}

		orderings = append(orderings, repo.Ordering{
			Field:     field,
			Direction: dir,
		})
	}

	return orderings
}

type PagingQuery struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}

const (
	DefaultPageSize = 20
)

func (p *PagingQuery) normalize() {
	if p.PageSize <= 0 {
		p.PageSize = DefaultPageSize
	}
	if p.Page <= 0 {
		p.Page = 1
	}
}

func (p *PagingQuery) getLimitOffset() (limit, offset int) {
	limit = p.PageSize
	offset = (p.Page - 1) * p.PageSize
	return
}

func (p *PagingQuery) ToPaging() repo.Pagging {
	p.normalize()
	limit, offset := p.getLimitOffset()
	return repo.Pagging{Limit: limit, Offset: offset}
}
