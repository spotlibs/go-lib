// Package storage provides single entry point for storage operations.
// Ported 1:1 from php-lib Storage.
package storage

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/goravel/framework/contracts/filesystem"
	"github.com/goravel/framework/facades"
	"github.com/spotlibs/go-lib/stderr"
)

const (
	NFS         = "NFS"
	MINIO       = "MINIO"
	MINIO_BRIMEN = "MINIO_BRIMEN"
)

// Storage is the single entry point for storage operations. Ported 1:1 from
// php-lib Storage.
type Storage struct {
	pendingDriver     string // "" means unset (PHP null)
	pendingAutoDetect bool
	pendingFromDriver string // "" means unset
	pendingToDriver   string // "" means unset
}

// New returns a fresh Storage instance.
func New() *Storage {
	return &Storage{}
}

// Driver sets the explicit driver name for the pending operation.
// Returns the Storage instance for fluent chaining.
func (s *Storage) Driver(driver string) *Storage {
	s.pendingDriver = driver
	return s
}

// AutoDetect enables auto-detection of the driver for the pending operation.
// Returns the Storage instance for fluent chaining.
func (s *Storage) AutoDetect() *Storage {
	s.pendingAutoDetect = true
	return s
}

// FromDriver sets the source driver for cross-driver copy/move operations.
// Returns the Storage instance for fluent chaining.
func (s *Storage) FromDriver(driver string) *Storage {
	s.pendingFromDriver = driver
	return s
}

// ToDriver sets the destination driver for cross-driver copy/move operations.
// Returns the Storage instance for fluent chaining.
func (s *Storage) ToDriver(driver string) *Storage {
	s.pendingToDriver = driver
	return s
}

// reset clears pending state. Used after read/simple operations.
// Partial reset: clears only pendingDriver and pendingAutoDetect.
func (s *Storage) resetPartial() {
	s.pendingDriver = ""
	s.pendingAutoDetect = false
}

// resetFull clears all pending state. Used after copy/move/zip operations.
func (s *Storage) resetFull() {
	s.pendingDriver = ""
	s.pendingAutoDetect = false
	s.pendingFromDriver = ""
	s.pendingToDriver = ""
}

// makeDriver returns a concrete StorageDriver implementation for the given driver name.
// Ported 1:1 from php-lib Storage::makeDriver.
func (s *Storage) makeDriver(driverName string) (StorageDriver, error) {
	switch driverName {
	case NFS:
		return newNfsDriver(), nil
	case MINIO:
		return newMinioDriver("minio"), nil
	case MINIO_BRIMEN:
		return newMinioDriver("minio_brimen"), nil
	default:
		return nil, stderr.ErrRuntime("Unknown storage driver: " + driverName)
	}
}

// ReadStream returns a readable stream for a file. Supports auto-detect,
// explicit driver, and explicit source driver (for future multi-driver).
// Ported 1:1 from php-lib Storage::readStream.
func (s *Storage) ReadStream(filepath string) (io.ReadCloser, error) {
	defer s.resetPartial()

	if s.pendingAutoDetect {
		// Auto-detect: try each driver in order, return first that exists
		for _, driverName := range []string{MINIO, MINIO_BRIMEN, NFS} {
			driver, err := s.makeDriver(driverName)
			if err != nil {
				continue
			}
			if driver.Exists(filepath) {
				return driver.ReadStream(filepath)
			}
		}
		return nil, stderr.ErrRuntime("File not found: " + filepath)
	}

	// Explicit or default driver resolution
	driverName := s.pendingDriver
	if driverName == "" {
		var err error
		driverName, err = driverResolver{}.resolveDefault()
		if err != nil {
			return nil, err
		}
	} else {
		var err error
		driverName, err = driverResolver{}.resolveExplicit(driverName)
		if err != nil {
			return nil, err
		}
	}

	driver, err := s.makeDriver(driverName)
	if err != nil {
		return nil, err
	}

	return driver.ReadStream(filepath)
}

