package model

import (
	"time"

	"github.com/google/uuid"
)

// CREATE TABLE library_manga (
//     id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
//     user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
//     manga_id    UUID NOT NULL REFERENCES manga(id),
//     tags        TEXT[] NOT NULL DEFAULT '{}',
//     added_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
//     UNIQUE (user_id, manga_id)
// );

type Library struct {
	OwnerID uuid.UUID      `json:"owner_id"`
	Mangas  []LibraryManga `json:"mangas"`
}

type LibraryManga struct {
	MangaID uuid.UUID `json:"manga_id"`
	Tags    []string  `json:"tags"`
	AddedAt time.Time `json:"added_at"`
}

func NewLibrary(ownerID uuid.UUID) *Library {
	return &Library{
		OwnerID: ownerID,
		Mangas:  []LibraryManga{},
	}
}

func (l *Library) AllTags() []string {
	tagSet := make(map[string]struct{})
	for _, manga := range l.Mangas {
		for _, tag := range manga.Tags {
			tagSet[tag] = struct{}{}
		}
	}
	tags := make([]string, 0, len(tagSet))
	for tag := range tagSet {
		tags = append(tags, tag)
	}
	return tags
}

func (l *Library) UpsertManga(mangaID uuid.UUID, tags []string) {
	for i, manga := range l.Mangas {
		if manga.MangaID == mangaID {
			l.Mangas[i].Tags = tags
			return
		}
	}
	l.Mangas = append(l.Mangas, LibraryManga{
		MangaID: mangaID,
		Tags:    tags,
		AddedAt: time.Now(),
	})
}

func (l *Library) RemoveManga(mangaID uuid.UUID) {
	for i, manga := range l.Mangas {
		if manga.MangaID == mangaID {
			l.Mangas = append(l.Mangas[:i], l.Mangas[i+1:]...)
			return
		}
	}
}

func (l *Library) UpdateMangaTags(mangaID uuid.UUID, tags []string) {
	for i, manga := range l.Mangas {
		if manga.MangaID == mangaID {
			l.Mangas[i].Tags = tags
			return
		}
	}
}

func (l *Library) GetManga(mangaID uuid.UUID) (*LibraryManga, bool) {
	for i, manga := range l.Mangas {
		if manga.MangaID == mangaID {
			return &l.Mangas[i], true
		}
	}
	return nil, false
}
