package config

import "time"

type Config struct {
	App     AppConfig
	DB      DatabaseConfig
	HTTP    HTTPConfig
	JWT     JWTConfig
	Storage StorageConfig
	Cleanup CleanupConfig
}

type AppConfig struct {
	LogLevel string
}

type DatabaseConfig struct {
	DSN      string
	LogLevel string
}

type HTTPConfig struct {
	Addr string
}

type JWTConfig struct {
	Secret         []byte
	AccessTokenTTL time.Duration
}

type StorageConfig struct {
	PublicBucket    BucketConfig
	TemporaryBucket BucketConfig
}

type BucketConfig struct {
	StorageType string

	// localstorage
	BasePath string

	// minio
	MinIOEndpoint        string
	MinIOAccessKeyID     string
	MinIOSecretAccessKey string
	MinIOBucketName      string
	MinIOUseSSL          bool
}

type CleanupConfig struct {
	Interval time.Duration
	TTL      time.Duration
}
