package storage

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/goravel/framework/contracts/filesystem"
	"github.com/goravel/framework/facades"
	"github.com/spotlibs/go-lib/stderr"
)

// minioDriver is the MinIO-backed storage driver. Ported 1:1 from
// php-lib MinioDriver.
type minioDriver struct {
	disk string
}

func newMinioDriver(disk string) *minioDriver {
	return &minioDriver{disk: disk}
}

// driverName returns the Storage driver constant corresponding to this
// MinioDriver's disk name. Ported 1:1 from php-lib MinioDriver's repeated
// ternary ($this->disk === 'minio_brimen' ? MINIO_BRIMEN : MINIO).
func (d *minioDriver) driverName() string {
	if d.disk == "minio_brimen" {
		return MINIO_BRIMEN
	}
	return MINIO
}

// Exists checks whether a file exists on the MinIO disk.
func (d *minioDriver) Exists(filepath string) bool {
	return facades.Storage().Disk(d.disk).Exists(filepath)
}

// ReadStream opens a readable stream for a MinIO file.
//
// NOTE: the goravel filesystem.Driver contract
// (vendor/github.com/goravel/framework/contracts/filesystem/storage.go) has
// no readStream/writeStream method. The closest 1:1 is GetBytes (read) /
// Put (write). We wrap the bytes in an io.NopCloser to satisfy the
// io.ReadCloser return type.
func (d *minioDriver) ReadStream(filepath string) (io.ReadCloser, error) {
	data, err := facades.Storage().Disk(d.disk).GetBytes(filepath)
	if err != nil || data == nil {
		return nil, stderr.ErrRuntime("Failed to open read stream for: " + filepath)
	}

	return io.NopCloser(strings.NewReader(string(data))), nil
}

// Upload stores an uploaded file on the MinIO disk.
func (d *minioDriver) Upload(file filesystem.File, dirpath string, filename string) (*Result, error) {
	fileName := filename
	if fileName == "" {
		fileName = file.GetClientOriginalName()
	}
	destPath := strings.TrimRight(dirpath, "/") + "/" + fileName

	content, err := os.ReadFile(file.File())
	if err != nil {
		return nil, stderr.ErrRuntime("Failed to upload file to MinIO: " + destPath)
	}

	if err := facades.Storage().Disk(d.disk).Put(destPath, string(content)); err != nil {
		return nil, stderr.ErrRuntime("Failed to upload file to MinIO: " + destPath)
	}

	result := &Result{
		Driver:   d.driverName(),
		PathFile: fileName,
		FullPath: destPath,
		Folder:   strings.TrimRight(dirpath, "/") + "/",
		Size:     nil,
	}

	return result, nil
}

// WriteStream writes a stream to the given destination path on the MinIO disk.
//
// NOTE: the goravel filesystem.Driver contract has no writeStream method.
// The closest 1:1 is reading the stream fully and calling Put with the
// resulting content string.
func (d *minioDriver) WriteStream(stream io.ReadCloser, destPath string) (*Result, error) {
	defer stream.Close()

	content, err := io.ReadAll(stream)
	if err != nil {
		return nil, stderr.ErrRuntime("Failed to write stream to MinIO: " + destPath)
	}

	if err := facades.Storage().Disk(d.disk).Put(destPath, string(content)); err != nil {
		return nil, stderr.ErrRuntime("Failed to write stream to MinIO: " + destPath)
	}

	result := &Result{
		Driver:   d.driverName(),
		PathFile: filepath.Base(destPath),
		FullPath: destPath,
		Folder:   filepath.Dir(destPath) + "/",
		Size:     nil,
	}

	return result, nil
}

// CopySameDriver copies a file within the same MinIO disk.
func (d *minioDriver) CopySameDriver(srcPath, destPath string) (*Result, error) {
	if err := facades.Storage().Disk(d.disk).Copy(srcPath, destPath); err != nil {
		return nil, stderr.ErrRuntime("MinIO copy failed: " + srcPath + " -> " + destPath)
	}

	result := &Result{
		Driver:   d.driverName(),
		PathFile: filepath.Base(destPath),
		FullPath: destPath,
		Folder:   filepath.Dir(destPath) + "/",
		Size:     nil,
	}

	return result, nil
}

