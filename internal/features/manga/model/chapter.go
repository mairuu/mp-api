package model

import (
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
)

type Chapter struct {
	ID        uuid.UUID
	MangaID   uuid.UUID
	Title     string
	Volume    *string
	Number    string
	Pages     []Page
	State     ChapterState
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

func NewChapter(mangaID uuid.UUID, title, number string, volume *string) (*Chapter, error) {
	if err := validateTitle(title); err != nil {
		return nil, err
	}
	if err := validateVolume(volume); err != nil {
		return nil, err
	}
	if err := validateNumber(number); err != nil {
		return nil, err
	}
	now := time.Now()
	return &Chapter{
		ID:        uuid.New(),
		MangaID:   mangaID,
		Title:     title,
		Volume:    volume,
		Number:    number,
		State:     ChapterStateDraft,
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
		if err := validateTitle(*title); err != nil {
			return err
		}
		c.Title = *title
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

func (u *ChapterUpdater) Pages(pages *[]Page) *ChapterUpdater {
	if pages == nil {
		return u
	}
	u.opts = append(u.opts, func(c *Chapter) error {
		if err := validatePages(*pages); err != nil {
			return err
		}
		c.Pages = *pages
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
		return fmt.Errorf("%w; pages cannot be empty", ErrEmptyPages)
	}
	for _, page := range pages {
		if page.Width <= 0 {
			return fmt.Errorf("%w; page width must be greater than zero", ErrInvalidPageWidth)
		}
		if page.Height <= 0 {
			return fmt.Errorf("%w; page height must be greater than zero", ErrInvalidPageHeight)
		}
		if page.ObjectName == "" {
			return fmt.Errorf("%w; page object name cannot be empty", ErrEmptyPageObjectName)
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
		return fmt.Errorf("%w; must follow format: number, decimal, or number with letter suffix (e.g., 1, 1.5, 1a)", ErrInvalidChapterNumber)
	}
	return nil
}
