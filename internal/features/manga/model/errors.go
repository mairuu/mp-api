package model

import "errors"

var (
	ErrMangaNotFound          = errors.New("manga not found")
	ErrMangaAlreadyExists     = errors.New("manga already exists")
	ErrInvalidTitle           = errors.New("title cannot be empty")
	ErrInvalidStatus          = errors.New("status must be one of: ongoing, completed, hiatus, cancelled")
	ErrInvalidVolume          = errors.New("volume must follow format: number, decimal, or number with letter suffix (e.g., 1, 1.5, 1a)")
	ErrDuplicateVolume        = errors.New("volume already exists for this manga")
	ErrCoverNotFound          = errors.New("cover art not found")
	ErrUnsupportedImageFormat = errors.New("image format not supported")
	ErrEmptyPages             = errors.New("pages cannot be empty")
	ErrInvalidPageWidth       = errors.New("page width must be greater than zero")
	ErrInvalidPageHeight      = errors.New("page height must be greater than zero")
	ErrEmptyPageObjectName    = errors.New("page object name cannot be empty")
)
