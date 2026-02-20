package model

import "github.com/mairuu/mp-api/internal/platform/errors"

var (
	ErrMangaNotFound          = errors.New("manga_not_found")
	ErrMangaAlreadyExists     = errors.New("manga_already_exists")
	ErrInvalidTitle           = errors.New("invalid_title")
	ErrInvalidStatus          = errors.New("invalid_status")
	ErrInvalidVolume          = errors.New("invalid_volume")
	ErrChapterNotFound        = errors.New("chapter_not_found")
	ErrChapterAlreadyExists   = errors.New("chapter_already_exists")
	ErrInvalidChapterNumber   = errors.New("invalid_chapter_number")
	ErrVolumeAlreadyExists    = errors.New("volume_already_exists")
	ErrCoverNotFound          = errors.New("cover_not_found")
	ErrUnsupportedImageFormat = errors.New("unsupported_image_format")
	ErrEmptyPages             = errors.New("empty_pages")
	ErrPageNotFound           = errors.New("page_not_found")
	ErrInvalidPageWidth       = errors.New("invalid_page_width")
	ErrInvalidPageHeight      = errors.New("invalid_page_height")
	ErrEmptyPageObjectName    = errors.New("empty_page_object_name")
)
