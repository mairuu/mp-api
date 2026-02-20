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
