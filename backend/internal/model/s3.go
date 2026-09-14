package model

import (
	"context"
	"io"
)

// S3Client interface to abstract the S3 operations
type S3Client interface {
	ListAllImages(ctx context.Context) ([]string, error)
	DeleteImage(ctx context.Context, filename string) error
	UploadImage(ctx context.Context, file io.Reader, filename string, contentType string) error
	GetImage(ctx context.Context, filename string) (io.ReadCloser, *string, error)
}
