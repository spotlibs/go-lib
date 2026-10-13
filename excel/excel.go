package excel

import (
	"io"

	"github.com/spotlibs/go-lib/stderr"
	"github.com/xuri/excelize/v2"
)

// ExcelFile wraps excelize.File and exposes a simplified API.
type ExcelFile struct {
	f *excelize.File
}

// New creates a new empty ExcelFile.
func New() *ExcelFile {
	return &ExcelFile{f: excelize.NewFile()}
}

// Open opens an existing Excel file from the given file path.
// Returns error if the file cannot be opened.
func Open(filePath string) (*ExcelFile, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, stderr.ErrRuntime("failed to open excel file: " + filePath)
	}
	return &ExcelFile{f: f}, nil
}

// OpenReader opens an existing Excel file from an io.Reader.
// Returns error if the reader cannot be parsed.
func OpenReader(r io.Reader) (*ExcelFile, error) {
	f, err := excelize.OpenReader(r)
	if err != nil {
		return nil, stderr.ErrRuntime("failed to open excel from reader")
	}
	return &ExcelFile{f: f}, nil
}

// Close releases resources held by the ExcelFile.
// Always call Close when done.
func (e *ExcelFile) Close() error {
	return e.f.Close()
}
