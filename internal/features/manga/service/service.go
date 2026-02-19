package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/mairuu/mp-api/internal/app"
	"github.com/mairuu/mp-api/internal/features/manga/model"
	repo "github.com/mairuu/mp-api/internal/features/manga/repository"
	"github.com/mairuu/mp-api/internal/platform/authorization"
	"github.com/mairuu/mp-api/internal/platform/storage"
	"github.com/nfnt/resize"
)

type Service struct {
	log             *slog.Logger
	repo            repo.Repository
	enforcer        *authorization.Enforcer
	publicBucket    storage.Bucket
	temporaryBucket storage.Bucket
	mapper          mapper
}

func NewService(log *slog.Logger, repo repo.Repository, enforcer *authorization.Enforcer, publicBucket storage.Bucket, temporaryBucket storage.Bucket) *Service {
	return &Service{
		log:             log,
		repo:            repo,
		enforcer:        enforcer,
		publicBucket:    publicBucket,
		temporaryBucket: temporaryBucket,
		mapper:          mapper{},
	}
}

func (s *Service) CreateManga(ctx context.Context, ur *app.UserRole, req CreateMangaDTO) (*MangaDTO, error) {
	err := s.enforce(ur, model.ActionCreate, nil)
	if err != nil {
		return nil, err
	}

	m, err := model.NewManga(ur.ID, req.Title, req.Synopsis, model.MangaStatus(req.Status))
	if err != nil {
		return nil, err
	}

	if err := s.repo.SaveManga(ctx, m); err != nil {
		return nil, err
	}

	dto := s.mapper.ToMangaDTO(m)
	return &dto, nil
}

func (s *Service) ListMangas(ctx context.Context, ur *app.UserRole, q *MangaListQuery) (*PagedDTO, error) {
	err := s.enforce(ur, model.ActionList, nil)
	if err != nil {
		return nil, err
	}

	filter := q.ToMangaFilter()
	filter.State = ptr(string(model.MangaStatePublish)) // only list active mangas
	pagging := q.ToPaging()
	ordering := q.ToOrdering([]string{"title", "created_at", "updated_at"})

	total, err := s.repo.CountMangas(ctx, filter)
	if err != nil {
		return nil, err
	}

	ms, err := s.repo.ListMangas(ctx, filter, pagging, ordering)
	if err != nil {
		return nil, err
	}

	items := make([]MangaSummaryDTO, len(ms))
	for i, m := range ms {
		items[i] = s.mapper.ToMangaSummaryDTO(&m)
	}

	totalPages := (total + q.PageSize - 1) / q.PageSize
	dto := NewPagedDTO(total, totalPages, q.PageSize, q.Page, items)

	return &dto, nil
}

func (s *Service) GetMangaByID(ctx context.Context, ur *app.UserRole, id uuid.UUID) (*MangaDTO, error) {
	m, err := s.repo.GetMangaByID(ctx, id)
	if err != nil {
		return nil, err
	}

	err = s.enforce(ur, model.ActionRead, m)
	if err != nil {
		return nil, err
	}

	dto := s.mapper.ToMangaDTO(m)
	return &dto, nil
}

