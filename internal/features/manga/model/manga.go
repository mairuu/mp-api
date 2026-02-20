package model

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Manga struct {
	ID        uuid.UUID
	OwnerID   uuid.UUID
	Title     string
	Synopsis  string
	Status    MangaStatus
	State     MangaState
	Covers    []CoverArt
	UpdatedAt time.Time
	CreatedAt time.Time
}

type CoverArt struct {
	Volume      string // unique per manga
	ObjectName  string
	Description string
}

type MangaStatus string

const (
	MangaStatusDraft     MangaStatus = "draft"
	MangaStatusOngoing   MangaStatus = "ongoing"
	MangaStatusCompleted MangaStatus = "completed"
	MangaStatusHiatus    MangaStatus = "hiatus"
	MangaStatusCancelled MangaStatus = "cancelled"
)

type MangaState string

const (
	MangaStateDraft   MangaState = "draft"
	MangaStatePublish MangaState = "published"
)

func NewManga(ownerID uuid.UUID, title, synopsis string, status MangaStatus) (*Manga, error) {
	if err := validateTitle(title); err != nil {
		return nil, err
	}
	if err := validateSynopsis(synopsis); err != nil {
		return nil, err
	}
	if err := validateStatus(status); err != nil {
		return nil, err
	}

	now := time.Now()

	return &Manga{
		ID:        uuid.New(),
		OwnerID:   ownerID,
		Title:     title,
		Synopsis:  synopsis,
		Status:    status,
		State:     MangaStateDraft,
		UpdatedAt: now,
		CreatedAt: now,
	}, nil
}

func NewCoverArt(volume, objectName, description string) (*CoverArt, error) {
	if err := validateVolume(&volume); err != nil {
		return nil, err
	}
	return &CoverArt{
		Volume:      volume,
		ObjectName:  objectName,
		Description: description,
	}, nil
}

type MangaUpdater struct {
	m    *Manga
	opts []MangaUpdateOption
}

type MangaUpdateOption func(*Manga) error

func (m *Manga) Updater() *MangaUpdater {
	return &MangaUpdater{m: m}
}

func (u *MangaUpdater) Title(title *string) *MangaUpdater {
	if title == nil {
		return u
	}

	u.opts = append(u.opts, func(m *Manga) error {
		t := strings.TrimSpace(*title)
		if err := validateTitle(t); err != nil {
			return err
		}
		m.Title = t
		return nil
	})

	return u
}

func (u *MangaUpdater) Synopsis(synopsis *string) *MangaUpdater {
	if synopsis == nil {
		return u
	}

	u.opts = append(u.opts, func(m *Manga) error {

		s := strings.TrimSpace(*synopsis)
		if err := validateSynopsis(s); err != nil {
			return err
		}
		m.Synopsis = s
		return nil
	})

	return u
}

func (u *MangaUpdater) Status(status *MangaStatus) *MangaUpdater {
	if status == nil {
		return u
	}

	u.opts = append(u.opts, func(m *Manga) error {
		if err := validateStatus(*status); err != nil {
			return err
		}
		m.Status = *status
		return nil
	})

	return u
}

func (u *MangaUpdater) State(state *MangaState) *MangaUpdater {
	if state == nil {
		return u
	}

	u.opts = append(u.opts, func(m *Manga) error {
		m.State = *state
		return nil
	})

	return u
}

func (u *MangaUpdater) CoverArts(covers []CoverArt) *MangaUpdater {
	if covers == nil {
		return u
	}

	u.opts = append(u.opts, func(m *Manga) error {
		uniqueVolumes := make(map[string]bool)
		for _, cover := range covers {
			if err := validateVolume(&cover.Volume); err != nil {
				return err
			}
			if uniqueVolumes[cover.Volume] {
				return fmt.Errorf("%w: volume = %s", ErrDuplicateVolume, cover.Volume)
			}
			uniqueVolumes[cover.Volume] = true
		}

		m.Covers = covers
		return nil
	})

	return u
}

// Apply applies the updates to the manga.
// any error in the update options will be rolled back and the manga will not be updated.
func (u *MangaUpdater) Apply() error {
	m := *u.m
	for _, opt := range u.opts {
		if err := opt(&m); err != nil {
			return err
		}
	}
	*u.m = m
	u.m.UpdatedAt = time.Now()
	return nil
}

func (status MangaStatus) IsValid() bool {
	switch status {
	case MangaStatusDraft, MangaStatusOngoing, MangaStatusCompleted, MangaStatusHiatus, MangaStatusCancelled:
		return true
	default:
		return false
	}
}

func (state MangaState) IsValid() bool {
	switch state {
	case MangaStateDraft, MangaStatePublish:
		return true
	default:
		return false
	}
}

func validateTitle(title string) error {
	if title == "" {
		return ErrInvalidTitle
	}
	return nil
}

func validateSynopsis(_ string) error {
	return nil
}

func validateStatus(status MangaStatus) error {
	if !status.IsValid() {
		return ErrInvalidStatus
	}
	return nil
}

var volumeRegex = `^(0|[1-9]\d*)(\.\d+)?([a-z]+)?$`

func validateVolume(volume *string) error {
	if volume == nil {
		return nil
	}
	matched, err := regexp.MatchString(volumeRegex, *volume)
	if err != nil {
		return ErrInvalidVolume
	}
	if !matched {
		return ErrInvalidVolume
	}
	return nil
}
