package storage

import (
	"context"
	"io"
	"os"
	"slices"
	"strings"
	"testing"
)

func TestLocalBucket(t *testing.T) {
	b := setupLocalBucketTest(t)

	objName := "test_object.txt"
	data := "Hello, Local Storage!"
	reader := strings.NewReader(data)

	err := b.Upload(context.Background(), objName, reader, &UploadOptions{
		ContentType: "text/plain",
	})
	if err != nil {
		t.Fatalf("failed to upload object: %v", err)
	}

	// assert content with seeking
	downloadedReader, err := b.Download(context.Background(), objName)
	if err != nil {
		t.Fatalf("failed to download object: %v", err)
	}
	downloadedReader.Seek(4, io.SeekStart)
	downloadedData, err := io.ReadAll(downloadedReader)
	if err != nil {
		t.Fatalf("failed to read downloaded data: %v", err)
	}
	if string(downloadedData) != data[4:] {
		t.Errorf("downloaded data mismatch: got %q, want %q", string(downloadedData), data[4:])
	}

	// assert metadata
	meta, err := b.GetMetadata(context.Background(), objName)
	if err != nil {
		t.Fatalf("failed to get metadata: %v", err)
	}
	if meta.ContentType != "text/plain" {
		t.Errorf("metadata content type mismatch: got %q, want %q", meta.ContentType, "text/plain")
	}

	// delete object
	err = b.Delete(context.Background(), "test_object.txt")
	if err != nil {
		t.Fatalf("failed to delete object: %v", err)
	}
}

func TestLocalBucket_UploadOverwrite(t *testing.T) {
	b := setupLocalBucketTest(t)

	objName := "overwrite_test.txt"
	initialData := "Initial Data"
	updatedData := "Updated Data"

	// initial upload
	err := b.Upload(context.Background(), objName, strings.NewReader(initialData), &UploadOptions{
		ContentType: "text/plain",
	})
	if err != nil {
		t.Fatalf("failed to upload initial object: %v", err)
	}

	// overwrite upload
	err = b.Upload(context.Background(), objName, strings.NewReader(updatedData), &UploadOptions{
		ContentType: "text/plain",
	})
	if err != nil {
		t.Fatalf("failed to upload updated object: %v", err)
	}

	// download and verify content
	downloadedReader, err := b.Download(context.Background(), objName)
	if err != nil {
		t.Fatalf("failed to download object: %v", err)
	}
	downloadedData, err := io.ReadAll(downloadedReader)
	if err != nil {
		t.Fatalf("failed to read downloaded data: %v", err)
	}
	if string(downloadedData) != updatedData {
		t.Errorf("downloaded data mismatch after overwrite: got %q, want %q", string(downloadedData), updatedData)
	}
}

func TestLocalBucket_PathTraversal(t *testing.T) {
	b := setupLocalBucketTest(t)

	tests := []string{
		"../outside.txt",
		"/outside.txt",
		"fool/../../outside.txt",
	}

	for _, path := range tests {
		_, err := b.objectPath(path)
		if err == nil {
			t.Fatalf("expected error for path traversal with path %q, got nil", path)
		}
	}
}

func TestLocalBucket_NilUpload(t *testing.T) {
	b := setupLocalBucketTest(t)

	err := b.Upload(context.Background(), "nil_upload.txt", nil, &UploadOptions{
		ContentType: "text/plain",
	})
	if err == nil {
		t.Fatalf("expected error for nil upload, got nil")
	}
}

func TestLocalBucket_NonExistentDownload(t *testing.T) {
	b := setupLocalBucketTest(t)

	_, err := b.Download(context.Background(), "non_existent.txt")
	if err == nil {
		t.Fatalf("expected error for non-existent download, got nil")
	}
}

func TestLocalBucket_ContextCancellation(t *testing.T) {
	b := setupLocalBucketTest(t)

	data := "This upload will be cancelled."
	reader := strings.NewReader(data)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	err := b.Upload(ctx, "cancelled_upload.txt", reader, &UploadOptions{
		ContentType: "text/plain",
	})
	if err == nil {
		t.Fatalf("expected error for cancelled context, got nil")
	}
}

func TestLocalBucket_ListIter(t *testing.T) {
	b := setupLocalBucketTest(t)
	ctx := context.Background()

	// upload test files
	testFiles := []string{
		"images/photo1.jpg",
		"images/photo2.jpg",
		"images/thumbnails/thumb1.jpg",
		"documents/report.pdf",
		"documents/notes.txt",
		"documents/archives/2025.zip",
		"readme.md",
	}

	for _, filename := range testFiles {
		reader := strings.NewReader("test content for " + filename)
		err := b.Upload(ctx, filename, reader, &UploadOptions{
			ContentType: "application/octet-stream",
		})
		if err != nil {
			t.Fatalf("failed to upload %q: %v", filename, err)
		}
	}

	t.Run("list all objects", func(t *testing.T) {
		it := b.ListIter(ctx, "")
		objects := slices.Collect(it)

		if len(objects) != len(testFiles) {
			t.Errorf("expected %d objects, got %d", len(testFiles), len(objects))
		}

		for _, expected := range testFiles {
			found := slices.Contains(objects, expected)
			if !found {
				t.Errorf("expected object %q not found in list", expected)
			}
		}
	})

	t.Run("list with prefix images/", func(t *testing.T) {
		it := b.ListIter(ctx, "images/")
		objects := slices.Collect(it)

		expected := []string{
			"images/photo1.jpg",
			"images/photo2.jpg",
			"images/thumbnails/thumb1.jpg",
		}

		if len(objects) != len(expected) {
			t.Errorf("expected %d objects, got %d", len(expected), len(objects))
		}

		for _, exp := range expected {
			found := slices.Contains(objects, exp)
			if !found {
				t.Errorf("expected object %q not found in list", exp)
			}
		}
	})

	t.Run("list with prefix documents/", func(t *testing.T) {
		it := b.ListIter(ctx, "documents/")
		objects := slices.Collect(it)

		if len(objects) != 3 {
			t.Errorf("expected 3 objects, got %d", len(objects))
		}
	})

	t.Run("list with non-existent prefix", func(t *testing.T) {
		it := b.ListIter(ctx, "nonexistent/")
		objects := slices.Collect(it)

		if len(objects) != 0 {
			t.Errorf("expected 0 objects, got %d", len(objects))
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(context.Background())
		cancel() // cancel immediately

		it := b.ListIter(cancelCtx, "")
		objects := slices.Collect(it)
		if len(objects) != 0 {
			t.Errorf("expected 0 objects due to cancelled context, got %d", len(objects))
		}
	})

	t.Run("early termination", func(t *testing.T) {
		count := 0
		for range b.ListIter(ctx, "") {
			count++
			if count >= 2 {
				break // stop after 2 items
			}
		}

		if count != 2 {
			t.Errorf("expected to iterate 2 items, got %d", count)
		}
	})
}

func setupLocalBucketTest(t *testing.T) *LocalBucket {
	b, err := NewLocalBucket("run/test/bucket")
	if err != nil {
		t.Fatalf("failed to create local bucket: %v", err)
	}
	t.Cleanup(func() {
		_ = os.RemoveAll("run/test/bucket")
	})
	return b
}
