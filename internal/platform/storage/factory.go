package storage

import "github.com/mairuu/mp-api/internal/platform/config"

func NewBucket(cfg *config.StorageConfig) (Bucket, error) {
	switch cfg.StorageType {
	case "local":
		return NewLocalBucket(cfg.BasePath)
	default:
		return nil, ErrUnsupportedStorageType
	}
}
