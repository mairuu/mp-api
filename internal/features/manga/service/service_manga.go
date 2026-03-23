package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/mairuu/mp-api/internal/app"
	"github.com/mairuu/mp-api/internal/app/paging"
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

	r, err := s.processCoverArtChanges(nil, &req.Covers)
	if err != nil {
		return nil, err
	}

	m, err := model.NewManga(ur.ID, req.Title, req.Synopsis, model.MangaStatus(req.Status), r.Merged())
	if err != nil {
		return nil, err
	}

	covers, err := s.processStagingCoverArts(ctx, m, ur)
	if err != nil {
		return nil, err
	}

	err = m.Updater().
		CoverArts(covers).
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

func (s *Service) ListMangas(ctx context.Context, ur *app.UserRole, q *MangaListQuery) (*paging.PagedDTO, error) {
	err := s.enforce(ur, model.ResourceManga, model.ActionRead, nil)
	if err != nil {
		return nil, err
	}

	if len(q.Orders) == 0 {
		q.Orders = []string{"created_at,desc"}
	}

	r, err := s.repo.ListMangas(ctx, q.ToMangaFilter(), q.ToPaging(), q.ToOrdering())
	if err != nil {
		return nil, err
	}

	items := make([]MangaSummaryDTO, len(r.Items))
	for i := range r.Items {
		items[i] = s.mapper.ToMangaSummaryDTO(&r.Items[i])
	}

	dto := paging.NewPagedDTOFromPaging(r.Total, r.Limit, r.Offset, items)

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

	r, err := s.processCoverArtChanges(m.Covers, req.Covers)
	if err != nil {
		return nil, err
	}

	err = m.Updater().
		Title(req.Title).
		Synopsis(req.Synopsis).
		Status((*model.MangaStatus)(req.Status)).
		CoverArts(r.Merged()).
		Apply()
	if err != nil {
		return nil, err
	}

	covers, err := s.processStagingCoverArts(ctx, m, ur)
	if err != nil {
		return nil, err
	}

	err = m.Updater().
		CoverArts(covers).
		Apply()
	if err != nil {
		return nil, err
	}

	if err := s.repo.SaveManga(ctx, m); err != nil {
		return nil, err
	}

	// delete removed cover arts from storage
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
		isPrimary := false
		if dto.IsPrimary != nil {
			isPrimary = *dto.IsPrimary
		}
		cv, err := model.NewStagingCoverArt(dto.ObjectName, isPrimary, dto.Volume, dto.Description)
		if err != nil {
			return nil, err
		}
		newCovers = append(newCovers, *cv)
	}

	differ := collections.IdentifiableDiffer[string, model.CoverArt]{
		GetKey: func(c *model.CoverArt) string {
			if c.Volume != nil {
				return *c.Volume
			}
			return c.ObjectName
		},
		AddItem: func(existings []model.CoverArt, adding *model.CoverArt) (added *model.CoverArt, toUpdate *model.CoverArt, toDelete *model.CoverArt, err error) {
			// check if this object already exists with a different volume (volume changed scenario)
			for i := range existings {
				if existings[i].ObjectName != adding.ObjectName {
					continue
				}
				// same object, different volume; treat as update
				updated, err := adding.ToStaged(adding.ObjectName)
				if err != nil {
					return nil, nil, nil, err
				}
				return nil, updated, nil, nil
			}
			// truly new cover
			return adding, nil, nil, nil
		},
		UpdateItem: func(existing *model.CoverArt, updating *model.CoverArt) (*model.CoverArt, *model.CoverArt, *model.CoverArt, error) {
			// if object name changed, treat as replacement (delete old + add new)
			if existing.ObjectName != updating.ObjectName {
				// return: updated=nil, toAdd=updating, toDelete=old (for cleanup)
				return nil, updating, existing, nil
			}

			// otherwise, it's a simple update
			updated, err := updating.ToStaged(existing.ObjectName)
			if err != nil {
				return nil, nil, nil, err
			}
			return updated, nil, nil, nil
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

	return differ.Diff(existing, newCovers)
}

func (s *Service) processStagingCoverArts(ctx context.Context, m *model.Manga, ur *app.UserRole) ([]model.CoverArt, error) {
	covers := make([]model.CoverArt, len(m.Covers))
	for i := range m.Covers {
		c := &m.Covers[i]
		if !c.IsStaging() {
			covers[i] = *c
			continue
		}

		permanentObjectName, err := s.processNewCoverArt(ctx, ur, m.ID, c.ObjectName)
		if err != nil {
			return nil, err
		}
		c, err = c.ToStaged(permanentObjectName)
		if err != nil {
			return nil, err
		}
		covers[i] = *c
	}
	return covers, nil
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

	// check if the staging object belongs to the user
	// todo: create a session-based temporary bucket
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
