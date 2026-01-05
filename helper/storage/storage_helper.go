package storage

import (
	"bytes"
	"context"
	"math/rand"
	"mime"
	"net/url"
	"os"
	"os/exec"
	"time"

	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/filesystem"
	"github.com/minio/minio-go/v7"
	spotlibsCtx "github.com/spotlibs/go-lib/ctx"
	"github.com/spotlibs/go-lib/log"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type storageHelper struct {
	minioClient *minio.Client
}
type StorageHelper interface{}

func NewStorageHelper(minioClient *minio.Client) StorageHelper {
	return &storageHelper{
		minioClient: minioClient,
	}
}

// Upload Goravel filesystem.File to MinIO
// dirpath example: "2025/12/25/Documents"
// logging to runtime info
func (h *storageHelper) MinioUpload(ctx context.Context, file *filesystem.File, dirpath string) error {
	filename := file.GetClientOriginalName()
	extension := file.GetClientOriginalExtension()
	b, err := os.ReadFile(file.File())
	if err != nil {
		log.Runtime(ctx).Warning(log.Map{
			"message": "error to read filesystem.File",
			"error":   err.Error(),
		})
		return err
	}
	metadata := spotlibsCtx.Get(ctx)
	identifier := metadata.UrlPath
	if identifier == "" {
		identifier = metadata.SignaturePath
	}
	_, err = h.minioClient.PutObject(
		ctx,
		facades.Config().GetString("MINIO_BUCKET", ""),
		dirpath+"/"+filename,
		bytes.NewBuffer(b),
		int64(len(b)),
		minio.PutObjectOptions{
			UserMetadata: map[string]string{
				"requestFrom": metadata.ReqUser,
				"requestID":   metadata.ReqId,
				"identifier":  identifier,
			},
			ContentType: mime.TypeByExtension(extension),
		},
	)
	return err
}
func (h *storageHelper) MinioTemporaryUrl(ctx context.Context, filepath string) (string, error) {
	reqParam := make(url.Values)
	tempFileUrl, err := h.minioClient.PresignedGetObject(
		ctx,
		facades.Config().GetString("MINIO_BUCKET", ""),
		filepath,
		time.Duration(facades.Config().GetInt("MINIO_EXPIRED_URL", 300)),
		reqParam,
	)
	if err != nil {
		log.Runtime(ctx).Warning(log.Map{
			"message": "error in generating temporary url from minio",
			"error":   err.Error(),
		})
		return "", err
	}
	return tempFileUrl.String(), nil
}
func (h *storageHelper) MinioMove(ctx context.Context, srcPath string, destPath string) error {
	_, err := h.minioClient.CopyObject(
		ctx,
		minio.CopyDestOptions{
			Bucket: facades.Config().GetString("MINIO_BUCKET", ""),
			Object: destPath,
		},
		minio.CopySrcOptions{
			Bucket: facades.Config().GetString("MINIO_BUCKET", ""),
			Object: srcPath,
		},
	)
	if err != nil {
		return err
	}
	err = h.minioClient.RemoveObject(
		ctx,
		facades.Config().GetString("MINIO_BUCKET", ""),
		srcPath,
		minio.RemoveObjectOptions{},
	)
	if err != nil {
		log.Runtime(ctx).Error(log.Map{
			"message": "error to remove source object after minio move",
			"error":   err.Error(),
		})
		return err
	}
	return nil
}
func (h *storageHelper) MinioCopy(ctx context.Context, srcPath string, destPath string) error {
	info, err := h.minioClient.CopyObject(
		ctx,
		minio.CopyDestOptions{
			Bucket: facades.Config().GetString("MINIO_BUCKET", ""),
			Object: destPath,
		},
		minio.CopySrcOptions{
			Bucket: facades.Config().GetString("MINIO_BUCKET", ""),
			Object: srcPath,
		},
	)
	if err != nil {
		return err
	}
	log.Runtime(ctx).Info(log.Map{
		"task": "minio copy done",
		"info": info,
	})
	return nil
}
func (h *storageHelper) MinioDelete(ctx context.Context, filepath string) error {
	err := h.minioClient.RemoveObject(
		ctx,
		facades.Config().GetString("MINIO_BUCKET", ""),
		filepath,
		minio.RemoveObjectOptions{},
	)
	return err
}
func (h *storageHelper) NFSUpload(ctx context.Context, file *filesystem.File, dirpath string) error {
	err := checkDir(ctx, dirpath)
	if err != nil {
		return err
	}
	_, err = file.Store(dirpath)
	if err != nil {
		log.Runtime(ctx).Error(log.Map{
			"message": "error to store file to nfs",
			"error":   err.Error(),
		})
		return err
	}
	return nil
}
func (h *storageHelper) NFSSecurelinkGenerate(ctx context.Context, filepath string) (string, error) {
	securelinkPath := generateSecurelinkName()
	err := exec.CommandContext(ctx, "ln", "-s", filepath, "/var/www/html/public/securelink/"+securelinkPath).Run()
	if err != nil {
		log.Runtime(ctx).Error(log.Map{
			"message": "error to create securelink",
			"error":   err.Error(),
		})
		return "", err
	}
	return securelinkPath, nil
}
func (h *storageHelper) NFSDelete(ctx context.Context, filepath string) error {
	_, err := os.Stat(filepath)
	if err != nil {
		log.Runtime(ctx).Error(log.Map{
			"message": "error to get file info before delete",
			"error":   err.Error(),
		})
		return err
	}
	err = os.Remove(filepath)
	if err != nil {
		log.Runtime(ctx).Error(log.Map{
			"message": "error to remove file",
			"error":   err.Error(),
		})
		return err
	}
	return nil
}
func (h *storageHelper) NFSMove(ctx context.Context, srcPath string, destPath string) error {
	err := exec.CommandContext(ctx, "mv", srcPath, destPath).Run()
	if err != nil {
		log.Runtime(ctx).Error(log.Map{
			"message": "error to move file",
			"error":   err.Error(),
		})
		return err
	}
	return nil
}
func (h *storageHelper) NFSCopy(ctx context.Context, srcPath string, destPath string) error {
	err := exec.CommandContext(ctx, "cp", srcPath, destPath).Run()
	if err != nil {
		log.Runtime(ctx).Error(log.Map{
			"message": "error to copy file",
			"error":   err.Error(),
		})
		return err
	}
	return nil
}
func generateSecurelinkName() string {
	b := make([]byte, 40) // 40 bytes = 80 hex characters = 40 characters
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
func checkDir(ctx context.Context, dirpath string) error {
	info, err := os.Stat(dirpath)
	if err != nil {
		log.Runtime(ctx).Error(log.Map{
			"message": "error to get dir info",
			"error":   err.Error(),
		})
		return err
	}
	if !info.IsDir() {
		err = os.MkdirAll(dirpath, 0664)
		if err != nil {
			log.Runtime(ctx).Error(log.Map{
				"message": "error to create dir",
				"error":   err.Error(),
			})
			return err
		}
	}
	return nil
}
