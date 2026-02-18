package config

import "time"

type Config struct {
	App     AppConfig
	DB      DatabaseConfig
	HTTP    HTTPConfig
	JWT     JWTConfig
	Storage StorageConfig
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
	StorageType string

	// LocalStorage
	BasePath string
}
