package storage

import (
	"context"
	"strings"

	"github.com/goravel/framework/contracts/filesystem"
	"github.com/minio/minio-go/v7"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type nfsHelper struct {
	err error
}
type minioHelper struct {
	err         error
	minioClient *minio.Client
	bucketName  string
}
type StorageHelper interface {
	Upload(ctx context.Context, file filesystem.File, dirpath string) error
	Move(ctx context.Context, srcPath string, destPath string) error
	Copy(ctx context.Context, srcPath string, destPath string) error
	Delete(ctx context.Context, filepath string) error
	Securelink(ctx context.Context, filepath string) (string, error)
	SecurelinkFolder(ctx context.Context, dirpath string) (string, error)
}

func Disk(ctx context.Context, diskConfig string) StorageHelper {
	if strings.Contains(diskConfig, "minio") {
		return ConfigureMinio(ctx, diskConfig)
	}
	return &nfsHelper{}
}
