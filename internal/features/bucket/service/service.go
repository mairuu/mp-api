package service

import (
	"context"
	"fmt"
	"mime/multipart"

	"github.com/google/uuid"
	"github.com/mairuu/mp-api/internal/app"
	"github.com/mairuu/mp-api/internal/features/bucket/model"
	"github.com/mairuu/mp-api/internal/platform/authorization"
	"github.com/mairuu/mp-api/internal/platform/storage"
)

type Service struct {
	enforcer *authorization.Enforcer
	bucket   storage.Bucket
}

func NewService(enforcer *authorization.Enforcer, bucket storage.Bucket) *Service {
	return &Service{
		enforcer: enforcer,
		bucket:   bucket,
	}
}

func (s *Service) UploadFiles(ctx context.Context, ur *app.UserRole, files []*multipart.FileHeader) (*UploadDTO, error) {
	err := s.enforce(ur, model.ActionUpload, nil)
	if err != nil {
		return nil, err
	}

	// todo: quota check, e.g. max file size, max number of files, etc.

	var acceptedFiles []AcceptedFile
	var rejectedFiles []RejectedFile

	for _, file := range files {
		opts := &storage.UploadOptions{
			MetaData: map[string]string{
				"user_id": ur.ID.String(),
			},
		}

		f, err := file.Open()
		if err != nil {
			rejectedFiles = append(rejectedFiles, RejectedFile{
				OriginalFileName: file.Filename,
				Error:            err.Error(),
			})
			continue
		}

		objectName := fmt.Sprintf("_tmp/%s", uuid.New().String())
		if err := s.bucket.Upload(ctx, objectName, f, opts); err != nil {
			f.Close()
			rejectedFiles = append(rejectedFiles, RejectedFile{
				OriginalFileName: file.Filename,
				Error:            err.Error(),
			})
			continue
		}
		f.Close()

		// todo: track these temporary files and delete them after some period of time, e.g. 24 hours
		acceptedFiles = append(acceptedFiles, AcceptedFile{
			OriginalFileName: file.Filename,
			ObjectName:       objectName,
		})
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

func (s *Service) enforce(ur *app.UserRole, action authorization.Action, target authorization.ScopeResolvable) error {
	return s.enforcer.Enforce(ur.ID, ur.Role, model.ResourceBucket, action, target)
}