func (s *Service) UpdateManga(ctx context.Context, ur *app.UserRole, id uuid.UUID, req UpdateMangaDTO) (*MangaDTO, error) {
	m, err := s.repo.GetMangaByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := s.enforce(ur, model.ActionUpdate, m); err != nil {
		return nil, err
	}

	r, err := s.processCoverArtChanges(m, req.CoverArts)
	if err != nil {
		return nil, err
	}

	err = m.Updater().
		Title(req.Title).
		Synopsis(req.Synopsis).
		Status((*model.MangaStatus)(req.Status)).
		CoverArts(r.merged()). // for validation of cover arts before processing new ones
		Apply()
	if err != nil {
		return nil, err
	}

	// stage new cover arts
	if len(r.added) > 0 {
		for _, c := range r.added {
			permanentObjectName, err := s.processNewCoverArt(ctx, ur, m.ID, c.ObjectName)
			if err != nil {
				return nil, err
			}
			c.ObjectName = permanentObjectName
		}
	}

	err = m.Updater().
		CoverArts(r.merged()).
		Apply()
	if err != nil {
		return nil, err
	}

	if err := s.repo.SaveManga(ctx, m); err != nil {
		return nil, err
	}

	// delete removed cover arts from storage
	if len(r.deleted) > 0 {
		for _, c := range r.deleted {
			for _, spec := range coverImageSpecs {
				for _, suffix := range spec.suffixes {
					objectName := c.ObjectName + suffix
					if err := s.publicBucket.Delete(ctx, objectName); err != nil {
						s.log.WarnContext(ctx, "failed to delete cover art object", "object_name", objectName, "error", err)
					}
				}
			}
		}
	}

	dto := s.mapper.ToMangaDTO(m)
	return &dto, nil
}

type processCoverArtChangesResult struct {
	m       *model.Manga
	added   []*model.CoverArt
	updated []*model.CoverArt
	deleted []*model.CoverArt
}

func (r *processCoverArtChangesResult) merged() []model.CoverArt {
	if r.m != nil {
		return r.m.Covers
	}

	all := make([]model.CoverArt, 0, len(r.added)+len(r.updated))
	for _, c := range r.added {
		all = append(all, *c)
	}
	for _, c := range r.updated {
		all = append(all, *c)
	}
	return all
}

// processCoverArtChanges compares the existing cover arts of a manga with the
// provided DTOs and determines which covers need to be added, updated, or deleted.
func (s *Service) processCoverArtChanges(m *model.Manga, dtos *[]UpdateCoverArtDTO) (*processCoverArtChangesResult, error) {
	if dtos == nil {
		return &processCoverArtChangesResult{m: m}, nil
	}

	var added []*model.CoverArt
	var updated []*model.CoverArt
	var deleted []*model.CoverArt

	existingCovers := make(map[uuid.UUID]*model.CoverArt)
	for i := range m.Covers {
		existingCovers[m.Covers[i].ID] = &m.Covers[i]
	}

	for _, coverDTO := range *dtos {
		// case 1: update existing cover
		if coverDTO.ID != nil {
			if existing, ok := existingCovers[*coverDTO.ID]; ok {
				delete(existingCovers, *coverDTO.ID)

				u := *existing
				u.Volume = coverDTO.Volume
				u.Description = coverDTO.Description

				// handle object name change
				if coverDTO.ObjectName != nil {
					// the goal is to replace the existing cover art with the new one and keep the same ID
					u.ObjectName = *coverDTO.ObjectName
					added = append(added, &u)
					// create dummy cover art to be deleted after the new one is processed and saved
					deleted = append(deleted, &model.CoverArt{ObjectName: existing.ObjectName})
				} else {
					updated = append(updated, &u)
				}
			} else {
				// id was not found in existing covers
				return nil, fmt.Errorf("%w: id = %s", model.ErrCoverNotFound, *coverDTO.ID)
			}

			continue
		}

		// case 2: add new cover from staging
		if coverDTO.ObjectName != nil {
			cv, err := model.NewCoverArt(coverDTO.Volume, *coverDTO.ObjectName, coverDTO.Description)
			if err != nil {
				return nil, err
			}
			added = append(added, cv)

			continue
		}

		return nil, fmt.Errorf("%w: %s", model.ErrCoverNotFound, "cover DTO must have either ID or ObjectName")
	}

	for i := range existingCovers {
		deleted = append(deleted, existingCovers[i])
	}

	return &processCoverArtChangesResult{
		added:   added,
		updated: updated,
		deleted: deleted,
	}, nil
}

var coverImageSpecs = []struct {
	width    uint
	height   uint
	suffixes []string
}{
	{10000, 10000, []string{""}},   // original
	{256, 10000, []string{"_256"}}, // thumbnail
	{512, 10000, []string{"_512"}}, // banner
}

