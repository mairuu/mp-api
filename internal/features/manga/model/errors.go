package model

import "errors"

var (
	ErrMangaNotFound          = errors.New("manga not found")
	ErrMangaAlreadyExists     = errors.New("manga already exists")
	ErrInvalidTitle           = errors.New("invalid title")
	ErrInvalidStatus          = errors.New("invalid status")
	ErrInvalidVolume          = errors.New("invalid volume")
	ErrChapterAlreadyExists   = errors.New("chapter already exists")
	ErrInvalidChapterNumber   = errors.New("invalid chapter number")
	ErrVolumeAlreadyExists    = errors.New("volume already exists")
	ErrCoverNotFound          = errors.New("cover art not found")
	ErrUnsupportedImageFormat = errors.New("image format not supported")
	ErrEmptyPages             = errors.New("empty pages")
	ErrInvalidPageWidth       = errors.New("invalid page width")
	ErrInvalidPageHeight      = errors.New("invalid page height")
	ErrEmptyPageObjectName    = errors.New("empty page object name")
)
