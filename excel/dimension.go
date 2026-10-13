package excel

import "github.com/spotlibs/go-lib/stderr"

// SetRowHeight sets the height (in points) of a row.
// sheet: sheet name. row: 1-based row number. height: height in points.
// Returns error if setting fails.
func (e *ExcelFile) SetRowHeight(sheet string, row int, height float64) error {
	if err := e.f.SetRowHeight(sheet, row, height); err != nil {
		return stderr.ErrRuntime("failed to set row height")
	}
	return nil
}

// SetColWidth sets the width of columns from startCol to endCol (inclusive).
// sheet: sheet name. startCol: column letter e.g. "A". endCol: column letter e.g. "D". width: width in character units.
// Returns error if setting fails.
func (e *ExcelFile) SetColWidth(sheet, startCol, endCol string, width float64) error {
	if err := e.f.SetColWidth(sheet, startCol, endCol, width); err != nil {
		return stderr.ErrRuntime("failed to set col width")
	}
	return nil
}

// GetRowHeight returns the height (in points) of a row.
// sheet: sheet name. row: 1-based row number.
// Returns height and error.
func (e *ExcelFile) GetRowHeight(sheet string, row int) (float64, error) {
	h, err := e.f.GetRowHeight(sheet, row)
	if err != nil {
		return 0, stderr.ErrRuntime("failed to get row height")
	}
	return h, nil
}

// GetColWidth returns the width of a column.
// sheet: sheet name. col: column letter e.g. "A".
// Returns width and error.
func (e *ExcelFile) GetColWidth(sheet, col string) (float64, error) {
	w, err := e.f.GetColWidth(sheet, col)
	if err != nil {
		return 0, stderr.ErrRuntime("failed to get col width")
	}
	return w, nil
}
