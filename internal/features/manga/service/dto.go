package service

// manga

type CreateMangaDTO struct {
	Title    string `json:"title" binding:"required"`
	Synopsis string `json:"synopsis" binding:""`
	Status   string `json:"status" binding:"required"`
}

type UpdateMangaDTO struct {
	Title     *string              `json:"title"`
	Synopsis  *string              `json:"synopsis"`
	Status    *string              `json:"status"`
	CoverArts *[]UpdateCoverArtDTO `json:"covers"`
}

type UpdateCoverArtDTO struct {
	ObjectName  string `json:"object_name" binding:"required"`
	Volume      string `json:"volume" binding:"required"`
	Description string `json:"description"`
}

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
	Description string `json:"description"`
	ObjectName  string `json:"object_name"`
}

type MangaSummaryDTO struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// chapter

type CreateChapterDTO struct {
	MangaID string  `json:"manga_id" binding:"required,uuid"`
	Title   string  `json:"title" binding:"required"`
	Volume  *string `json:"volume" binding:""`
	Number  string  `json:"number" binding:"required"`
}

type UpdateChapterDTO struct {
	Title  *string `json:"title"`
	Volume *string `json:"volume"`
	Number *string `json:"number"`
}

type ChapterDTO struct {
	ID      string  `json:"id"`
	MangaID string  `json:"manga_id"`
	Title   string  `json:"title"`
	Volume  *string `json:"volume"`
	Number  string  `json:"number"`
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
