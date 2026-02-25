package service

// manga

type CreateMangaDTO struct {
	Title    string              `json:"title" binding:"required"`
	Synopsis string              `json:"synopsis"`
	Status   string              `json:"status" binding:"required"`
	Covers   []CreateCoverArtDTO `json:"covers" binding:"dive"`
}

type CreateCoverArtDTO struct {
	Volume      string `json:"volume"`
	IsPrimary   bool   `json:"is_primary"`
	ObjectName  string `json:"object_name" binding:"required"`
	Description string `json:"description"`
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
	Volume      string `json:"volume"`
	IsPrimary   bool   `json:"is_primary"`
	Description string `json:"description"`
	ObjectName  string `json:"object_name"`
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
	ID      string  `json:"id"`
	MangaID string  `json:"manga_id"`
	Title   string  `json:"title"`
	Volume  *string `json:"volume"`
	Number  string  `json:"number"`
}

// pagination

type PagedDTO struct {
	Items      any           `json:"items"`
	Pagination PaginationDTO `json:"pagination"`
}

type PaginationDTO struct {
	TotalItems int `json:"total_items"`
	TotalPages int `json:"total_pages"`
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
}

func NewPagedDTO(total, totalPage, pageSize, page int, items any) PagedDTO {
	return PagedDTO{
		Items: items,
		Pagination: PaginationDTO{
			TotalItems: total,
			TotalPages: totalPage,
			PageSize:   pageSize,
			Page:       page,
		},
	}
}
