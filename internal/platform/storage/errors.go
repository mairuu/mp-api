package storage

import (
	"errors"
)

var (
	ErrObjectNotFound         = errors.New("object not found")
	ErrUnsupportedStorageType = errors.New("unsupported storage type")
)
