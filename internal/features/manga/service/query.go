package service

import (
	"github.com/mairuu/mp-api/internal/app/ordering"
	"github.com/mairuu/mp-api/internal/app/paging"
	repo "github.com/mairuu/mp-api/internal/features/manga/repository"
)

type MangaListQuery struct {
	MangaFilterQuery
	PagingQuery
	OrderingQuery
}

func (q *MangaListQuery) ToOrdering() []ordering.Ordering {
	return q.OrderingQuery.ToOrdering(
		repo.OrderByTitle,
		repo.OrderByCreatedAt,
		repo.OrderByUpdatedAt,
	)
}

type MangaFilterQuery struct {
	IDs      []string `form:"ids[]"`
	OwnerIDs []string `form:"owner_ids[]"`
	Title    *string  `form:"title"`
	Status   *string  `form:"status"`
}

func (f *MangaFilterQuery) ToMangaFilter() repo.MangaFilter {
	return repo.MangaFilter{
		IDs:      f.IDs,
		OwnerIDs: f.OwnerIDs,
		Title:    f.Title,
		Status:   f.Status,
	}
}

type PagingQuery struct {
	paging.Query
}

type OrderingQuery struct {
	ordering.Query
}

type ChapterListQuery struct {
	ChapterFilterQuery
	PagingQuery
	OrderingQuery
}

func (q *ChapterListQuery) ToOrdering() []ordering.Ordering {
	return q.OrderingQuery.ToOrdering(
		repo.OrderByTitle,
		repo.OrderByChapterNumber,
		repo.OrderByChapterVolume,
		repo.OrderByCreatedAt,
	)
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
