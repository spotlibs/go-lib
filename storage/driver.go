package storage

import (
	"io"

	"github.com/goravel/framework/contracts/filesystem"
)

// FileEntry is one entry returned by Files/AllFiles. Ported from PHP
// array{path: string, size: int|string}. We diverge from a literal any/union
// port and use *int64 for size: nil when the driver can't obtain it cheaply
// (PHP's ""), a real byte count otherwise.
type FileEntry struct {
	Path string
	Size *int64 // nil when unknown; byte count when known
}

// StorageDriver is the interface for storage driver implementations.
// Ported 1:1 from php-lib StorageDriverInterface.
type StorageDriver interface {
	// Upload stores an uploaded file at the given directory path and returns the result.
	Upload(file filesystem.File, dirpath string, filename string) (*Result, error)

	// CopySameDriver copies a file within the same driver (same device/bucket).
	CopySameDriver(srcPath, destPath string) (*Result, error)

	// MoveSameDriver moves a file within the same driver.
	MoveSameDriver(srcPath, destPath string) (*Result, error)

	// WriteStream writes an io.ReadCloser stream to the destination path.
	WriteStream(stream io.ReadCloser, destPath string) (*Result, error)

	// Delete removes a file at the given filepath.
	Delete(filepath string) error

	// Securelink generates a temporary secure URL for the given filepath.
	// ttl is optional; nil means use a default TTL.
	Securelink(filepath string, ttl *int) (string, error)

	// Exists checks whether a file exists at the given filepath.
	Exists(filepath string) bool

	// ReadStream opens a read stream for the given filepath.
	ReadStream(filepath string) (io.ReadCloser, error)

	// Files returns a shallow list of files (not directories) in the given dirpath.
	Files(dirpath string) ([]FileEntry, error)

	// AllFiles returns a recursive list of all files in the given dirpath.
	AllFiles(dirpath string) ([]FileEntry, error)

	// Info returns metadata for the given filepath.
	Info(filepath string) (*Result, error)
}
