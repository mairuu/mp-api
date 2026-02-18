package storage

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"hash/adler32"
	"io"
	"iter"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type LocalBucket struct {
	basePath string
}

type LocalMetadata struct {
	Hash        string            `json:"hash"`
	ContentType string            `json:"content_type,omitempty"`
	MetaData    map[string]string `json:"__metadata,omitempty"`
}

var (
	ErrPathTraversal = errors.New("path traversal detected")
)

func NewLocalBucket(basePath string) (*LocalBucket, error) {
	absPath, err := filepath.Abs(basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve base path: %w", err)
	}

	if err := os.MkdirAll(absPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	return &LocalBucket{
		basePath: absPath,
	}, nil
}

func (b *LocalBucket) Upload(ctx context.Context, objectName string, reader io.Reader, opts *UploadOptions) error {
	if reader == nil {
		return errors.New("reader cannot be nil")
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	fullPath, err := b.objectPath(objectName)
	if err != nil {
		return fmt.Errorf("invalid object name: %w", err)
	}

	hasher := adler32.New()
	tee := io.TeeReader(reader, hasher)

	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	dst, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	_, err = io.Copy(dst, tee)
	if err != nil {
		dst.Close()
		os.Remove(fullPath) // clean up partial file
		return fmt.Errorf("failed to write file: %w", err)
	}

	// ensure data is flushed to disk
	if err := dst.Sync(); err != nil {
		os.Remove(fullPath)
		return fmt.Errorf("failed to sync file: %w", err)
	}

	meta := LocalMetadata{
		Hash: hex.EncodeToString(hasher.Sum(nil)),
	}

	if opts != nil {
		meta.MetaData = opts.MetaData

		if opts.ContentType != "" {
			meta.ContentType = opts.ContentType
		} else {
			meta.ContentType = "application/octet-stream"
		}
	}

	if err := b.writeMeta(objectName, &meta); err != nil {
		os.Remove(fullPath)
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	return nil
}

func (b *LocalBucket) Download(ctx context.Context, objectName string) (ObjectReader, error) {
	// check context before starting
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	fullPath, err := b.objectPath(objectName)
	if err != nil {
		return nil, fmt.Errorf("invalid object name: %w", err)
	}

	file, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrObjectNotFound
		}
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	return file, nil
}

func (b *LocalBucket) Delete(ctx context.Context, objectName string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	fullPath, err := b.objectPath(objectName)
	if err != nil {
		return fmt.Errorf("invalid object name: %w", err)
	}

	metaPath, err := b.metaPath(objectName)
	if err != nil {
		return fmt.Errorf("invalid object name: %w", err)
	}

	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete object: %w", err)
	}

	os.Remove(metaPath)

	return nil
}

func (b *LocalBucket) GetMetadata(ctx context.Context, objectName string) (*ObjectMetadata, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	fullPath, err := b.objectPath(objectName)
	if err != nil {
		return nil, fmt.Errorf("invalid object name: %w", err)
	}

	stat, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrObjectNotFound
		}
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	meta := &ObjectMetadata{
		Size:         stat.Size(),
		LastModified: stat.ModTime(),
		ContentType:  "application/octet-stream",
	}

	localMeta := b.readMeta(objectName)
	if localMeta != nil {
		meta.Hash = localMeta.Hash
		if localMeta.ContentType != "" {
			meta.ContentType = localMeta.ContentType
		}
		meta.MetaData = localMeta.MetaData
	}

	return meta, nil
}

func (b *LocalBucket) List(ctx context.Context, prefix string) ([]string, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	it := b.ListIter(ctx, prefix)
	return slices.Collect(it), nil
}

func (b *LocalBucket) ListIter(ctx context.Context, prefix string) iter.Seq[string] {
	return func(yield func(string) bool) {
		// check context before starting
		select {
		case <-ctx.Done():
			return
		default:
		}

		// determine the starting path for walking
		var walkRoot string
		if prefix == "" {
			walkRoot = b.basePath
		} else {
			prefixPath, err := b.objectPath(prefix)
			if err != nil {
				return
			}

			// check if prefix is a directory or file prefix
			stat, err := os.Stat(prefixPath)
			if err == nil && stat.IsDir() {
				walkRoot = prefixPath
			} else {
				// prefix might be a partial file name, walk from parent dir
				walkRoot = filepath.Dir(prefixPath)
			}
		}

		// walk the directory tree
		filepath.WalkDir(walkRoot, func(path string, d os.DirEntry, err error) error {
			// check context during iteration
			select {
			case <-ctx.Done():
				return filepath.SkipAll
			default:
			}

			if err != nil {
				return nil // skip errors and continue
			}

			// skip directories and metadata files
			if d.IsDir() || strings.HasSuffix(path, ".meta") {
				return nil
			}

			// convert absolute path to object name
			relPath, err := filepath.Rel(b.basePath, path)
			if err != nil {
				return nil
			}

			// normalize path separators
			objectName := filepath.ToSlash(relPath)

			// filter by prefix if specified
			if prefix != "" && !strings.HasPrefix(objectName, prefix) {
				return nil
			}

			if !yield(objectName) {
				return filepath.SkipAll
			}

			return nil
		})
	}
}

func (b *LocalBucket) writeMeta(objectName string, meta *LocalMetadata) error {
	metaPath, err := b.metaPath(objectName)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(metaPath), 0755); err != nil {
		return fmt.Errorf("failed to create metadata directory: %w", err)
	}

	metaFile, err := os.Create(metaPath)
	if err != nil {
		return fmt.Errorf("failed to create metadata file: %w", err)
	}
	defer metaFile.Close()

	if err := json.NewEncoder(metaFile).Encode(meta); err != nil {
		return fmt.Errorf("failed to encode metadata: %w", err)
	}

	return metaFile.Sync()
}

func (b *LocalBucket) readMeta(objectName string) *LocalMetadata {
	metaPath, err := b.metaPath(objectName)
	if err != nil {
		return nil
	}

	metaFile, err := os.Open(metaPath)
	if err != nil {
		return nil
	}
	defer metaFile.Close()

	var meta LocalMetadata
	if err := json.NewDecoder(metaFile).Decode(&meta); err != nil {
		return nil
	}
	return &meta
}

func (b *LocalBucket) objectPath(objectName string) (string, error) {
	// clean the object name to prevent path traversal
	cleanName := filepath.Clean(objectName)

	// prevent absolute paths and parent directory references
	if filepath.IsAbs(cleanName) || strings.HasPrefix(cleanName, "..") || strings.Contains(cleanName, "/../") {
		return "", ErrPathTraversal
	}

	fullPath := filepath.Join(b.basePath, cleanName)

	// ensure the resolved path is within basepath
	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve path: %w", err)
	}

	if !strings.HasPrefix(absPath, b.basePath) {
		return "", ErrPathTraversal
	}

	return absPath, nil
}

func (b *LocalBucket) metaPath(objectName string) (string, error) {
	objPath, err := b.objectPath(objectName)
	if err != nil {
		return "", err
	}
	return objPath + ".meta", nil
}
