package service

import "github.com/google/uuid"

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

// - to add cover art object_name is required,
// object_name is a handle to staging object in storage
//
// - to update cover art the id is required, if object_name is provided,
// it will be treated as new cover art and staged, otherwise the existing cover art will be kept
//
// - to delete cover art simply omit the cover art from the update DTO,
// the service will determine which cover arts to delete based on the existing cover arts of the manga
type UpdateCoverArtDTO struct {
	ID *uuid.UUID `json:"id"`

	ObjectName *string `json:"object_name"`

	Volume      string `json:"volume"`
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
	ID          string `json:"id"`
	Volume      string `json:"volume"`
	Description string `json:"description"`
	ObjectName  string `json:"object_name"`
}

type MangaSummaryDTO struct {
	ID    string `json:"id"`
	Title string `json:"title"`
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
