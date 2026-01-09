package storage

import (
	"context"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/goravel/framework/contracts/filesystem"
	"github.com/goravel/framework/facades"
	"github.com/minio/minio-go/v7"
	spotlibsCtx "github.com/spotlibs/go-lib/ctx"
	"github.com/spotlibs/go-lib/log"
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
}

func Disk(ctx context.Context, diskConfig string) StorageHelper {
	if strings.Contains(diskConfig, "minio") {
		return ConfigureMinio(ctx, diskConfig)
	}
	return &nfsHelper{}
}

func (h *minioHelper) Upload(ctx context.Context, file filesystem.File, dirpath string) error {
	if h.err != nil {
		return h.err
	}
	path, err := file.Store("tempfiles")
	if err != nil {
		return err
	}
	defer exec.CommandContext(ctx, "rm", path).Run()
	ctxSpotlibs := spotlibsCtx.Get(ctx)
	identifier := ctxSpotlibs.ReqId
	if identifier == "" {
		identifier = ctxSpotlibs.SignaturePath
	}
	info, err := h.minioClient.FPutObject(
		ctx,
		h.bucketName,
		file.File(),
		path,
		minio.PutObjectOptions{
			UserMetadata: map[string]string{
				"original-filename": file.File(),
				"uploader-user":     ctxSpotlibs.ReqUser,
				"uploader-name":     ctxSpotlibs.ReqNama,
				"identifier":        identifier,
				"traceID":           ctxSpotlibs.ReqId,
			},
		},
	)
	if err != nil {
		return err
	}
	log.Runtime(ctx).Info(log.Map{
		"message": "File uploaded to MinIO successfully",
		"bucket":  info.Bucket,
		"object":  info.Key,
		"size":    info.Size,
	})
	return nil
}
func (h *minioHelper) Move(ctx context.Context, srcPath string, destPath string) error {
	if h.err != nil {
		return h.err
	}
	info, err := h.minioClient.CopyObject(
		ctx,
		minio.CopyDestOptions{
			Bucket: h.bucketName,
			Object: destPath,
		},
		minio.CopySrcOptions{
			Bucket: h.bucketName,
			Object: srcPath,
		},
	)
	if err != nil {
		return err
	}
	log.Runtime(ctx).Info(log.Map{
		"message":     "File moved in MinIO successfully",
		"source":      srcPath,
		"destination": destPath,
		"etag":        info.ETag,
	})
	return h.Delete(ctx, srcPath)
}
func (h *minioHelper) Copy(ctx context.Context, srcPath string, destPath string) error {
	if h.err != nil {
		return h.err
	}
	info, err := h.minioClient.CopyObject(
		ctx,
		minio.CopyDestOptions{
			Bucket: h.bucketName,
			Object: destPath,
		},
		minio.CopySrcOptions{
			Bucket: h.bucketName,
			Object: srcPath,
		},
	)
	if err != nil {
		return err
	}
	log.Runtime(ctx).Info(log.Map{
		"message":     "File copied in MinIO successfully",
		"source":      srcPath,
		"destination": destPath,
		"etag":        info.ETag,
	})
	return nil
}
func (h *minioHelper) Delete(ctx context.Context, filepath string) error {
	if h.err != nil {
		return h.err
	}
	err := h.minioClient.RemoveObject(
		ctx,
		h.bucketName,
		filepath,
		minio.RemoveObjectOptions{},
	)
	if err != nil {
		log.Runtime(ctx).Warning(log.Map{
			"message":  "Failed to delete file from MinIO",
			"filepath": filepath,
			"error":    err.Error(),
		})
		return err
	}
	return nil
}
func (h *minioHelper) Securelink(ctx context.Context, filepath string) (string, error) {
	if h.err != nil {
		return "", h.err
	}
	link, err := h.minioClient.PresignedGetObject(
		ctx,
		h.bucketName,
		filepath,
		time.Duration(facades.Config().GetInt("MINIO_EXPIRED_URL", 600)),
		nil,
	)
	if err != nil {
		return "", err
	}
	return link.String(), nil
}

func (h *nfsHelper) Upload(ctx context.Context, file filesystem.File, dirpath string) error {
	if h.err != nil {
		return h.err
	}
	info, err := os.Stat(dirpath)
	if err != nil || !info.IsDir() {
		err = os.MkdirAll(dirpath, 0664)
		if err != nil {
			log.Runtime(ctx).Error(log.Map{
				"message": "Failed to create directory for NFS upload",
				"dirpath": dirpath,
				"error":   err.Error(),
			})
			return err
		}
	}
	path, err := file.Store("tempfiles")
	if err != nil {
		return err
	}
	defer exec.CommandContext(ctx, "rm", path).Run()
	destPath := dirpath + "/" + file.File()
	if err = exec.CommandContext(ctx, "mv", path, destPath).Run(); err != nil {
		log.Runtime(ctx).Error(log.Map{
			"message":  "Failed to move file to NFS directory",
			"destPath": destPath,
			"error":    err.Error(),
		})
		return err
	}
	return nil
}
func (h *nfsHelper) Move(ctx context.Context, srcPath string, destPath string) error {
	if h.err != nil {
		return h.err
	}
	if err := h.Copy(ctx, srcPath, destPath); err != nil {
		return err
	}
	return h.Delete(ctx, srcPath)
}
func (h *nfsHelper) Copy(ctx context.Context, srcPath string, destPath string) error {
	if h.err != nil {
		return h.err
	}
	_, err := os.Stat(srcPath)
	if err != nil {
		log.Runtime(ctx).Warning(log.Map{
			"message": "Source file not found for NFS copy",
			"srcPath": srcPath,
			"error":   err.Error(),
		})
		return err
	}
	temp := strings.Split(destPath, "/")
	if len(temp) > 1 {
		dirpath := strings.Join(temp[0:len(temp)-1], "/")
		info, err := os.Stat(dirpath)
		if err != nil || !info.IsDir() {
			err = os.MkdirAll(dirpath, 0664)
			if err != nil {
				log.Runtime(ctx).Error(log.Map{
					"message": "Failed to create directory for NFS copy",
					"dirpath": dirpath,
					"error":   err.Error(),
				})
				return err
			}
		}
	}
	err = exec.CommandContext(ctx, "cp", srcPath, destPath).Run()
	if err != nil {
		log.Runtime(ctx).Error(log.Map{
			"message":     "Failed to copy file in NFS",
			"source":      srcPath,
			"destination": destPath,
			"error":       err.Error(),
		})
		return err
	}
	return nil
}
func (h *nfsHelper) Delete(ctx context.Context, filepath string) error {
	if h.err != nil {
		return h.err
	}
	err := exec.CommandContext(ctx, "rm", filepath).Run()
	if err != nil {
		log.Runtime(ctx).Warning(log.Map{
			"message":  "Failed to delete file from NFS",
			"filepath": filepath,
			"error":    err.Error(),
		})
		return err
	}
	return nil
}
func (h *nfsHelper) Securelink(ctx context.Context, filepath string) (string, error) {
	if h.err != nil {
		return "", h.err
	}
	_, err := os.Stat(filepath)
	if err != nil {
		log.Runtime(ctx).Warning(log.Map{
			"message":  "File not found for secure link generation",
			"filepath": filepath,
			"error":    err.Error(),
		})
		return "", err
	}
	// Generate a secure link for the file
	linkname := randomString(40)
	err = exec.CommandContext(ctx, "ln", "-s", filepath, "/var/www/html/public/securelink/"+linkname).Run()
	if err != nil {
		return "", err
	}
	return linkname, nil
}

func randomString(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
