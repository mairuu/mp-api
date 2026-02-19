package storage

import "github.com/mairuu/mp-api/internal/platform/config"

func NewBucket(cfg *config.BucketConfig) (Bucket, error) {
	switch cfg.StorageType {
	case "local":
		return NewLocalBucket(cfg.BasePath)
	case "minio":
		minioCfg := &MinioConfig{
			Endpoint:        cfg.MinIOEndpoint,
			AccessKeyID:     cfg.MinIOAccessKeyID,
			SecretAccessKey: cfg.MinIOSecretAccessKey,
			BucketName:      cfg.MinIOBucketName,
			UseSSL:          cfg.MinIOUseSSL,
		}
		return NewMinioBucket(minioCfg)
	default:
		return nil, ErrUnsupportedStorageType
	}
}
