package service

import (
	"context"
	"mime/multipart"
	"time"

	"github.com/google/uuid"
	"github.com/mairuu/mp-api/internal/app"
	"github.com/mairuu/mp-api/internal/features/bucket/model"
	"github.com/mairuu/mp-api/internal/platform/authorization"
	"github.com/mairuu/mp-api/internal/platform/storage"
)

type Service struct {
	enforcer      *authorization.Enforcer
	bucket        storage.Bucket
	cleanupTicker *time.Ticker
	cleanupDone   chan struct{}
}

func NewService(enforcer *authorization.Enforcer, bucket storage.Bucket) *Service {
	return &Service{
		enforcer:    enforcer,
		bucket:      bucket,
		cleanupDone: make(chan struct{}),
	}
}

func (s *Service) UploadFiles(ctx context.Context, ur *app.UserRole, files []*multipart.FileHeader, refIDs []string) (*UploadDTO, error) {
	if len(files) == 0 {
		return nil, model.ErrFileRequired.
			WithMessage("at least one file is required")
	}

	if len(refIDs) > 0 && len(refIDs) != len(files) {
		return nil, model.ErrRefIDCountMismatch.
			WithMessage("number of ref_ids must match number of files")
	}

	err := s.enforce(ur, model.ActionUpload, nil)
	if err != nil {
		return nil, err
	}

	// todo: quota check, e.g. max file size, max number of files, etc.

	acceptedFiles := make([]AcceptedFile, 0)
	rejectedFiles := make([]RejectedFile, 0)

	for i, file := range files {
		opts := &storage.UploadOptions{
			MetaData: map[string]string{
				"user_id": ur.ID.String(),
			},
		}

		f, err := file.Open()
		if err != nil {
			rf := RejectedFile{
				OriginalFileName: file.Filename,
				Error:            err.Error(), // todo: do not expose internal error message
			}
			if len(refIDs) > 0 {
				rf.RefID = &refIDs[i]
			}
			rejectedFiles = append(rejectedFiles, rf)
			continue
		}

		objectName := uuid.New().String()
		if err := s.bucket.Upload(ctx, objectName, f, opts); err != nil {
			f.Close()
			rf := RejectedFile{
				OriginalFileName: file.Filename,
				Error:            err.Error(), // todo: do not expose internal error message
			}
			if len(refIDs) > 0 {
				rf.RefID = &refIDs[i]
			}
			rejectedFiles = append(rejectedFiles, rf)
			continue
		}
		f.Close()

		af := AcceptedFile{
			OriginalFileName: file.Filename,
			ObjectName:       objectName,
		}
		if len(refIDs) > 0 {
			af.RefID = &refIDs[i]
		}
		acceptedFiles = append(acceptedFiles, af)
	}

	return &UploadDTO{
		AcceptedFiles: acceptedFiles,
		RejectedFiles: rejectedFiles,
	}, nil
}

func (s *Service) GetMetadata(ctx context.Context, objectName string) (*storage.ObjectMetadata, error) {
	return s.bucket.GetMetadata(ctx, objectName)
}

func (s *Service) Download(ctx context.Context, objectName string) (storage.ObjectReader, error) {
	return s.bucket.Download(ctx, objectName)
}

func (s *Service) StartCleanup(interval time.Duration, ttl time.Duration) {
	if s.cleanupTicker != nil {
		return // already running
	}
	s.cleanupTicker = time.NewTicker(interval)

	go func() {
		defer s.cleanupTicker.Stop()

		s.cleanupTemporaryFiles(ttl)

		for {
			select {
			case <-s.cleanupTicker.C:
				s.cleanupTemporaryFiles(ttl)
			case <-s.cleanupDone:
				return
			}
		}
	}()
}

func (s *Service) cleanupTemporaryFiles(ttl time.Duration) {
	ctx := context.Background()

	for objectName := range s.bucket.ListIter(ctx, "") {
		meta, err := s.bucket.GetMetadata(ctx, objectName)
		if err != nil {
			continue
		}

		if time.Since(meta.LastModified) > ttl {
			if err := s.bucket.Delete(ctx, objectName); err != nil {
				continue
			}
		}
	}
}

func (s *Service) StopCleanup() {
	if s.cleanupDone != nil {
		close(s.cleanupDone)
	}
}

func (s *Service) enforce(ur *app.UserRole, action authorization.Action, target authorization.ScopeResolvable) error {
	return s.enforcer.Enforce(ur.ID, ur.Role, model.ResourceBucket, action, target)
}
