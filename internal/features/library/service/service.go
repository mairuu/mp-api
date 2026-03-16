package service

import (
	"context"

	"github.com/mairuu/mp-api/internal/app"
	repo "github.com/mairuu/mp-api/internal/features/library/repository"
)

type Service struct {
	mapper mapper
	repo   repo.Repository
}

func NewService(repo repo.Repository) *Service {
	return &Service{
		mapper: mapper{},
		repo:   repo,
	}
}

// functionality to be implemented:
// - get library (with manga details)
// - upsert manga in library (add/remove manga, update tags)

func (s *Service) GetLibrary(ctx context.Context, ur *app.UserRole) (*LibraryDTO, error) {
	// todo: enforce authorization policies

	lib, err := s.repo.GetLibrary(ctx, ur.ID)
	if err != nil {
		return nil, err
	}

	return s.mapper.ToLibraryDTO(lib), nil
}

func (s *Service) GetLibrarySummary(ctx context.Context, ur *app.UserRole) (*LibrarySummaryDTO, error) {
	// todo: enforce authorization policies

	lib, err := s.repo.GetLibrarySummary(ctx, ur.ID)
	if err != nil {
		return nil, err
	}

	return s.mapper.ToLibrarySummaryDTO(lib), nil
}

// UpsertLibraryManga adds a manga to the user's library or updates its tags if it already exists.
// empty tags will be treated as an instruction to remove the manga from the library.
func (s *Service) UpsertLibraryMangas(ctx context.Context, ur *app.UserRole, mangas []UpsertLibraryMangaDTO) error {
	// todo: enforce authorization policies

	lib, err := s.repo.GetLibrary(ctx, ur.ID)
	if err != nil {
		return err
	}

	for i := range mangas {
		m, err := mangas[i].ToModel()
		if err != nil {
			return err
		}
		if len(m.Tags) == 0 {
			lib.RemoveManga(m.MangaID)
		} else {
			lib.UpsertManga(m.MangaID, m.Tags)
		}
	}

	return s.repo.SaveLibrary(ctx, lib)
}
