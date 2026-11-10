package excel

import (
	"io"

	"github.com/spotlibs/go-lib/stderr"
	"github.com/spotlibs/go-lib/storage"
)

// OpenFromMinIOWithOptions opens an existing Excel file stored on MinIO at srcPath using opts.
// srcPath: full path on MinIO including filename, e.g. "reports/report.xlsx".
// Returns error if the file cannot be read or parsed.
func OpenFromMinIOWithOptions(srcPath string, opts OpenOptions) (*ExcelFile, error) {
	return openFromDriverWithOptions(storage.MINIO, srcPath, opts)
}

// OpenFromNFSWithOptions opens an existing Excel file stored on NFS at srcPath using opts.
// srcPath: full absolute path on NFS including filename, e.g. "/data/NFS_spotlib/reports/report.xlsx".
// Returns error if the file cannot be read or parsed.
func OpenFromNFSWithOptions(srcPath string, opts OpenOptions) (*ExcelFile, error) {
	return openFromDriverWithOptions(storage.NFS, srcPath, opts)
}

// openFromDriverWithOptions reads srcPath from the given storage driver and parses it using opts.
func openFromDriverWithOptions(driverName string, srcPath string, opts OpenOptions) (*ExcelFile, error) {
	stream, err := storage.ReadStreamFromDriver(driverName, srcPath)
	if err != nil {
		return nil, stderr.ErrRuntime("failed to read excel from " + driverName + ": " + srcPath)
	}
	defer stream.Close()

	return OpenReaderWithOptions(io.Reader(stream), opts)
}
