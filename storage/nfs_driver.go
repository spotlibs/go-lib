package storage

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/goravel/framework/contracts/filesystem"
	"github.com/goravel/framework/facades"
	"github.com/spotlibs/go-lib/stderr"
)

// nfsDriver is the native-filesystem storage driver. Ported 1:1 from
// php-lib NfsDriver.
type nfsDriver struct{}

func newNfsDriver() *nfsDriver {
	return &nfsDriver{}
}

// resolvePath resolves the given path by checking NFS fallback locations.
func (d *nfsDriver) resolvePath(path string) string {
	// If path starts with /data/NFS_ or /; return path as-is
	if strings.HasPrefix(path, "/data/NFS_") || strings.HasPrefix(path, "/") {
		return path
	}

	// Read fallback env
	fallback := facades.Config().Env("PATH_NFS_FALLBACK_STORAGE_SPOTLIB", "").(string)
	if fallback == "" {
		return path
	}

	// Split on comma, trim, filter empties
	bases := strings.Split(fallback, ",")
	for _, base := range bases {
		base = strings.TrimSpace(base)
		if base == "" {
			continue
		}

		// Build full path: rtrim(base, '/') + '/' + ltrim(path, '/')
		fullPath := strings.TrimRight(base, "/") + "/" + strings.TrimLeft(path, "/")

		// Check if exists
		if _, err := os.Stat(fullPath); err == nil {
			return fullPath
		}
	}

	// No match found, return original path
	return path
}

// Exists checks if a file exists at the given path.
func (d *nfsDriver) Exists(filepath string) bool {
	_, err := os.Stat(d.resolvePath(filepath))
	return err == nil
}

// Upload uploads a file to the given directory.
func (d *nfsDriver) Upload(file filesystem.File, dirpath string, filename string) (*Result, error) {
	// Ensure destination directory exists
	if _, err := os.Stat(dirpath); os.IsNotExist(err) {
		if err := os.MkdirAll(dirpath, 0755); err != nil {
			return nil, stderr.ErrRuntime("Failed to create destination directory: " + dirpath)
		}
	}

	// Determine filename
	fileName := filename
	if fileName == "" {
		fileName = file.GetClientOriginalName()
	}

	// Use StoreAs to store the file directly to the NFS path
	fullPath := strings.TrimRight(dirpath, "/") + "/" + fileName
	filePath := file.File() // Get the temporary file path

	// Read the temporary file and write to destination
	src, err := os.Open(filePath)
	if err != nil {
		return nil, stderr.ErrRuntime("Failed to read uploaded file: " + fileName)
	}
	defer src.Close()

	dst, err := os.Create(fullPath)
	if err != nil {
		return nil, stderr.ErrRuntime("Failed to create destination file: " + fullPath)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return nil, stderr.ErrRuntime("Failed to write file: " + fullPath)
	}

	// Return result DTO
	result := &Result{
		Driver:   NFS,
		PathFile: fileName,
		FullPath: fullPath,
		Folder:   strings.TrimRight(dirpath, "/") + "/",
		Size:     nil,
	}

	return result, nil
}

// WriteStream writes a stream to the given destination path.
func (d *nfsDriver) WriteStream(stream io.ReadCloser, destPath string) (*Result, error) {
	defer stream.Close()

	// Ensure destination directory exists
	destDir := filepath.Dir(destPath)
	if _, err := os.Stat(destDir); os.IsNotExist(err) {
		if err := os.MkdirAll(destDir, 0755); err != nil {
			return nil, stderr.ErrRuntime("Failed to create destination directory: " + destDir)
		}
	}

	// Create destination file
	f, err := os.Create(destPath)
	if err != nil {
		return nil, stderr.ErrRuntime("Failed to write stream to NFS: " + destPath)
	}
	defer f.Close()

	// Copy stream to file
	if _, err := io.Copy(f, stream); err != nil {
		return nil, stderr.ErrRuntime("Failed to write stream to NFS: " + destPath)
	}

	// Return result DTO
	result := &Result{
		Driver:   NFS,
		PathFile: filepath.Base(destPath),
		FullPath: destPath,
		Folder:   destDir + "/",
		Size:     nil,
	}

	return result, nil
}

