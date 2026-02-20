package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/mairuu/mp-api/internal/app"
	"github.com/mairuu/mp-api/internal/features/manga/model"
)

func (s *Service) CreateChapter(ctx context.Context, ur *app.UserRole, req CreateChapterDTO) (*ChapterDTO, error) {
	mangaID, err := uuid.Parse(req.MangaID) // should be valid due to binding validation
	if err != nil {
		return nil, err
	}
	m, err := s.repo.GetMangaByID(ctx, mangaID)
	if err != nil {
		return nil, err
	}

	// pass manga as target for scope resolution, so that owner can create chapters for their manga
	err = s.enforce(ur, model.ResourceChapter, model.ActionCreate, m)
	if err != nil {
		return nil, err
	}

	c, err := model.NewChapter(m.ID, req.Title, req.Number, req.Volume)
	if err != nil {
		return nil, err
	}

	err = s.repo.SaveChapter(ctx, c)
	if err != nil {
		return nil, err
	}

	dto := s.mapper.ToChapterDTO(c)
	return &dto, nil
}

func (s *Service) ListChapters(ctx context.Context, ur *app.UserRole, q *ChapterListQuery) (*PagedDTO, error) {
	err := s.enforce(ur, model.ResourceChapter, model.ActionRead, nil)
	if err != nil {
		return nil, err
	}

	if len(q.Orders) == 0 {
		q.Orders = []string{"created_at,desc"}
	}

	filter := q.ToChapterFilter()
	filter.State = ptr(string(model.ChapterStatePublish))
	paging := q.ToPaging()
	ordering := q.ToOrdering([]string{"title", "number", "volume", "created_at"})

	total, err := s.repo.CountChapters(ctx, filter)
	if err != nil {
		return nil, err
	}

	cs, err := s.repo.ListChapters(ctx, filter, paging, ordering)
	if err != nil {
		return nil, err
	}

	items := make([]ChapterSummaryDTO, len(cs))
	for i, c := range cs {
		items[i] = s.mapper.ToChapterSummaryDTO(&c)
	}

	totalPages := (total + q.PageSize - 1) / q.PageSize
	dto := NewPagedDTO(total, totalPages, q.PageSize, q.Page, items)

	return &dto, nil
}

func (s *Service) GetChapterByID(ctx context.Context, ur *app.UserRole, id uuid.UUID) (*ChapterDTO, error) {
	c, err := s.repo.GetChapterByID(ctx, id)
	if err != nil {
		return nil, err
	}

	m, err := s.repo.GetMangaByID(ctx, c.MangaID)
	if err != nil {
		return nil, err
	}

	err = s.enforce(ur, model.ResourceChapter, model.ActionRead, m)
	if err != nil {
		return nil, err
	}

	dto := s.mapper.ToChapterDTO(c)
	return &dto, nil
}

func (s *Service) UpdateChapter(ctx context.Context, ur *app.UserRole, id uuid.UUID, req UpdateChapterDTO) (*ChapterDTO, error) {
	c, err := s.repo.GetChapterByID(ctx, id)
	if err != nil {
		return nil, err
	}

	m, err := s.repo.GetMangaByID(ctx, c.MangaID)
	if err != nil {
		return nil, err
	}

	err = s.enforce(ur, model.ResourceChapter, model.ActionUpdate, m)
	if err != nil {
		return nil, err
	}

	err = c.Updater().
		Title(req.Title).
		Volume(req.Volume).
		Number(req.Number).
		Apply()
	if err != nil {
		return nil, err
	}

	err = s.repo.SaveChapter(ctx, c)
	if err != nil {
		return nil, err
	}

	dto := s.mapper.ToChapterDTO(c)
	return &dto, nil
}

func (s *Service) DeleteChapter(ctx context.Context, ur *app.UserRole, id uuid.UUID) error {
	c, err := s.repo.GetChapterByID(ctx, id)
	if err != nil {
		return err
	}

	m, err := s.repo.GetMangaByID(ctx, c.MangaID)
	if err != nil {
		return err
	}

	err = s.enforce(ur, model.ResourceChapter, model.ActionDelete, m)
	if err != nil {
		return err
	}

	return s.repo.DeleteChapterByID(ctx, id)
}
