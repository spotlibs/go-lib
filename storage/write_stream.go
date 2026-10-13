package storage

import "io"

// WriteStreamToDriver writes the given stream to destPath on the specified driver.
// driverName: one of storage.NFS, storage.MINIO, storage.MINIO_BRIMEN.
// stream: an io.ReadCloser with the content to write. The driver closes it.
// destPath: full destination path including filename.
// Returns the Result DTO from the driver and error.
func WriteStreamToDriver(driverName string, stream io.ReadCloser, destPath string) (*Result, error) {
	s := &Storage{}
	driver, err := s.makeDriver(driverName)
	if err != nil {
		return nil, err
	}
	return driver.WriteStream(stream, destPath)
}
