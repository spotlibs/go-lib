package storage

import "io"

// ReadStreamFromDriver opens a read stream for srcPath on the specified driver.
// driverName: one of storage.NFS, storage.MINIO, storage.MINIO_BRIMEN.
// srcPath: full source path including filename.
// The caller is responsible for closing the returned stream.
func ReadStreamFromDriver(driverName string, srcPath string) (io.ReadCloser, error) {
	s := &Storage{}
	driver, err := s.makeDriver(driverName)
	if err != nil {
		return nil, err
	}
	return driver.ReadStream(srcPath)
}