// CopySameDriver copies a file within the same driver.
func (d *nfsDriver) CopySameDriver(srcPath, destPath string) (*Result, error) {
	srcPath = d.resolvePath(srcPath)

	// Check source exists
	if _, err := os.Stat(srcPath); os.IsNotExist(err) {
		return nil, stderr.ErrRuntime("Source file not found: " + srcPath)
	}

	// Ensure destination directory exists
	destDir := filepath.Dir(destPath)
	if _, err := os.Stat(destDir); os.IsNotExist(err) {
		if err := os.MkdirAll(destDir, 0755); err != nil {
			return nil, stderr.ErrRuntime("Failed to create destination directory: " + destDir)
		}
	}

	// Copy file
	src, err := os.Open(srcPath)
	if err != nil {
		return nil, stderr.ErrRuntime("Failed to copy file to: " + destPath)
	}
	defer src.Close()

	dst, err := os.Create(destPath)
	if err != nil {
		return nil, stderr.ErrRuntime("Failed to copy file to: " + destPath)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return nil, stderr.ErrRuntime("Failed to copy file to: " + destPath)
	}

	// Return result DTO
	result := &Result{
		Driver:   NFS,
		PathFile: filepath.Base(destPath),
		FullPath: destPath,
		Folder:   destDir + "/",
		Size:     nil,
	}

	return result, nil
}

// MoveSameDriver moves a file within the same driver.
func (d *nfsDriver) MoveSameDriver(srcPath, destPath string) (*Result, error) {
	srcPath = d.resolvePath(srcPath)

	// Check source exists
	if _, err := os.Stat(srcPath); os.IsNotExist(err) {
		return nil, stderr.ErrRuntime("Source file not found: " + srcPath)
	}

	// Ensure destination directory exists (0775 mode for move)
	destDir := filepath.Dir(destPath)
	if _, err := os.Stat(destDir); os.IsNotExist(err) {
		if err := os.MkdirAll(destDir, 0775); err != nil {
			return nil, stderr.ErrRuntime("Failed to create destination directory: " + destDir)
		}
	}

	// Move file
	if err := os.Rename(srcPath, destPath); err != nil {
		return nil, stderr.ErrRuntime("Failed to move file to: " + destPath)
	}

	// Return result DTO
	result := &Result{
		Driver:   NFS,
		PathFile: filepath.Base(destPath),
		FullPath: destPath,
		Folder:   destDir + "/",
		Size:     nil,
	}

	return result, nil
}

// Delete deletes a file.
func (d *nfsDriver) Delete(filepath string) error {
	filepath = d.resolvePath(filepath)

	// Check if it's a folder
	if strings.HasSuffix(filepath, "/") {
		return stderr.ErrRuntime("Cannot delete a folder: " + filepath)
	}

	// Check if it exists
	fi, err := os.Stat(filepath)
	if os.IsNotExist(err) {
		return stderr.ErrRuntime("File not found: " + filepath)
	}
	if err == nil && fi.IsDir() {
		return stderr.ErrRuntime("Cannot delete a folder: " + filepath)
	}

	// Delete file
	if err := os.Remove(filepath); err != nil {
		return stderr.ErrRuntime("Failed to delete file: " + filepath)
	}

	return nil
}

