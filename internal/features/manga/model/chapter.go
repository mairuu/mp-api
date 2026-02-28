package model

import (
	"errors"
	"regexp"
	"time"

	"github.com/google/uuid"
)

type Chapter struct {
	ID        uuid.UUID
	MangaID   uuid.UUID
	Number    string
	Title     *string
	Volume    *string
	State     ChapterState
	Pages     []Page
	UpdatedAt time.Time
	CreatedAt time.Time
}

type ChapterState string

const (
	ChapterStateDraft   ChapterState = "draft"
	ChapterStatePublish ChapterState = "published"
)

type Page struct {
	Width      int
	Height     int
	ObjectName string
}

func NewChapter(mangaID uuid.UUID, number string, title, volume *string, pages []Page) (*Chapter, error) {
	if err := validateTitle(title); err != nil {
		return nil, err
	}
	if err := validateVolume(volume); err != nil {
		return nil, err
	}
	if err := validateNumber(number); err != nil {
		return nil, err
	}
	if err := validatePages(pages); err != nil {
		return nil, err
	}

	now := time.Now()
	return &Chapter{
		ID:        uuid.New(),
		MangaID:   mangaID,
		Title:     title,
		Volume:    volume,
		Number:    number,
		State:     ChapterStatePublish,
		Pages:     pages,
		UpdatedAt: now,
		CreatedAt: now,
	}, nil
}

func (c *Chapter) Updater() *ChapterUpdater {
	return &ChapterUpdater{c: c}
}

type ChapterUpdater struct {
	c    *Chapter
	opts []ChapterUpdateOption
}

type ChapterUpdateOption func(*Chapter) error

func (u *ChapterUpdater) Title(title *string) *ChapterUpdater {
	if title == nil {
		return u
	}
	u.opts = append(u.opts, func(c *Chapter) error {
		if err := validateTitle(title); err != nil {
			return err
		}
		c.Title = title
		return nil
	})
	return u
}

func (u *ChapterUpdater) Volume(volume *string) *ChapterUpdater {
	if volume == nil {
		return u
	}
	u.opts = append(u.opts, func(c *Chapter) error {
		if err := validateVolume(volume); err != nil {
			return err
		}
		c.Volume = volume
		return nil
	})
	return u
}

func (u *ChapterUpdater) Number(number *string) *ChapterUpdater {
	if number == nil {
		return u
	}
	u.opts = append(u.opts, func(c *Chapter) error {
		if err := validateNumber(*number); err != nil {
			return err
		}
		c.Number = *number
		return nil
	})
	return u
}

func (u *ChapterUpdater) State(state *ChapterState) *ChapterUpdater {
	if state == nil {
		return u
	}
	u.opts = append(u.opts, func(c *Chapter) error {
		c.State = *state
		return nil
	})
	return u
}

func (u *ChapterUpdater) Pages(pages *[]Page) *ChapterUpdater {
	if pages == nil {
		return u
	}
	u.opts = append(u.opts, func(c *Chapter) error {
		if err := validatePages(*pages); err != nil {
			if !errors.Is(err, ErrEmptyPages) || len(c.Pages) > 0 {
				return err
			}
		}
		c.Pages = *pages
		return nil
	})
	return u
}

func (u *ChapterUpdater) Apply() error {
	for _, opt := range u.opts {
		if err := opt(u.c); err != nil {
			return err
		}
	}
	u.c.UpdatedAt = time.Now()
	return nil
}

func validatePages(pages []Page) error {
	if len(pages) == 0 {
		return ErrEmptyPages.WithMessage("pages cannot be empty")
	}
	for _, page := range pages {
		if page.Width <= 0 {
			return ErrInvalidPageWidth.WithMessage("page width must be greater than zero")
		}
		if page.Height <= 0 {
			return ErrInvalidPageHeight.WithMessage("page height must be greater than zero")
		}
		if page.ObjectName == "" {
			return ErrEmptyPageObjectName.WithMessage("page object name cannot be empty")
		}
	}
	return nil
}

var numberRegex = `^(0|[1-9]\d*)(\.\d+)?([a-z]+)?$`

func validateNumber(number string) error {
	matched, err := regexp.MatchString(numberRegex, number)
	if err != nil {
		return err
	}
	if !matched {
		return ErrInvalidChapterNumber.
			WithMessage("must follow format: number, decimal, or number with letter suffix (e.g., 1, 1.5, 1a)").
			WithArg("value", number)
	}
	return nil
}
