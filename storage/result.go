package storage

// Result is the result DTO for storage operations. Ported 1:1 from
// php-lib StorageResult.
type Result struct {
	Driver   string
	Folder   string
	PathFile string
	FullPath string
	Size     *int64 // nil when unknown (PHP's empty-string default); set to byte count when known.
}
