package excel

import (
	"bytes"
	"io"

	"github.com/goravel/framework/facades"
	"github.com/spotlibs/go-lib/stderr"
	"github.com/spotlibs/go-lib/storage"
)

// SaveAsMinIO serializes the workbook to memory and writes it to MinIO at destPath.
// destPath: full path on MinIO including filename, e.g. "minio_storage/reports/report.xlsx".
// Returns the storage.Result from the write operation and error.
func (e *ExcelFile) SaveAsMinIO(destPath string) (*storage.Result, error) {
	buf, err := e.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	stream := io.NopCloser(bytes.NewReader(buf.Bytes()))
	result, err := storage.WriteStreamToDriver(storage.MINIO, stream, destPath)
	if err != nil {
		return nil, stderr.ErrRuntime("failed to save excel to MinIO: " + destPath)
	}
	return result, nil
}

// SaveAsNFS serializes the workbook to memory and writes it to NFS at destPath.
// destPath: full absolute path on NFS including filename, e.g. "/data/NFS_spotlib/reports/report.xlsx".
// Returns the storage.Result from the write operation and error.
func (e *ExcelFile) SaveAsNFS(destPath string) (*storage.Result, error) {
	buf, err := e.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	stream := io.NopCloser(bytes.NewReader(buf.Bytes()))
	result, err := storage.WriteStreamToDriver(storage.NFS, stream, destPath)
	if err != nil {
		return nil, stderr.ErrRuntime("failed to save excel to NFS: " + destPath)
	}
	return result, nil
}

// SaveAsStorageNextGen serializes the workbook to memory and writes it to the driver
// configured in the DEFAULT_DRIVER_STORAGE environment variable.
// Valid values for DEFAULT_DRIVER_STORAGE: "NFS", "MINIO", "MINIO_BRIMEN".
// destPath: full destination path on the resolved driver including filename.
// Returns the storage.Result from the write operation and error.
func (e *ExcelFile) SaveAsStorageNextGen(destPath string) (*storage.Result, error) {
	driverEnv := facades.Config().Env("DEFAULT_DRIVER_STORAGE", "")
	driverStr, _ := driverEnv.(string)
	if driverStr == "" {
		return nil, stderr.ErrRuntime("DEFAULT_DRIVER_STORAGE is not configured")
	}

	buf, err := e.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	stream := io.NopCloser(bytes.NewReader(buf.Bytes()))
	result, err := storage.WriteStreamToDriver(driverStr, stream, destPath)
	if err != nil {
		return nil, stderr.ErrRuntime("failed to save excel to storage: " + destPath)
	}
	return result, nil
}
