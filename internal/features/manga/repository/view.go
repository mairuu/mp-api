package repository

import "github.com/google/uuid"

type MangaSummary struct {
	ID    uuid.UUID
	Title string
}
