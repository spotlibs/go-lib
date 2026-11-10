package excel

import (
	"strconv"

	"github.com/spotlibs/go-lib/stderr"
	"github.com/xuri/excelize/v2"
)

// RowReader reads a sheet row by row without loading the whole sheet into memory.
// Typical loop: for r.Next() { cols, err := r.Columns() ... }, then check r.Error(), then Close().
// RowNum() is the 1-based Excel row number of the current row. Rows missing from the file
// between two stored rows are still visited and return an empty (nil) slice, so RowNum()
// always matches the real row.
// Always call Close. Close the RowReader before closing the ExcelFile.
// Returned slices can have different lengths (trailing empty cells are dropped).
// Use PadRow if a fixed width is needed.
type RowReader struct {
	rows   *excelize.Rows
	rowNum int
}

// NewRowReader creates a RowReader for the given sheet name.
// Always call Close on the returned RowReader.
// Returns stderr.ErrNotFound if the sheet does not exist.
func (e *ExcelFile) NewRowReader(sheet string) (*RowReader, error) {
	rows, err := e.f.Rows(sheet)
	if err != nil {
		return nil, readErr(err, sheet, "failed to create row reader for sheet: "+sheet)
	}
	return &RowReader{rows: rows}, nil
}

// Next advances to the next row. Returns false when there are no more rows.
func (r *RowReader) Next() bool {
	if r.rows.Next() {
		r.rowNum++
		return true
	}
	return false
}

// RowNum returns the 1-based Excel row number of the current row.
// Returns 0 before the first call to Next.
func (r *RowReader) RowNum() int {
	return r.rowNum
}

// Columns returns the display-formatted cell values of the current row.
// Returns an empty slice for a row that has no cells.
func (r *RowReader) Columns() ([]string, error) {
	cols, err := r.rows.Columns()
	if err != nil {
		return nil, stderr.ErrRuntime("failed to read row " + strconv.Itoa(r.rowNum))
	}
	return cols, nil
}

// ColumnsRaw returns the stored cell values of the current row without display formatting.
// Returns an empty slice for a row that has no cells.
func (r *RowReader) ColumnsRaw() ([]string, error) {
	cols, err := r.rows.Columns(excelize.Options{RawCellValue: true})
	if err != nil {
		return nil, stderr.ErrRuntime("failed to read row " + strconv.Itoa(r.rowNum))
	}
	return cols, nil
}

// Error returns an error if iteration stopped because of a read failure.
// Call it after the Next loop ends.
func (r *RowReader) Error() error {
	if err := r.rows.Error(); err != nil {
		return stderr.ErrRuntime("failed while reading rows")
	}
	return nil
}

// Close releases resources held by the RowReader.
// Always call Close, and do it before closing the ExcelFile.
func (r *RowReader) Close() error {
	return r.rows.Close()
}
