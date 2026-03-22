package service

// manga

type CreateMangaDTO struct {
	Title    string              `json:"title" binding:"required"`
	Synopsis string              `json:"synopsis"`
	Status   string              `json:"status" binding:"required"`
	Covers   []CreateCoverArtDTO `json:"covers" binding:"dive"`
}

type CreateCoverArtDTO struct {
	ObjectName  string  `json:"object_name" binding:"required"`
	Volume      *string `json:"volume"`
	IsPrimary   *bool   `json:"is_primary"`
	Description *string `json:"description"`
}

type UpdateMangaDTO struct {
	Title    *string              `json:"title"`
	Synopsis *string              `json:"synopsis"`
	Status   *string              `json:"status"`
	Covers   *[]UpdateCoverArtDTO `json:"covers"`
}

type UpdateCoverArtDTO = CreateCoverArtDTO

type MangaDTO struct {
	ID        string        `json:"id"`
	Title     string        `json:"title"`
	Synopsis  string        `json:"synopsis"`
	Status    string        `json:"status"`
	State     string        `json:"state"`
	CoverArts []CoverArtDTO `json:"covers"`
}

type CoverArtDTO struct {
	ObjectName  string  `json:"object_name"`
	IsPrimary   bool    `json:"is_primary"`
	Volume      *string `json:"volume"`
	Description *string `json:"description"`
}

type MangaSummaryDTO struct {
	ID              string  `json:"id"`
	Title           string  `json:"title"`
	CoverObjectName *string `json:"cover_object_name"`
}

// chapter

type CreateChapterDTO struct {
	MangaID string   `json:"manga_id" binding:"required,uuid"`
	Number  string   `json:"number" binding:"required"`
	Title   *string  `json:"title" binding:""`
	Volume  *string  `json:"volume" binding:""`
	Pages   []string `json:"pages" binding:"required,dive"`
}

type UpdateChapterDTO struct {
	Title  *string   `json:"title"`
	Volume *string   `json:"volume"`
	Number *string   `json:"number"`
	Pages  *[]string `json:"pages"` // list of page object names
}

type ChapterDTO struct {
	ID      string    `json:"id"`
	MangaID string    `json:"manga_id"`
	Number  string    `json:"number"`
	Title   *string   `json:"title"`
	Volume  *string   `json:"volume"`
	Pages   []PageDTO `json:"pages"`
}

type PageDTO struct {
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	ObjectName string `json:"object_name"`
}

type ChapterSummaryDTO struct {
	ID        string  `json:"id"`
	MangaID   string  `json:"manga_id"`
	Number    string  `json:"number"`
	Title     *string `json:"title"`
	Volume    *string `json:"volume"`
	CreatedAt string  `json:"created_at"`
}
