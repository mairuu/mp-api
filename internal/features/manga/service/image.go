package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"

	"github.com/chai2010/webp"
	"github.com/mairuu/mp-api/internal/features/manga/model"
	"github.com/mairuu/mp-api/internal/platform/storage"
)

const (
	quality = 81 // webp quality
)

func (s *Service) decodeImage(f io.Reader) (image.Image, error) {
	// decode image
	img, _, err := image.Decode(f)
	if err != nil {
		if errors.Is(err, image.ErrFormat) {
			return nil, fmt.Errorf("%w; please use WebP, JPEG, PNG or GIF", model.ErrUnsupportedImageFormat)
		}
		return nil, err
	}

	return img, nil
}

func (s *Service) uploadImage(ctx context.Context, objectName string, img image.Image) error {
	var buf bytes.Buffer
	if err := webp.Encode(&buf, img, &webp.Options{Quality: quality}); err != nil {
		return fmt.Errorf("webp encoding failed: %w", err)
	}

	opts := &storage.UploadOptions{
		ContentType: "image/webp",
	}
	if err := s.publicBucket.Upload(ctx, objectName, bytes.NewReader(buf.Bytes()), opts); err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}

	return nil
}
