package storage

import (
	"context"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
)

type StorageService interface {
	UploadFile(ctx context.Context, localFilePath, destFolder string) (string, error)
	DeleteFile(ctx context.Context, publicID string) error
	GetDownloadURL(ctx context.Context, resourceType, publicID string, expires time.Duration) (string, error)
	GetSecureDownloadURL(ctx context.Context, resourceType, publicID string, expires time.Duration) (string, error)
	UploadKYPFile(ctx context.Context, localFilePath, destFolder, adminKey string) (string, error)
}

type StorageServiceImpl struct {
	cld       *cloudinary.Cloudinary
	cloudName string
	apiSecret string
}
