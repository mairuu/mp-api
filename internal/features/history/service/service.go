package service

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/mairuu/mp-api/internal/app"
	"github.com/mairuu/mp-api/internal/app/paging"
	"github.com/mairuu/mp-api/internal/features/history/model"
	"github.com/mairuu/mp-api/internal/features/history/repository"
)

type Service struct {
	mapper mapper
	repo   repository.Repository
}

func NewService(log *slog.Logger, repo repository.Repository) *Service {
	return &Service{
		mapper: mapper{},
		repo:   repo,
	}
}

func (s *Service) ListRecent(ctx context.Context, ur *app.UserRole, q HistoryListQuery) (*paging.PagedDTO, error) {
	r, err := s.repo.ListRecent(ctx, ur.ID, q.ToPaging())
	if err != nil {
		return nil, err
	}

	items := make([]RecentReadDTO, len(r.Items))
	for i := range r.Items {
		items[i] = s.mapper.ToRecentReadDTO(&r.Items[i])
	}

	totalPages := (r.Total + q.PageSize - 1) / q.PageSize
	paged := paging.NewPagedDTO(r.Total, totalPages, q.PageSize, q.Page, items)
	return &paged, nil
}

func (s *Service) ListByManga(ctx context.Context, ur *app.UserRole, mangaID uuid.UUID, q HistoryListQuery) (*paging.PagedDTO, error) {
	r, err := s.repo.ListByManga(ctx, ur.ID, mangaID, q.ToPaging())
	if err != nil {
		return nil, err
	}

	items := make([]MangaReadDTO, len(r.Items))
	for i := range r.Items {
		items[i] = s.mapper.ToMangaReadDTO(&r.Items[i])
	}

	totalPages := (r.Total + q.PageSize - 1) / q.PageSize
	dto := paging.NewPagedDTO(r.Total, totalPages, q.PageSize, q.Page, items)
	return &dto, nil
}

func (s *Service) MarkChaptersRead(ctx context.Context, ur *app.UserRole, req MarkChaptersAsReadDTO) error {
	histories := make([]model.History, len(req.Chapters))

	for i := range req.Chapters {
		chapterUUID, err := uuid.Parse(req.Chapters[i].ChapterID)
		if err != nil {
			return err
		}
		histories[i] = model.NewHistory(ur.ID, chapterUUID, req.Chapters[i].Progress)
	}

	return s.repo.SaveMany(ctx, histories)
}

func (s *Service) UnmarkChaptersRead(ctx context.Context, ur *app.UserRole, req UnmarkChaptersAsReadDTO) error {
	chapterUUIDs := make([]uuid.UUID, len(req.Chapters))

	for i := range req.Chapters {
		chapterUUID, err := uuid.Parse(req.Chapters[i].ChapterID)
		if err != nil {
			return err
		}
		chapterUUIDs[i] = chapterUUID
	}

	return s.repo.DeleteByChapters(ctx, ur.ID, chapterUUIDs)
}