// Exists checks if a file exists. Returns false if the driver cannot be resolved.
// Ported 1:1 from php-lib Storage::exists.
func (s *Storage) Exists(filepath string) bool {
	defer s.resetPartial()

	if s.pendingAutoDetect {
		// Auto-detect: try each driver in order, return true on first hit
		for _, driverName := range []string{MINIO, MINIO_BRIMEN, NFS} {
			driver, err := s.makeDriver(driverName)
			if err != nil {
				continue
			}
			if driver.Exists(filepath) {
				return true
			}
		}
		return false
	}

	// Explicit or default driver resolution
	driverName := s.pendingDriver
	if driverName == "" {
		var err error
		driverName, err = driverResolver{}.resolveDefault()
		if err != nil {
			// PHP would throw; Go returns false per interface contract
			return false
		}
	} else {
		var err error
		driverName, err = driverResolver{}.resolveExplicit(driverName)
		if err != nil {
			return false
		}
	}

	driver, err := s.makeDriver(driverName)
	if err != nil {
		return false
	}

	return driver.Exists(filepath)
}

// Upload uploads a file to storage. Requires explicit driver or default.
// AutoDetect is not allowed for upload operations.
// Ported 1:1 from php-lib Storage::upload.
func (s *Storage) Upload(file filesystem.File, dirpath string, filename string) (*Result, error) {
	defer s.resetPartial()

	if s.pendingAutoDetect {
		return nil, stderr.ErrRuntime("autoDetect is not allowed for upload")
	}

	// Explicit or default driver resolution
	driverName := s.pendingDriver
	if driverName == "" {
		var err error
		driverName, err = driverResolver{}.resolveDefault()
		if err != nil {
			return nil, err
		}
	} else {
		var err error
		driverName, err = driverResolver{}.resolveExplicit(driverName)
		if err != nil {
			return nil, err
		}
	}

	driver, err := s.makeDriver(driverName)
	if err != nil {
		return nil, err
	}

	return driver.Upload(file, dirpath, filename)
}

// Delete deletes a file from storage. Requires explicit driver.
// AutoDetect and defaulting are not allowed.
// DELETE_STORAGE_SPOTLIB env must be true to allow deletion.
// Ported 1:1 from php-lib Storage::delete.
func (s *Storage) Delete(filepath string) error {
	// Check deletion permission first (before setting up defer for resetPartial)
	if !facades.Config().GetBool("ALLOW_DELETE_STORAGE_SPOTLIB", false) {
		return stderr.ErrRuntime("disallow action")
	}

	defer s.resetPartial()

	if s.pendingAutoDetect {
		return stderr.ErrRuntime("autoDetect is not allowed for delete")
	}

	if s.pendingDriver == "" {
		return stderr.ErrRuntime("delete requires explicit driver. Call ->driver()")
	}

	driverName, err := driverResolver{}.resolveExplicit(s.pendingDriver)
	if err != nil {
		return err
	}

	driver, err := s.makeDriver(driverName)
	if err != nil {
		return err
	}

	return driver.Delete(filepath)
}

// Securelink generates a secure temporary URL for a file.
// Supports auto-detect, explicit driver.
// Ported 1:1 from php-lib Storage::securelink.
func (s *Storage) Securelink(filepath string, ttl *int) (string, error) {
	defer s.resetPartial()

	if s.pendingAutoDetect {
		// Auto-detect: try each driver in order, return first that exists
		for _, driverName := range []string{MINIO, MINIO_BRIMEN, NFS} {
			driver, err := s.makeDriver(driverName)
			if err != nil {
				continue
			}
			if driver.Exists(filepath) {
				return driver.Securelink(filepath, ttl)
			}
		}
		return "", stderr.ErrRuntime("File not found: " + filepath)
	}

	// Explicit or default driver resolution
	driverName := s.pendingDriver
	if driverName == "" {
		var err error
		driverName, err = driverResolver{}.resolveDefault()
		if err != nil {
			return "", err
		}
	} else {
		var err error
		driverName, err = driverResolver{}.resolveExplicit(driverName)
		if err != nil {
			return "", err
		}
	}

	driver, err := s.makeDriver(driverName)
	if err != nil {
		return "", err
	}

	return driver.Securelink(filepath, ttl)
}

