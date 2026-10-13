package excel

import (
	"bytes"

	"github.com/spotlibs/go-lib/stderr"
)

// SaveAs writes the Excel file to disk at the given filePath.
// Returns error if writing fails.
func (e *ExcelFile) SaveAs(filePath string) error {
	if err := e.f.SaveAs(filePath); err != nil {
		return stderr.ErrRuntime("failed to save excel file: " + filePath)
	}
	return nil
}

// WriteToBuffer writes the Excel file into a bytes.Buffer and returns it.
// Returns error if serialization fails.
func (e *ExcelFile) WriteToBuffer() (*bytes.Buffer, error) {
	buf, err := e.f.WriteToBuffer()
	if err != nil {
		return nil, stderr.ErrRuntime("failed to write excel to buffer")
	}
	return buf, nil
}
