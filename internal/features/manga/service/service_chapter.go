package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/mairuu/mp-api/internal/app"
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

	pages := make([]model.Page, len(req.Pages))
	for i, objName := range req.Pages {
		pages[i] = model.Page{
			ObjectName: objName,
			Width:      1,
			Height:     1,
		}
	}
	c, err := model.NewChapter(m.ID, req.Number, req.Title, req.Volume, pages)
	if err != nil {
		return nil, err
	}

	for i := range c.Pages {
		img, err := s.processNewChapterPage(ctx, ur, c.ID, c.Pages[i].ObjectName)
		if err != nil {
			return nil, err
		}
		c.Pages[i].Width = img.width
		c.Pages[i].Height = img.height
		c.Pages[i].ObjectName = img.objectName
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

	r, err := s.processChapterPageChanges(c.Pages, req.Pages)
	if err != nil {
		return nil, err
	}

	err = c.Updater().
		Title(req.Title).
		Volume(req.Volume).
		Number(req.Number).
		Pages(ptr(r.Merged())).
		Apply()
	if err != nil {
		return nil, err
	}

	if len(r.Added) > 0 {
		for _, p := range r.Added {
			img, err := s.processNewChapterPage(ctx, ur, c.ID, p.ObjectName)
			if err != nil {
				return nil, err
			}
			p.Width = img.width
			p.Height = img.height
			p.ObjectName = img.objectName
		}
	}

	err = c.Updater().
		Pages(ptr(r.Merged())).
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

func (s *Service) processChapterPageChanges(existing []model.Page, dtos *[]string) (*collections.DiffResult[model.Page], error) {
	if dtos == nil {
		r := collections.DiffResult[model.Page]{
			Added:   nil,
			Updated: make([]*model.Page, len(existing)),
			Deleted: nil,
		}
		for i := range existing {
			r.Updated[i] = &existing[i]
		}
		return &r, nil
	}

	newPages := make([]model.Page, len(*dtos))
	for i, objName := range *dtos {
		newPages[i] = model.Page{ObjectName: objName, Width: 1, Height: 1}
	}

	differ := collections.IdentifiableDiffer[string, model.Page]{
		GetKey: func(p *model.Page) string {
			return p.ObjectName
		},
		UpdateItem: func(existing, updating *model.Page) (updated *model.Page, toAdd *model.Page, toDelete *model.Page, err error) {
			// pages are immutable, so we treat all matched items as updated without changes
			return existing, nil, nil, nil
		},
	}

	return differ.Diff(existing, newPages)
}

type pageImage struct {
	width      int
	height     int
	objectName string
}

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
