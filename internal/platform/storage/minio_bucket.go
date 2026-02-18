package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"iter"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioBucket struct {
	client     *minio.Client
	bucketName string
}

type MinioConfig struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
	UseSSL          bool
}

type minioObjectReader struct {
	*minio.Object
}

func (r *minioObjectReader) Seek(offset int64, whence int) (int64, error) {
	return r.Object.Seek(offset, whence)
}

func NewMinioBucket(cfg *MinioConfig) (*MinioBucket, error) {
	if cfg.Endpoint == "" {
		return nil, errors.New("minio endpoint is required")
	}
	if cfg.AccessKeyID == "" {
		return nil, errors.New("minio access key ID is required")
	}
	if cfg.SecretAccessKey == "" {
		return nil, errors.New("minio secret access key is required")
	}
	if cfg.BucketName == "" {
		return nil, errors.New("minio bucket name is required")
	}

	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	bucket := &MinioBucket{
		client:     client,
		bucketName: cfg.BucketName,
	}

	// ensure the bucket exists
	ctx := context.Background()
	exists, err := client.BucketExists(ctx, cfg.BucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		err = client.MakeBucket(ctx, cfg.BucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	return bucket, nil
}

// normalizeMetadataKeys converts all metadata keys to lowercase
// to ensure consistent behavior with local storage
func normalizeMetadataKeys(metadata map[string]string) map[string]string {
	if metadata == nil {
		return nil
	}
	normalized := make(map[string]string, len(metadata))
	for key, value := range metadata {
		normalized[strings.ToLower(key)] = value
	}
	return normalized
}

func (b *MinioBucket) Upload(ctx context.Context, objectName string, reader io.Reader, opts *UploadOptions) error {
	if reader == nil {
		return errors.New("reader cannot be nil")
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	putOpts := minio.PutObjectOptions{}

	if opts != nil {
		if opts.ContentType != "" {
			putOpts.ContentType = opts.ContentType
		} else {
			putOpts.ContentType = "application/octet-stream"
		}

		if opts.MetaData != nil {
			putOpts.UserMetadata = normalizeMetadataKeys(opts.MetaData)
		}
	} else {
		putOpts.ContentType = "application/octet-stream"
	}

	// upload with -1 size to stream data
	_, err := b.client.PutObject(ctx, b.bucketName, objectName, reader, -1, putOpts)
	if err != nil {
		return fmt.Errorf("failed to upload object: %w", err)
	}

	return nil
}

func (b *MinioBucket) Download(ctx context.Context, objectName string) (ObjectReader, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	obj, err := b.client.GetObject(ctx, b.bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}

	// verify the object exists by getting stat
	_, err = obj.Stat()
	if err != nil {
		obj.Close()
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return nil, ErrObjectNotFound
		}
		return nil, fmt.Errorf("failed to stat object: %w", err)
	}

	return &minioObjectReader{Object: obj}, nil
}

func (b *MinioBucket) Delete(ctx context.Context, objectName string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	err := b.client.RemoveObject(ctx, b.bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		errResp := minio.ToErrorResponse(err)
		if errResp.Code == "NoSuchKey" {
			return ErrObjectNotFound
		}
		return fmt.Errorf("failed to delete object: %w", err)
	}

	return nil
}

func (b *MinioBucket) GetMetadata(ctx context.Context, objectName string) (*ObjectMetadata, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	objInfo, err := b.client.StatObject(ctx, b.bucketName, objectName, minio.StatObjectOptions{})
	if err != nil {
		errResp := minio.ToErrorResponse(err)
		if errResp.Code == "NoSuchKey" {
			return nil, ErrObjectNotFound
		}
		return nil, fmt.Errorf("failed to get object metadata: %w", err)
	}

	metadata := &ObjectMetadata{
		Size:         objInfo.Size,
		ContentType:  objInfo.ContentType,
		LastModified: objInfo.LastModified,
		MetaData:     normalizeMetadataKeys(objInfo.UserMetadata),
	}

	// minio provides etag which is usually the md5 hash
	if objInfo.ETag != "" {
		metadata.Hash = objInfo.ETag
	}

	return metadata, nil
}

func (b *MinioBucket) List(ctx context.Context, prefix string) ([]string, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	var objects []string

	opts := minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	}

	for obj := range b.client.ListObjects(ctx, b.bucketName, opts) {
		if obj.Err != nil {
			return nil, fmt.Errorf("error listing objects: %w", obj.Err)
		}
		objects = append(objects, obj.Key)
	}

	return objects, nil
}

func (b *MinioBucket) ListIter(ctx context.Context, prefix string) iter.Seq[string] {
	return func(yield func(string) bool) {
		opts := minio.ListObjectsOptions{
			Prefix:    prefix,
			Recursive: true,
		}

		for obj := range b.client.ListObjects(ctx, b.bucketName, opts) {
			if obj.Err != nil {
				// can't propagate error in iterator, so we just stop
				return
			}
			if !yield(obj.Key) {
				return
			}
		}
	}
}