// Securelink creates a secure symlink to a file.
func (d *nfsDriver) Securelink(filepath string, ttl *int) (string, error) {
	_ = ttl // ttl parameter is ignored for NFS
	filepath = d.resolvePath(filepath)

	// Check if file exists and is a regular file
	fi, err := os.Stat(filepath)
	if os.IsNotExist(err) || fi.IsDir() {
		return "", stderr.ErrRuntime("File not found within filepath: " + filepath)
	}

	// Get extension
	ext := ""
	if lastDot := strings.LastIndex(filepath, "."); lastDot != -1 {
		ext = filepath[lastDot+1:]
	}

	// Generate random filename
	random := randomString(40)
	if ext != "" {
		random = random + "." + ext
	}

	// Create securelink directory
	securelinkDir := "/var/www/html/public/securelink"
	if _, err := os.Stat(securelinkDir); os.IsNotExist(err) {
		os.MkdirAll(securelinkDir, 0755)
	}

	// Create symlink
	linkPath := securelinkDir + "/" + random
	if err := os.Symlink(filepath, linkPath); err != nil {
		return "", stderr.ErrRuntime("Failed to create secure link for file: " + filepath)
	}

	// Return URL
	appURL := facades.Config().Env("APP_URL", "").(string)
	return appURL + "/securelink/" + random, nil
}

// ReadStream opens a file for reading as a stream.
func (d *nfsDriver) ReadStream(filepath string) (io.ReadCloser, error) {
	filepath = d.resolvePath(filepath)

	// Check if file exists
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return nil, stderr.ErrRuntime("File not found: " + filepath)
	}

	// Open file for reading
	f, err := os.Open(filepath)
	if err != nil {
		return nil, stderr.ErrRuntime("Failed to open read stream for: " + filepath)
	}

	return f, nil
}

// Files returns a list of files (non-recursive) in a directory.
func (d *nfsDriver) Files(dirpath string) ([]FileEntry, error) {
	dirpath = d.resolvePath(dirpath)

	// Check if directory exists
	if _, err := os.Stat(dirpath); os.IsNotExist(err) {
		return nil, nil
	}

	entries, err := os.ReadDir(dirpath)
	if err != nil {
		return nil, nil
	}

	var result []FileEntry
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		fullPath := strings.TrimRight(dirpath, "/") + "/" + entry.Name()
		fi, err := entry.Info()
		if err != nil {
			result = append(result, FileEntry{
				Path: fullPath,
				Size: nil,
			})
			continue
		}

		sz := fi.Size()
		result = append(result, FileEntry{
			Path: fullPath,
			Size: &sz,
		})
	}

	return result, nil
}

// AllFiles returns a list of all files (recursive) in a directory.
func (d *nfsDriver) AllFiles(dirpath string) ([]FileEntry, error) {
	dirpath = d.resolvePath(dirpath)

	// Check if directory exists
	if _, err := os.Stat(dirpath); os.IsNotExist(err) {
		return nil, nil
	}

	var result []FileEntry

	// Walk directory recursively
	err := filepath.WalkDir(dirpath, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories, only process files
		if entry.IsDir() {
			return nil
		}

		fi, err := entry.Info()
		if err != nil {
			result = append(result, FileEntry{
				Path: path,
				Size: nil,
			})
			return nil
		}

		sz := fi.Size()
		result = append(result, FileEntry{
			Path: path,
			Size: &sz,
		})

		return nil
	})

	if err != nil {
		return nil, nil
	}

	return result, nil
}

// Info returns information about a file.
func (d *nfsDriver) Info(filepath string) (*Result, error) {
	filepath = d.resolvePath(filepath)

	// Check if file exists
	fi, err := os.Stat(filepath)
	if os.IsNotExist(err) {
		return nil, stderr.ErrRuntime("File not found: " + filepath)
	}

	sz := fi.Size()
	result := &Result{
		Driver:   NFS,
		PathFile: filepath[strings.LastIndex(filepath, "/")+1:],
		FullPath: filepath,
		Folder:   filepath[:strings.LastIndex(filepath, "/")] + "/",
		Size:     &sz,
	}

	return result, nil
}

// randomString generates a random string of the given length.
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[randIntn(len(charset))]
	}
	return string(b)
}

// randIntn is a simple pseudo-random integer generator.
func randIntn(n int) int {
	return int(os.Getpid()%n+1) % n
}
