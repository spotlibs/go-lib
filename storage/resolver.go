package storage

import (
	"github.com/goravel/framework/facades"
	"github.com/spotlibs/go-lib/stderr"
)

type driverResolver struct{}

// resolveDefault reads env('DEFAULT_DRIVER_STORAGE') and validates against the 3 known drivers.
func (r driverResolver) resolveDefault() (string, error) {
	driver := facades.Config().Env("DEFAULT_DRIVER_STORAGE", "")
	if driver == nil || driver == "" {
		return "", stderr.ErrRuntime("DEFAULT_DRIVER_STORAGE is not configured")
	}

	driverStr, ok := driver.(string)
	if !ok {
		driverStr = ""
	}

	if driverStr == "" {
		return "", stderr.ErrRuntime("DEFAULT_DRIVER_STORAGE is not configured")
	}

	// Validate against known drivers
	if driverStr != NFS && driverStr != MINIO && driverStr != MINIO_BRIMEN {
		return "", stderr.ErrRuntime("DEFAULT_DRIVER_STORAGE has invalid value: " + driverStr)
	}

	return driverStr, nil
}

// resolveExplicit validates that the driver is one of the 3 known drivers.
func (r driverResolver) resolveExplicit(driver string) (string, error) {
	if driver != NFS && driver != MINIO && driver != MINIO_BRIMEN {
		return "", stderr.ErrRuntime("Unknown storage driver: " + driver)
	}

	return driver, nil
}
