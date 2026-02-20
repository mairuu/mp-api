package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/mairuu/mp-api/internal/app"
	"github.com/mairuu/mp-api/internal/features/manga/model"
	"github.com/mairuu/mp-api/internal/platform/authorization"
	"github.com/mairuu/mp-api/internal/platform/collections"
	"github.com/mairuu/mp-api/internal/platform/storage"
	"github.com/nfnt/resize"
)

func (s *Service) CreateManga(ctx context.Context, ur *app.UserRole, req CreateMangaDTO) (*MangaDTO, error) {
	err := s.enforce(ur, model.ResourceManga, model.ActionCreate, nil)
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
	err := s.enforce(ur, model.ResourceManga, model.ActionCreate, nil)
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

	err = s.enforce(ur, model.ResourceManga, model.ActionRead, m)
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

	if err := s.enforce(ur, model.ResourceManga, model.ActionUpdate, m); err != nil {
		return nil, err
	}

	r, err := s.processCoverArtChanges(m.Covers, req.CoverArts)
	if err != nil {
		return nil, err
	}

	err = m.Updater().
		Title(req.Title).
		Synopsis(req.Synopsis).
		Status((*model.MangaStatus)(req.Status)).
		CoverArts(r.Merged()). // for validation of cover arts before processing new ones
		Apply()
	if err != nil {
		return nil, err
	}

	// stage new cover arts
	if len(r.Added) > 0 {
		for _, c := range r.Added {
			permanentObjectName, err := s.processNewCoverArt(ctx, ur, m.ID, c.ObjectName)
			if err != nil {
				return nil, err
			}
			c.ObjectName = permanentObjectName
		}
	}

	err = m.Updater().
		CoverArts(r.Merged()).
		Apply()
	if err != nil {
		return nil, err
	}

	if err := s.repo.SaveManga(ctx, m); err != nil {
		return nil, err
	}

	// delete removed cover arts from storage
	if len(r.Deleted) > 0 {
		for _, c := range r.Deleted {
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

type processCoverArtChangesResult = collections.DiffResult[model.CoverArt]

// processCoverArtChanges processes cover art changes based on the provided DTOs
// and existing cover arts, returning what needs to be added, updated, or deleted
// handled scenarios:
// - add new cover
// - update existing cover (same volume, same object)
// - replace cover (same volume, different object)
// - rename volume (different volume, same object)
// - delete cover
func (s *Service) processCoverArtChanges(existing []model.CoverArt, dtos *[]UpdateCoverArtDTO) (*processCoverArtChangesResult, error) {
	if dtos == nil {
		// return empty diff result with existing items as "updated"
		result := &collections.DiffResult[model.CoverArt]{
			Added:   nil,
			Updated: make([]*model.CoverArt, len(existing)),
			Deleted: nil,
		}
		for i := range existing {
			result.Updated[i] = &existing[i]
		}
		return result, nil
	}

	// convert
	newCovers := make([]model.CoverArt, 0, len(*dtos))
	for _, dto := range *dtos {
		cv, err := model.NewCoverArt(dto.Volume, dto.ObjectName, dto.Description)
		if err != nil {
			return nil, err
		}
		newCovers = append(newCovers, *cv)
	}

	differ := collections.IdentifiableDiffer[string, model.CoverArt]{
		GetKey: func(c *model.CoverArt) string {
			return c.Volume
		},
		AddItem: func(existings []model.CoverArt, adding *model.CoverArt) (added *model.CoverArt, toUpdate *model.CoverArt, toDelete *model.CoverArt, err error) {
			// check if this object already exists with a different volume (volume rename scenario)
			for i := range existings {
				if existings[i].ObjectName == adding.ObjectName {
					// same object, different volume; treat as update
					return nil, adding, nil, nil
				}
			}
			// truly new cover
			return adding, nil, nil, nil
		},
		UpdateItem: func(existing *model.CoverArt, updating *model.CoverArt) (*model.CoverArt, *model.CoverArt, *model.CoverArt, error) {
			// if object name changed, treat as replacement (delete old + add new)
			if existing.ObjectName != updating.ObjectName {
				// return: updated=nil, toAdd=updating, toDelete=old (for cleanup)
				return nil, updating, &model.CoverArt{ObjectName: existing.ObjectName}, nil
			}

			// otherwise, it's a simple update (description might have changed)
			updated := *existing
			updated.Description = updating.Description
			return &updated, nil, nil, nil
		},
		DeleteItem: func(new []model.CoverArt, deleting *model.CoverArt) (deleted *model.CoverArt, toAdd *model.CoverArt, toUpdate *model.CoverArt, err error) {
			for i := range new {
				if new[i].ObjectName == deleting.ObjectName {
					// same object, different volume; skip deletion (additem will handle it)
					return nil, nil, nil, nil
				}
			}
			// truly deleted cover
			return deleting, nil, nil, nil
		},
	}

	result, err := differ.Diff(existing, newCovers)
	if err != nil {
		return nil, err
	}

	return result, nil
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
			return "", model.ErrCoverNotFound.WithArg("object_name", statingObjectName)
		}
		return "", err
	}

	if meta.MetaData["user_id"] != ur.ID.String() {
		return "", model.ErrCoverNotFound.WithArg("object_name", statingObjectName)
	}

	f, err := s.temporaryBucket.Download(ctx, statingObjectName)
	if err != nil {
		// practically redundant check since we already got the metadata
		if errors.Is(err, storage.ErrObjectNotFound) {
			return "", model.ErrCoverNotFound.WithArg("object_name", statingObjectName)
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

	if err := s.enforce(ur, model.ResourceManga, model.ActionDelete, m); err != nil {
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

	err = s.enforce(ur, model.ResourceManga, model.ActionPublish, m)
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

func (s *Service) enforce(ur *app.UserRole, resource authorization.Resource, action authorization.Action, target authorization.ScopeResolvable) error {
	return s.enforcer.Enforce(ur.ID, ur.Role, resource, action, target)
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
