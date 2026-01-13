package storage

import (
	"context"
	"errors"
	"os/exec"
	"time"

	"github.com/goravel/framework/contracts/filesystem"
	"github.com/goravel/framework/facades"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	spotlibsCtx "github.com/spotlibs/go-lib/ctx"
	"github.com/spotlibs/go-lib/log"
)

func ConfigureMinio(ctx context.Context, diskConfig string) *minioHelper {
	configMap, ok := facades.Config().Get("filesystems.disks." + diskConfig).(map[string]any)
	if !ok {
		log.Runtime(ctx).Error(log.Map{
			"message":    "MinIO client configuration not found",
			"diskConfig": diskConfig,
		})
		return &minioHelper{err: errors.New("MinIO client configuration not found")}
	}
	minioClient := facades.Config().Get(diskConfig + ".client").(*minio.Client)
	if minioClient == nil {
		minioClient, err := minio.New(
			configMap["endpoint"].(string),
			&minio.Options{
				Creds: credentials.NewStaticV4(
					configMap["accessKey"].(string),
					configMap["secretKey"].(string),
					"",
				),
				Secure: configMap["secure"].(bool),
			},
		)
		if err != nil {
			log.Runtime(ctx).Error(log.Map{
				"message":    "Failed to create MinIO client",
				"diskConfig": diskConfig,
				"error":      err.Error(),
			})
			return &minioHelper{err: err}
		}
		facades.Config().Add(diskConfig+".client", minioClient) // Cache the client for future use
		return &minioHelper{minioClient: minioClient, bucketName: configMap["bucketName"].(string)}
	}
	return &minioHelper{minioClient: minioClient, bucketName: configMap["bucketName"].(string)}
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
func (h *minioHelper) SecurelinkFolder(ctx context.Context, dirpath string) (string, error) {
	return "", nil
}
