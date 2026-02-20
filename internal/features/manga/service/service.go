package service

import (
	"log/slog"

	repo "github.com/mairuu/mp-api/internal/features/manga/repository"
	"github.com/mairuu/mp-api/internal/platform/authorization"
	"github.com/mairuu/mp-api/internal/platform/storage"
)

type Service struct {
	log             *slog.Logger
	repo            repo.Repository
	enforcer        *authorization.Enforcer
	publicBucket    storage.Bucket
	temporaryBucket storage.Bucket
	mapper          mapper
}

func NewService(log *slog.Logger, repo repo.Repository, enforcer *authorization.Enforcer, publicBucket storage.Bucket, temporaryBucket storage.Bucket) *Service {
	return &Service{
		log:             log,
		repo:            repo,
		enforcer:        enforcer,
		publicBucket:    publicBucket,
		temporaryBucket: temporaryBucket,
		mapper:          mapper{},
	}
}
