package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/mairuu/mp-api/internal/app"
	"github.com/mairuu/mp-api/internal/app/paging"
	"github.com/mairuu/mp-api/internal/features/manga/model"
	"github.com/mairuu/mp-api/internal/platform/collections"
	"github.com/mairuu/mp-api/internal/platform/storage"
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

	err = s.enforce(ur, model.ResourceChapter, model.ActionCreate, m)
	if err != nil {
		return nil, err
	}

	r, err := s.processChapterPageChanges(nil, &req.Pages)
	if err != nil {
		return nil, err
	}

	c, err := model.NewChapter(m.ID, req.Number, req.Title, req.Volume, r.Merged())
	if err != nil {
		return nil, err
	}

	pages, err := s.processStagingChapterPages(ctx, ur, c)
	if err != nil {
		return nil, err
	}

	err = c.Updater().
		Pages(pages).
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

func (s *Service) ListChapters(ctx context.Context, ur *app.UserRole, q *ChapterListQuery) (*paging.PagedDTO, error) {
	err := s.enforce(ur, model.ResourceChapter, model.ActionRead, nil)
	if err != nil {
		return nil, err
	}

	if len(q.Orders) == 0 {
		q.Orders = []string{"created_at,desc"}
	}

	f := q.ToChapterFilter()
	f.State = ptr(string(model.ChapterStatePublish))
	r, err := s.repo.ListChapters(ctx, f, q.ToPaging(), q.ToOrdering())
	if err != nil {
		return nil, err
	}

	items := make([]ChapterSummaryDTO, len(r.Items))
	for i := range r.Items {
		items[i] = s.mapper.ToChapterSummaryDTO(&r.Items[i])
	}

	dto := paging.NewPagedDTOFromPaging(r.Total, r.Limit, r.Offset, items)

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

	r, err := s.processChapterPageChanges(c.Pages, req.Pages)
	if err != nil {
		return nil, err
	}

	err = c.Updater().
		Title(req.Title).
		Volume(req.Volume).
		Number(req.Number).
		Pages(r.Merged()).
		Apply()
	if err != nil {
		return nil, err
	}

	pages, err := s.processStagingChapterPages(ctx, ur, c)
	if err != nil {
		return nil, err
	}

	err = c.Updater().
		Pages(pages).
		Apply()
	if err != nil {
		return nil, err
	}

	err = s.repo.SaveChapter(ctx, c)
	if err != nil {
		return nil, err
	}

	if len(r.Deleted) > 0 {
		for _, p := range r.Deleted {
			objectName := p.ObjectName
			if err := s.publicBucket.Delete(ctx, objectName); err != nil {
				s.log.WarnContext(ctx, "failed to delete page object", "object_name", objectName, "error", err)
			}
		}
	}

	dto := s.mapper.ToChapterDTO(c)
	return &dto, nil
}

func (s *Service) processChapterPageChanges(existing []model.ChapterPage, dtos *[]string) (*collections.DiffResult[model.ChapterPage], error) {
	if dtos == nil {
		r := collections.DiffResult[model.ChapterPage]{
			Added:   nil,
			Updated: make([]*model.ChapterPage, len(existing)),
			Deleted: nil,
		}
		for i := range existing {
			r.Updated[i] = &existing[i]
		}
		return &r, nil
	}

	newPages := make([]model.ChapterPage, len(*dtos))
	for i, objName := range *dtos {
		newPages[i] = model.NewStagingChapterPage(objName)
	}

	differ := collections.IdentifiableDiffer[string, model.ChapterPage]{
		GetKey: func(p *model.ChapterPage) string {
			return p.ObjectName
		},
		UpdateItem: func(existing, updating *model.ChapterPage) (updated *model.ChapterPage, toAdd *model.ChapterPage, toDelete *model.ChapterPage, err error) {
			// pages are immutable, so we treat all matched items as updated without changes
			return existing, nil, nil, nil
		},
	}

	return differ.Diff(existing, newPages)
}

func (s *Service) processStagingChapterPages(ctx context.Context, ur *app.UserRole, c *model.Chapter) ([]model.ChapterPage, error) {
	pages := make([]model.ChapterPage, len(c.Pages))
	for i := range c.Pages {
		p := &c.Pages[i]
		if !p.IsStaging() {
			pages[i] = *p
			continue
		}

		img, err := s.processNewChapterPage(ctx, ur, c.ID, p.ObjectName)
		if err != nil {
			return nil, err
		}
		pages[i] = model.NewChapterPage(img.objectName, img.width, img.height)
	}
	return pages, nil
}

type pageImage struct {
	width      int
	height     int
	objectName string
}

// processNewChapterPage processes a new chapter page by validating the staging object,
// downloading and decoding the image, uploading it to the public bucket, and returning the page image info.
// it also deletes the staging object after processing.
// note: the process is not atomic
func (s *Service) processNewChapterPage(ctx context.Context, ur *app.UserRole, chapterID uuid.UUID, stagingObjectName string) (*pageImage, error) {
	// check if user owns the staging object
	meta, err := s.temporaryBucket.GetMetadata(ctx, stagingObjectName)
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotFound) {
			return nil, model.ErrPageNotFound.WithArg("object_name", stagingObjectName)
		}
		return nil, err
	}

	// check if the staging object belongs to the user
	// todo: create a session-based temporary bucket
	if meta.MetaData["user_id"] != ur.ID.String() {
		return nil, model.ErrPageNotFound.WithArg("object_name", stagingObjectName)
	}

	f, err := s.temporaryBucket.Download(ctx, stagingObjectName)
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotFound) {
			return nil, model.ErrPageNotFound.WithArg("object_name", stagingObjectName)
		}
		return nil, err
	}
	defer f.Close()

	img, err := s.decodeImage(f)
	if err != nil {
		return nil, err
	}

	fileName := uuid.New().String()
	objectName := chapterPageObjectName(chapterID, fileName)

	if err := s.uploadImage(ctx, objectName, img); err != nil {
		return nil, err
	}

	if err := s.temporaryBucket.Delete(ctx, stagingObjectName); err != nil {
		s.log.WarnContext(ctx, "failed to delete staging object after processing new page", "object_name", stagingObjectName, "error", err)
	}

	return &pageImage{
		width:      img.Bounds().Dx(),
		height:     img.Bounds().Dy(),
		objectName: objectName,
	}, nil
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

func chapterPageObjectName(chapterID uuid.UUID, filename string) string {
	return chapterResourcePrefix(chapterID) + filename
}

func chapterResourcePrefix(chapterID uuid.UUID) string {
	return chapterID.String() + "/"
}