// resolveNamed returns the validated driver name: explicit when set, otherwise
// the configured default. Mirrors the PHP explicit-or-default resolution.
func (s *Storage) resolveNamed() (string, error) {
	if s.pendingDriver != "" {
		return driverResolver{}.resolveExplicit(s.pendingDriver)
	}
	return driverResolver{}.resolveDefault()
}

// Info returns metadata for a file. Uses the explicit driver when set,
// otherwise probes MINIO, MINIO_BRIMEN, NFS in order.
// Ported 1:1 from php-lib Storage::info.
func (s *Storage) Info(filepath string) (*Result, error) {
	defer s.resetPartial()

	if s.pendingDriver != "" {
		driverName, err := driverResolver{}.resolveExplicit(s.pendingDriver)
		if err != nil {
			return nil, err
		}
		driver, err := s.makeDriver(driverName)
		if err != nil {
			return nil, err
		}
		return driver.Info(filepath)
	}

	for _, driverName := range []string{MINIO, MINIO_BRIMEN, NFS} {
		driver, err := s.makeDriver(driverName)
		if err != nil {
			return nil, err
		}
		if driver.Exists(filepath) {
			return driver.Info(filepath)
		}
	}

	return nil, stderr.ErrRuntime("File not found: " + filepath)
}

// AllFiles returns a recursive list of files in a directory on the resolved
// driver. AutoDetect is not allowed.
// Ported 1:1 from php-lib Storage::allFiles.
func (s *Storage) AllFiles(dirpath string) ([]FileEntry, error) {
	defer s.resetPartial()

	if s.pendingAutoDetect {
		return nil, stderr.ErrRuntime("autoDetect is not allowed for allFiles")
	}

	driverName, err := s.resolveNamed()
	if err != nil {
		return nil, err
	}

	driver, err := s.makeDriver(driverName)
	if err != nil {
		return nil, err
	}

	return driver.AllFiles(dirpath)
}

// validateCopyMove runs the three pre-checks shared by Copy and Move. The
// "copy()" wording is intentionally verbatim from PHP, including for Move.
func (s *Storage) validateCopyMove() error {
	if s.pendingDriver != "" && (s.pendingFromDriver != "" || s.pendingToDriver != "") {
		return stderr.ErrRuntime("copy() cannot mix ->driver() with ->fromDriver()/->toDriver()")
	}

	if s.pendingDriver == "" && s.pendingFromDriver == "" && s.pendingToDriver == "" {
		return stderr.ErrRuntime("copy/move requires explicit driver. Call ->driver() or ->fromDriver() + ->toDriver()")
	}

	if (s.pendingFromDriver == "") != (s.pendingToDriver == "") {
		return stderr.ErrRuntime("copy() cross-driver requires both ->fromDriver() and ->toDriver()")
	}

	return nil
}

// Copy copies a file, same-driver (->Driver) or cross-driver
// (->FromDriver + ->ToDriver). Like PHP, pending state is reset only on the
// success paths, not on validation/operation errors.
// Ported 1:1 from php-lib Storage::copy.
func (s *Storage) Copy(srcPath, destPath string) (*Result, error) {
	if err := s.validateCopyMove(); err != nil {
		return nil, err
	}

	if s.pendingDriver != "" {
		driverName, err := driverResolver{}.resolveExplicit(s.pendingDriver)
		if err != nil {
			return nil, err
		}
		driver, err := s.makeDriver(driverName)
		if err != nil {
			return nil, err
		}
		result, err := driver.CopySameDriver(srcPath, destPath)
		if err != nil {
			return nil, err
		}
		s.resetFull()
		return result, nil
	}

	fromName, err := driverResolver{}.resolveExplicit(s.pendingFromDriver)
	if err != nil {
		return nil, err
	}
	toName, err := driverResolver{}.resolveExplicit(s.pendingToDriver)
	if err != nil {
		return nil, err
	}
	fromDriver, err := s.makeDriver(fromName)
	if err != nil {
		return nil, err
	}
	toDriver, err := s.makeDriver(toName)
	if err != nil {
		return nil, err
	}

	stream, err := fromDriver.ReadStream(srcPath)
	if err != nil {
		return nil, err
	}
	// Go-only addition: PHP relies on GC to release the stream.
	defer stream.Close()

	result, err := toDriver.WriteStream(stream, destPath)
	if err != nil {
		return nil, err
	}
	s.resetFull()
	return result, nil
}

