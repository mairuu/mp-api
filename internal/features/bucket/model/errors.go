package model

import "github.com/mairuu/mp-api/internal/platform/errors"

var (
	ErrFileRequired       = errors.New("file_required")
	ErrRefIDCountMismatch = errors.New("ref_id_count_mismatch")
)