// MoveSameDriver moves a file within the same MinIO disk.
func (d *minioDriver) MoveSameDriver(srcPath, destPath string) (*Result, error) {
	if err := facades.Storage().Disk(d.disk).Move(srcPath, destPath); err != nil {
		return nil, stderr.ErrRuntime("MinIO move failed: " + srcPath + " -> " + destPath)
	}

	result := &Result{
		Driver:   d.driverName(),
		PathFile: filepath.Base(destPath),
		FullPath: destPath,
		Folder:   filepath.Dir(destPath) + "/",
		Size:     nil,
	}

	return result, nil
}

// Delete removes a file from the MinIO disk.
//
// NOTE: the goravel filesystem.Driver contract has no directoryExists
// method, so only the trailing-slash folder guard from the PHP source is
// replicated here (the method_exists($disk,'directoryExists') branch has
// no equivalent on this contract).
func (d *minioDriver) Delete(filepath string) error {
	if strings.HasSuffix(filepath, "/") {
		return stderr.ErrRuntime("Cannot delete a folder: " + filepath)
	}

	if err := facades.Storage().Disk(d.disk).Delete(filepath); err != nil {
		return stderr.ErrRuntime("MinIO delete failed: " + filepath)
	}

	return nil
}

// Securelink generates a temporary URL for a MinIO file.
func (d *minioDriver) Securelink(filepath string, ttl *int) (string, error) {
	ttlSeconds := 60
	if ttl != nil {
		ttlSeconds = *ttl
	} else {
		ttlSeconds = facades.Config().GetInt("MINIO_EXPIRED_URL", 60)
	}

	return facades.Storage().Disk(d.disk).TemporaryUrl(filepath, time.Now().Add(time.Duration(ttlSeconds)*time.Second))
}

// Files returns a shallow list of files in the given directory on the MinIO disk.
func (d *minioDriver) Files(dirpath string) ([]FileEntry, error) {
	disk := facades.Storage().Disk(d.disk)
	paths, err := disk.Files(dirpath)
	if err != nil {
		return nil, nil
	}

	var result []FileEntry
	for _, path := range paths {
		sz, sizeErr := disk.Size(path)
		if sizeErr != nil {
			result = append(result, FileEntry{Path: path, Size: nil})
			continue
		}
		size := sz
		result = append(result, FileEntry{Path: path, Size: &size})
	}

	return result, nil
}

// AllFiles returns a recursive list of all files in the given directory on the MinIO disk.
func (d *minioDriver) AllFiles(dirpath string) ([]FileEntry, error) {
	disk := facades.Storage().Disk(d.disk)
	paths, err := disk.AllFiles(dirpath)
	if err != nil {
		return nil, nil
	}

	var result []FileEntry
	for _, path := range paths {
		sz, sizeErr := disk.Size(path)
		if sizeErr != nil {
			result = append(result, FileEntry{Path: path, Size: nil})
			continue
		}
		size := sz
		result = append(result, FileEntry{Path: path, Size: &size})
	}

	return result, nil
}

// Info returns metadata for a file on the MinIO disk.
func (d *minioDriver) Info(filepath string) (*Result, error) {
	disk := facades.Storage().Disk(d.disk)
	if !disk.Exists(filepath) {
		return nil, stderr.ErrRuntime("File not found: " + filepath)
	}

	result := &Result{
		Driver:   d.driverName(),
		PathFile: filepathBase(filepath),
		FullPath: filepath,
		Folder:   filepathDir(filepath) + "/",
		Size:     nil,
	}

	sz, err := disk.Size(filepath)
	if err == nil {
		result.Size = &sz
	}

	return result, nil
}

// filepathBase and filepathDir are thin wrappers around path/filepath to
// avoid shadowing the `filepath` parameter name used throughout this file
// (PHP's `basename`/`dirname` 1:1 equivalents).
func filepathBase(p string) string {
	return filepath.Base(p)
}

func filepathDir(p string) string {
	return filepath.Dir(p)
}