// Move moves a file, same-driver or cross-driver. Cross-driver is Copy
// followed by a Delete on the source driver.
// Ported 1:1 from php-lib Storage::move.
func (s *Storage) Move(srcPath, destPath string) (*Result, error) {
	if err := s.validateCopyMove(); err != nil {
		return nil, err
	}

	if s.pendingDriver != "" {
		driverName, err := driverResolver{}.resolveExplicit(s.pendingDriver)
		if err != nil {
			return nil, err
		}
		driver, err := s.makeDriver(driverName)
		if err != nil {
			return nil, err
		}
		result, err := driver.MoveSameDriver(srcPath, destPath)
		if err != nil {
			return nil, err
		}
		s.resetFull()
		return result, nil
	}

	// Resolve the source driver BEFORE Copy: Copy resets the pending state.
	fromName, err := driverResolver{}.resolveExplicit(s.pendingFromDriver)
	if err != nil {
		return nil, err
	}
	sourceDriver, err := s.makeDriver(fromName)
	if err != nil {
		return nil, err
	}

	result, err := s.Copy(srcPath, destPath)
	if err != nil {
		return nil, err
	}

	if err := sourceDriver.Delete(srcPath); err != nil {
		return nil, stderr.ErrRuntime("move() copy succeeded but delete of source failed: " + err.Error())
	}

	s.resetFull()
	return result, nil
}

// ZipSource is one entry for BuildZip. Ported from PHP
// ['path_file'=>..., 'driver'=>..|null, 'zip_path'=>...].
type ZipSource struct {
	PathFile string
	Driver   string // "" means auto-detect (PHP null)
	ZipPath  string
}

// BuildZip builds a zip from the given sources and uploads it to destPath on
// the destination driver (->ToDriver or the default). If destPath already
// exists on the destination, it is returned without rebuilding.
// Ported 1:1 from php-lib Storage::buildZip.
func (s *Storage) BuildZip(sourceFiles []ZipSource, destPath string) (*Result, error) {
	defer s.resetFull()

	if len(sourceFiles) == 0 {
		return nil, stderr.ErrRuntime("buildZip requires at least one source file")
	}

	seen := make(map[string]struct{}, len(sourceFiles))
	for _, item := range sourceFiles {
		if _, dup := seen[item.ZipPath]; dup {
			return nil, stderr.ErrRuntime("buildZip has duplicate zip_path: " + item.ZipPath)
		}
		seen[item.ZipPath] = struct{}{}
	}

	var destName string
	var err error
	if s.pendingToDriver != "" {
		destName, err = driverResolver{}.resolveExplicit(s.pendingToDriver)
	} else {
		destName, err = driverResolver{}.resolveDefault()
	}
	if err != nil {
		return nil, err
	}
	destinationDriver, err := s.makeDriver(destName)
	if err != nil {
		return nil, err
	}

	if destinationDriver.Exists(destPath) {
		return &Result{
			Driver:   destName,
			PathFile: filepath.Base(destPath),
			FullPath: destPath,
			Folder:   filepath.Dir(destPath) + "/",
		}, nil
	}

	tempDir, err := os.MkdirTemp("", "zip_")
	if err != nil {
		return nil, stderr.ErrRuntime("Failed to create zip on pod")
	}
	defer os.RemoveAll(tempDir)

	zipPath := filepath.Join(tempDir, "final.zip")
	if err := s.writeZipArchive(sourceFiles, zipPath); err != nil {
		return nil, err
	}

	archive, err := os.Open(zipPath)
	if err != nil {
		return nil, stderr.ErrRuntime("Failed to open zip for upload")
	}
	defer archive.Close()

	return destinationDriver.WriteStream(archive, destPath)
}

