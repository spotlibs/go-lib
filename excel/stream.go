package excel

import (
	"github.com/spotlibs/go-lib/stderr"
	"github.com/xuri/excelize/v2"
)

// StreamWriter wraps excelize.StreamWriter for row-by-row writing on large sheets.
type StreamWriter struct {
	sw *excelize.StreamWriter
}

// NewStreamWriter creates a StreamWriter for the given sheet name.
// Use StreamWriter.SetRow to write rows, then call Flush when done.
// Returns error if the sheet does not exist or creation fails.
func (e *ExcelFile) NewStreamWriter(sheet string) (*StreamWriter, error) {
	sw, err := e.f.NewStreamWriter(sheet)
	if err != nil {
		return nil, stderr.ErrRuntime("failed to create stream writer for sheet: " + sheet)
	}
	return &StreamWriter{sw: sw}, nil
}

// SetRow writes a row of values at the given cell anchor.
// cell: top-left cell of the row e.g. "A1", "A2".
// values: slice of any values to write across columns starting at cell's column.
// Returns error if writing fails.
func (sw *StreamWriter) SetRow(cell string, values []interface{}) error {
	cells := make([]interface{}, len(values))
	copy(cells, values)
	if err := sw.sw.SetRow(cell, cells); err != nil {
		return stderr.ErrRuntime("failed to set row at " + cell)
	}
	return nil
}

// Flush finalizes the stream and writes all buffered data into the workbook.
// Must be called exactly once after all rows have been written.
// Do not call SetRow again after Flush.
// Returns error if flushing fails.
func (sw *StreamWriter) Flush() error {
	if err := sw.sw.Flush(); err != nil {
		return stderr.ErrRuntime("failed to flush stream writer")
	}
	return nil
}
