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
		StorageType:          getEnv("STORAGE_TYPE", "local"),
		BasePath:             getEnv("STORAGE_BASE_PATH", "run/upload"),
		MinIOEndpoint:        getEnv("MINIO_ENDPOINT", "localhost:9000"),
		MinIOAccessKeyID:     getEnv("MINIO_ACCESS_KEY_ID", ""),
		MinIOSecretAccessKey: getEnv("MINIO_SECRET_ACCESS_KEY", ""),
		MinIOBucketName:      getEnv("MINIO_BUCKET_NAME", "mp-api"),
		MinIOUseSSL:          getEnvBool("MINIO_USE_SSL", false),
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
