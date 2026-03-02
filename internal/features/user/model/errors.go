package model

import "github.com/mairuu/mp-api/internal/platform/errors"

var (
	ErrUserNotFound       = errors.New("user_not_found")
	ErrUserAlreadyExists  = errors.New("user_already_exists")
	ErrInvalidUsername    = errors.New("invalid_username")
	ErrInvalidEmail       = errors.New("invalid_email")
	ErrInvalidPassword    = errors.New("invalid_password")
	ErrInvalidCredentials = errors.New("invalid_credentials")
	ErrInvalidRole        = errors.New("invalid_role")

	ErrRefreshTokenNotFound = errors.New("refresh_token_not_found")
	ErrRefreshTokenExpired  = errors.New("refresh_token_expired")
	ErrRefreshTokenRevoked  = errors.New("refresh_token_revoked")
)
