package mappers

import (
	"github.com/google/uuid"
	"github.com/mairuu/mp-api/internal/features/manga/model"
	"github.com/mairuu/mp-api/internal/persistence/models"
	"github.com/shopspring/decimal"
)

func ToMangaDB(m *model.Manga) models.MangaDB {
	covers := make([]models.CoverArtDB, 0, len(m.Covers))
	for i, c := range m.Covers {
		cdb := toCoverArtDB(&c, m.ID)
		cdb.Order = i
		covers = append(covers, cdb)
	}

	return models.MangaDB{
		ID:        m.ID,
		OwnerID:   m.OwnerID,
		Title:     m.Title,
		Synopsis:  m.Synopsis,
		Status:    string(m.Status),
		Covers:    covers,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func MangaDBToModel(mdb *models.MangaDB) model.Manga {
	covers := make([]model.CoverArt, 0, len(mdb.Covers))
	for _, cdb := range mdb.Covers {
		covers = append(covers, coverArtDBToModel(&cdb))
	}

	return model.Manga{
		ID:        mdb.ID,
		OwnerID:   mdb.OwnerID,
		Title:     mdb.Title,
		Synopsis:  mdb.Synopsis,
		Status:    model.MangaStatus(mdb.Status),
		Covers:    covers,
		CreatedAt: mdb.CreatedAt,
		UpdatedAt: mdb.UpdatedAt,
	}
}

func toCoverArtDB(cover *model.CoverArt, mangaID uuid.UUID) models.CoverArtDB {
	var vol *decimal.Decimal
	if cover.Volume != nil {
		v, err := decimal.NewFromString(*cover.Volume)
		if err == nil {
			vol = &v
		}
	}
	return models.CoverArtDB{
		MangaID:     mangaID,
		IsPrimary:   cover.IsPrimary,
		Volume:      vol,
		ObjectName:  cover.ObjectName,
		Description: cover.Description,
	}
}

func coverArtDBToModel(cdb *models.CoverArtDB) model.CoverArt {
	var vol *string
	if cdb.Volume != nil {
		v := cdb.Volume.String()
		vol = &v
	}
	return model.CoverArt{
		IsPrimary:   cdb.IsPrimary,
		Volume:      vol,
		ObjectName:  cdb.ObjectName,
		Description: cdb.Description,
	}
}

func ToChapterDB(c *model.Chapter) models.ChapterDB {
	pages := make([]models.ChapterPageDB, 0, len(c.Pages))
	for i := range c.Pages {
		pages = append(pages, toPageDB(&c.Pages[i], c.ID, i+1))
	}

	var vol *decimal.Decimal
	if c.Volume != nil {
		v, err := decimal.NewFromString(*c.Volume)
		if err == nil {
			vol = &v
		}
	}
	num, err := decimal.NewFromString(c.Number)
	if err != nil {
		num = decimal.Zero
	}

	return models.ChapterDB{
		ID:        c.ID,
		MangaID:   c.MangaID,
		Title:     c.Title,
		Volume:    vol,
		Number:    num,
		State:     string(c.State),
		Pages:     pages,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

func ChapterDBToModel(cdb *models.ChapterDB) model.Chapter {
	pages := make([]model.ChapterPage, 0, len(cdb.Pages))
	for i := range cdb.Pages {
		pages = append(pages, pageDBToModel(&cdb.Pages[i]))
	}

	var vol *string
	if cdb.Volume != nil {
		v := cdb.Volume.String()
		vol = &v
	}

	return model.Chapter{
		ID:        cdb.ID,
		MangaID:   cdb.MangaID,
		Title:     cdb.Title,
		Volume:    vol,
		Number:    cdb.Number.String(),
		State:     model.ChapterState(cdb.State),
		Pages:     pages,
		CreatedAt: cdb.CreatedAt,
		UpdatedAt: cdb.UpdatedAt,
	}
}

func toPageDB(page *model.ChapterPage, chapterID uuid.UUID, number int) models.ChapterPageDB {
	return models.ChapterPageDB{
		ChapterID:  chapterID,
		Number:     number,
		Width:      page.Width,
		Height:     page.Height,
		ObjectName: page.ObjectName,
	}
}

func pageDBToModel(pdb *models.ChapterPageDB) model.ChapterPage {
	return model.NewChapterPage(pdb.ObjectName, pdb.Width, pdb.Height)
}
