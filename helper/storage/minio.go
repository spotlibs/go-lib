package storage

import (
	"context"
	"errors"

	"github.com/goravel/framework/facades"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/spotlibs/go-lib/log"
)

func ConfigureMinio(ctx context.Context, diskConfig string) *minioHelper {
	minioClient := facades.Config().Get(diskConfig + ".client").(*minio.Client)
	if minioClient == nil {
		configMap := facades.Config().Get("filesystems.disks." + diskConfig).(map[string]any)
		diskMap, ok := configMap[diskConfig].(map[string]any)
		if !ok {
			log.Runtime(ctx).Error(log.Map{
				"message":    "MinIO client configuration not found",
				"diskConfig": diskConfig,
			})
			return &minioHelper{err: errors.New("MinIO client configuration not found")}
		}
		minioClient, err := minio.New(
			diskMap["endpoint"].(string),
			&minio.Options{
				Creds: credentials.NewStaticV4(
					diskMap["accessKey"].(string),
					diskMap["secretKey"].(string),
					"",
				),
				Secure: diskMap["secure"].(bool),
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
		return &minioHelper{minioClient: minioClient}
	}
	return &minioHelper{minioClient: minioClient}
}
