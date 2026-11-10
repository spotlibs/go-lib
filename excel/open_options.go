package excel

import (
	"io"

	"github.com/spotlibs/go-lib/stderr"
	"github.com/xuri/excelize/v2"
)

// OpenOptions configures how a workbook is opened for reading.
// A zero value for any field means "use the excelize default".
type OpenOptions struct {
	// Password is used to open a password-protected workbook.
	Password string
	// UnzipSizeLimit is the maximum total uncompressed size (bytes) of the workbook.
	// Must be >= UnzipXMLSizeLimit. Set explicitly when opening user uploads.
	UnzipSizeLimit int64
	// UnzipXMLSizeLimit is the maximum uncompressed size (bytes) of a single XML part.
	// Set explicitly when opening user uploads.
	UnzipXMLSizeLimit int64
}

// toExcelize converts OpenOptions to excelize.Options.
func (o OpenOptions) toExcelize() excelize.Options {
	return excelize.Options{
		Password:          o.Password,
		UnzipSizeLimit:    o.UnzipSizeLimit,
		UnzipXMLSizeLimit: o.UnzipXMLSizeLimit,
	}
}

// OpenWithOptions opens an existing Excel file from the given file path using opts.
// Returns error if the file cannot be opened, the password is wrong, or a size limit is exceeded.
func OpenWithOptions(filePath string, opts OpenOptions) (*ExcelFile, error) {
	f, err := excelize.OpenFile(filePath, opts.toExcelize())
	if err != nil {
		return nil, stderr.ErrRuntime("failed to open excel file: " + filePath)
	}
	return &ExcelFile{f: f}, nil
}

// OpenReaderWithOptions opens an existing Excel file from an io.Reader using opts.
// Returns error if the reader cannot be parsed, the password is wrong, or a size limit is exceeded.
func OpenReaderWithOptions(r io.Reader, opts OpenOptions) (*ExcelFile, error) {
	f, err := excelize.OpenReader(r, opts.toExcelize())
	if err != nil {
		return nil, stderr.ErrRuntime("failed to open excel from reader")
	}
	return &ExcelFile{f: f}, nil
}