// writeZipArchive streams every source into a new zip file at zipPath.
// Fails fast on the first source error, wrapping it with the source path.
func (s *Storage) writeZipArchive(sourceFiles []ZipSource, zipPath string) error {
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return stderr.ErrRuntime("Failed to create zip on pod")
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)

	for _, item := range sourceFiles {
		if err := s.addZipEntry(zipWriter, item); err != nil {
			_ = zipWriter.Close()
			return stderr.ErrRuntime("buildZip failed to stream " + item.PathFile + ": " + err.Error())
		}
	}

	if err := zipWriter.Close(); err != nil {
		return stderr.ErrRuntime("Failed to create zip on pod")
	}

	return nil
}

// addZipEntry resolves the source driver for one item, reads it and copies it
// into the archive under item.ZipPath.
func (s *Storage) addZipEntry(zipWriter *zip.Writer, item ZipSource) error {
	sourceDriver, err := s.resolveZipSource(item)
	if err != nil {
		return err
	}

	stream, err := sourceDriver.ReadStream(item.PathFile)
	if err != nil {
		return err
	}
	if stream == nil {
		return stderr.ErrRuntime("Failed to open read stream for: " + item.PathFile)
	}
	defer stream.Close()

	entry, err := zipWriter.Create(item.ZipPath)
	if err != nil {
		return stderr.ErrRuntime("Failed to add file to zip: " + item.PathFile)
	}

	if _, err := io.Copy(entry, stream); err != nil {
		return stderr.ErrRuntime("Failed to write stream to pod: " + item.PathFile)
	}

	return nil
}

// resolveZipSource picks the driver for a zip source: the item's explicit
// driver, or auto-detect probing MINIO, MINIO_BRIMEN, NFS.
func (s *Storage) resolveZipSource(item ZipSource) (StorageDriver, error) {
	if item.Driver != "" {
		name, err := driverResolver{}.resolveExplicit(item.Driver)
		if err != nil {
			return nil, err
		}
		return s.makeDriver(name)
	}

	for _, name := range []string{MINIO, MINIO_BRIMEN, NFS} {
		candidate, err := s.makeDriver(name)
		if err != nil {
			return nil, err
		}
		if candidate.Exists(item.PathFile) {
			return candidate, nil
		}
	}

	return nil, stderr.ErrRuntime("File not found: " + item.PathFile)
}

// BuildZipFolder zips an entire folder from a single source driver
// (->FromDriver or the default; never auto-detect), preserving relative
// subfolder structure, and uploads it via BuildZip.
// Ported 1:1 from php-lib Storage::buildZipFolder.
func (s *Storage) BuildZipFolder(sourceFolder, destPath string) (*Result, error) {
	defer s.resetFull()

	var sourceName string
	var err error
	if s.pendingFromDriver != "" {
		sourceName, err = driverResolver{}.resolveExplicit(s.pendingFromDriver)
	} else {
		sourceName, err = driverResolver{}.resolveDefault()
	}
	if err != nil {
		return nil, err
	}

	sourceDriver, err := s.makeDriver(sourceName)
	if err != nil {
		return nil, err
	}

	files, err := sourceDriver.AllFiles(sourceFolder)
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, stderr.ErrRuntime("buildZipFolder found no files in " + sourceFolder)
	}

	trimmedFolder := strings.TrimRight(sourceFolder, "/")
	sourcePrefix := trimmedFolder + "/"
	sources := make([]ZipSource, 0, len(files))
	for _, file := range files {
		var zipPath string
		if strings.HasPrefix(file.Path, sourcePrefix) {
			zipPath = file.Path[len(sourcePrefix):]
		} else {
			zipPath = strings.TrimLeft(file.Path[min(len(trimmedFolder), len(file.Path)):], "/")
		}
		sources = append(sources, ZipSource{PathFile: file.Path, Driver: sourceName, ZipPath: zipPath})
	}

	// pendingToDriver is still set on s, so BuildZip resolves the destination from it.
	return s.BuildZip(sources, destPath)
}
