package excel

import "github.com/spotlibs/go-lib/stderr"

// SetCellValue sets the value of a cell.
// sheet: sheet name. cell: cell reference ex. "A1". value: any Go value.
// Returns error if setting fails.
func (e *ExcelFile) SetCellValue(sheet, cell string, value interface{}) error {
	if err := e.f.SetCellValue(sheet, cell, value); err != nil {
		return stderr.ErrRuntime("failed to set cell value at " + sheet + "!" + cell)
	}
	return nil
}

// GetCellValue returns the string value of a cell.
// sheet: sheet name. cell: cell reference e.g. "A1".
// Returns the cell value as string and error.
func (e *ExcelFile) GetCellValue(sheet, cell string) (string, error) {
	val, err := e.f.GetCellValue(sheet, cell)
	if err != nil {
		return "", stderr.ErrRuntime("failed to get cell value at " + sheet + "!" + cell)
	}
	return val, nil
}

// SetCellFormula sets a formula string on a cell.
// sheet: sheet name. cell: cell reference e.g. "A1". formula: Excel formula string e.g. "=SUM(A1:A10)".
// Returns error if setting fails.
func (e *ExcelFile) SetCellFormula(sheet, cell, formula string) error {
	if err := e.f.SetCellFormula(sheet, cell, formula); err != nil {
		return stderr.ErrRuntime("failed to set formula at " + sheet + "!" + cell)
	}
	return nil
}

// GetRows returns all rows in the sheet as a 2D string slice.
// sheet: sheet name.
// Returns [][]string where each inner slice is one row of cell values.
func (e *ExcelFile) GetRows(sheet string) ([][]string, error) {
	rows, err := e.f.GetRows(sheet)
	if err != nil {
		return nil, stderr.ErrRuntime("failed to get rows for sheet: " + sheet)
	}
	return rows, nil
}
