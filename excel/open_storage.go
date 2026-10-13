package excel

import (
	"io"

	"github.com/goravel/framework/facades"
	"github.com/spotlibs/go-lib/stderr"
	"github.com/spotlibs/go-lib/storage"
)

// OpenFromMinIO opens an existing Excel file stored on MinIO at srcPath.
// srcPath: full path on MinIO including filename, e.g. "reports/report.xlsx".
// Returns error if the file cannot be read or parsed.
func OpenFromMinIO(srcPath string) (*ExcelFile, error) {
	return openFromDriver(storage.MINIO, srcPath)
}

// OpenFromNFS opens an existing Excel file stored on NFS at srcPath.
// srcPath: full absolute path on NFS including filename, e.g. "/data/NFS_spotlib/reports/report.xlsx".
// Returns error if the file cannot be read or parsed.
func OpenFromNFS(srcPath string) (*ExcelFile, error) {
	return openFromDriver(storage.NFS, srcPath)
}

// OpenFromStorageNextGen opens an existing Excel file from the driver configured in
// the DEFAULT_DRIVER_STORAGE environment variable.
// Valid values for DEFAULT_DRIVER_STORAGE: "NFS", "MINIO", "MINIO_BRIMEN".
// srcPath: full source path on the resolved driver including filename.
// Returns error if the driver is not configured, or the file cannot be read or parsed.
func OpenFromStorageNextGen(srcPath string) (*ExcelFile, error) {
	driverEnv := facades.Config().Env("DEFAULT_DRIVER_STORAGE", "")
	driverStr, _ := driverEnv.(string)
	if driverStr == "" {
		return nil, stderr.ErrRuntime("DEFAULT_DRIVER_STORAGE is not configured")
	}
	return openFromDriver(driverStr, srcPath)
}

// openFromDriver reads srcPath from the given storage driver and parses it as a workbook.
func openFromDriver(driverName string, srcPath string) (*ExcelFile, error) {
	stream, err := storage.ReadStreamFromDriver(driverName, srcPath)
	if err != nil {
		return nil, stderr.ErrRuntime("failed to read excel from " + driverName + ": " + srcPath)
	}
	defer stream.Close()

	return OpenReader(io.Reader(stream))
}
