package storage

import (
	"context"
	"io"
	"iter"
	"time"
)

type UploadOptions struct {
	ContentType string
	MetaData    map[string]string
}

type ObjectMetadata struct {
	Hash         string
	Size         int64
	ContentType  string
	LastModified time.Time
	MetaData     map[string]string // normalized metadata with lowercase keys
}

type ObjectReader interface {
	io.ReadCloser
	io.Seeker
}

// Bucket defines the storage behavior
type Bucket interface {
	// Upload streams data from the reader to the storage
	// uploading an existing object overwrites it
	Upload(ctx context.Context, objectName string, reader io.Reader, opts *UploadOptions) error

	// Download returns a ObjectReader to read the object data
	Download(ctx context.Context, objectName string) (ObjectReader, error)

	// Delete removes the object
	Delete(ctx context.Context, objectName string) error

	// GetMetadata returns information without downloading the file
	GetMetadata(ctx context.Context, objectName string) (*ObjectMetadata, error)

	// List returns a list of object names with the given prefix
	// empty prefix lists all objects
	List(ctx context.Context, prefix string) ([]string, error)

	// ListIter returns an iterator to list object names with the given prefix
	// empty prefix lists all objects
	ListIter(ctx context.Context, prefix string) iter.Seq[string]
}