func (s *Service) processNewCoverArt(ctx context.Context, ur *app.UserRole, mangaID uuid.UUID, statingObjectName string) (string, error) {
	// check if user owns the staging object
	meta, err := s.temporaryBucket.GetMetadata(ctx, statingObjectName)
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotFound) {
			return "", fmt.Errorf("%w: object_name = %s", model.ErrCoverNotFound, statingObjectName)
		}
		return "", err
	}

	if meta.MetaData["user_id"] != ur.ID.String() {
		return "", fmt.Errorf("%w: object_name = %s", model.ErrCoverNotFound, statingObjectName)
	}

	f, err := s.temporaryBucket.Download(ctx, statingObjectName)
	if err != nil {
		// practically redundant check since we already got the metadata
		if errors.Is(err, storage.ErrObjectNotFound) {
			return "", fmt.Errorf("%w: object_name = %s", model.ErrCoverNotFound, statingObjectName)
		}
		return "", err
	}
	defer f.Close()

	img, err := s.decodeImage(f)
	if err != nil {
		return "", err
	}

	fileName := uuid.New().String()
	baseObjectName := mangaCoverObjectName(mangaID, fileName)

	for _, spec := range coverImageSpecs {
		scaled := resize.Thumbnail(spec.width, spec.height, img, resize.Lanczos3)
		for _, suffix := range spec.suffixes {
			if err := s.uploadImage(ctx, baseObjectName+suffix, scaled); err != nil {
				return "", err
			}
		}
	}

	// delete staging object from temporary bucket
	if err := s.temporaryBucket.Delete(ctx, statingObjectName); err != nil {
		s.log.WarnContext(ctx, "failed to delete staging object after processing new cover art", "object_name", statingObjectName, "error", err)
	}

	return baseObjectName, nil
}

func (s *Service) DeleteManga(ctx context.Context, ur *app.UserRole, id uuid.UUID) error {
	m, err := s.repo.GetMangaByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.enforce(ur, model.ActionDelete, m); err != nil {
		return err
	}

	err = s.repo.DeleteMangaByID(ctx, id)
	if err != nil {
		return err
	}

	// release resources from public bucket
	for objectName := range s.publicBucket.ListIter(ctx, mangaResourcePrefix(m.ID)) {
		if err = s.publicBucket.Delete(ctx, objectName); err != nil {
			s.log.WarnContext(ctx, "failed to delete cover art object during manga deletion", "object_name", objectName, "error", err)
		}
	}

	return nil
}

func (s *Service) PublishManga(ctx context.Context, ur *app.UserRole, id uuid.UUID) (*MangaDTO, error) {
	m, err := s.repo.GetMangaByID(ctx, id)
	if err != nil {
		return nil, err
	}

	err = s.enforce(ur, model.ActionPublish, m)
	if err != nil {
		return nil, err
	}

	ms := model.MangaStatePublish
	// for now, publishing just update the manga state to published
	err = m.Updater().
		State(&ms).
		Apply()

	if err != nil {
		return nil, err
	}

	if err := s.repo.SaveManga(ctx, m); err != nil {
		return nil, err
	}

	dto := s.mapper.ToMangaDTO(m)
	return &dto, nil
}

// enforce is a helper method to enforce authorization checks for manga-related actions.
func (s *Service) enforce(ur *app.UserRole, action authorization.Action, target authorization.ScopeResolvable) error {
	return s.enforcer.Enforce(ur.ID, ur.Role, model.ResourceManga, action, target)
}

func ptr[T any](v T) *T {
	return &v
}

func mangaCoverObjectName(mangaID uuid.UUID, fileName string) string {
	return mangaResourcePrefix(mangaID) + fileName
}

func mangaResourcePrefix(mangaID uuid.UUID) string {
	return mangaID.String() + "/"
}
