package config

import (
	"os"
	"time"

	"github.com/joho/godotenv"
)

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := Config{}

	cfg.App = AppConfig{
		LogLevel: getEnv("APP_LOG_LEVEL", "debug"),
	}

	cfg.DB = DatabaseConfig{
		DSN:      getEnv("DB_DSN", "host=localhost user=postgres password=postgres dbname=app_db port=5432 sslmode=disable"),
		LogLevel: getEnv("DB_LOG_LEVEL", "info"),
	}

	cfg.HTTP = HTTPConfig{
		Addr: getEnv("HTTP_ADDR", ":8080"),
	}

	cfg.JWT = JWTConfig{
		Secret:         []byte(getEnv("JWT_SECRET", "your-secret-key-change-this-in-production")),
		AccessTokenTTL: getEnvDuration("JWT_ACCESS_TOKEN_TTL", 24*time.Hour),
	}

	cfg.Storage = StorageConfig{
		PublicBucket: BucketConfig{
			StorageType:          getEnv("PUBLIC_STORAGE_TYPE", "local"),
			BasePath:             getEnv("PUBLIC_STORAGE_BASE_PATH", "run/upload/public"),
			MinIOEndpoint:        getEnv("PUBLIC_MINIO_ENDPOINT", "localhost:9000"),
			MinIOAccessKeyID:     getEnv("PUBLIC_MINIO_ACCESS_KEY_ID", ""),
			MinIOSecretAccessKey: getEnv("PUBLIC_MINIO_SECRET_ACCESS_KEY", ""),
			MinIOBucketName:      getEnv("PUBLIC_MINIO_BUCKET_NAME", "mp-api-public"),
			MinIOUseSSL:          getEnvBool("PUBLIC_MINIO_USE_SSL", false),
		},
		TemporaryBucket: BucketConfig{
			StorageType:          getEnv("TEMPORARY_STORAGE_TYPE", "local"),
			BasePath:             getEnv("TEMPORARY_STORAGE_BASE_PATH", "run/upload/temp"),
			MinIOEndpoint:        getEnv("TEMPORARY_MINIO_ENDPOINT", "localhost:9000"),
			MinIOAccessKeyID:     getEnv("TEMPORARY_MINIO_ACCESS_KEY_ID", ""),
			MinIOSecretAccessKey: getEnv("TEMPORARY_MINIO_SECRET_ACCESS_KEY", ""),
			MinIOBucketName:      getEnv("TEMPORARY_MINIO_BUCKET_NAME", "mp-api-temp"),
			MinIOUseSSL:          getEnvBool("TEMPORARY_MINIO_USE_SSL", false),
		},
	}

	cfg.Cleanup = CleanupConfig{
		Interval: getEnvDuration("CLEANUP_INTERVAL", 1*time.Hour),
		TTL:      getEnvDuration("CLEANUP_TTL", 24*time.Hour),
	}

	return &cfg, nil
}

func getEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	value, err := time.ParseDuration(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

func getEnvBool(key string, defaultValue bool) bool {
	valueStr, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return valueStr == "true" || valueStr == "1" || valueStr == "yes"
}
