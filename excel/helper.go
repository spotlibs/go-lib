package excel

import (
	"github.com/spotlibs/go-lib/stderr"
	"github.com/xuri/excelize/v2"
)

// CellNameToCoordinates converts a cell name (e.g. "A1") to 1-based column and row numbers.
// Returns (col, row, error). Both col and row are 1-based integers.
func CellNameToCoordinates(cell string) (int, int, error) {
	col, row, err := excelize.CellNameToCoordinates(cell)
	if err != nil {
		return 0, 0, stderr.ErrRuntime("invalid cell name: " + cell)
	}
	return col, row, nil
}

// CoordinatesToCellName converts 1-based column and row numbers to a cell name (e.g. "A1").
// col: 1-based column number. row: 1-based row number.
// Returns cell name string and error.
func CoordinatesToCellName(col, row int) (string, error) {
	name, err := excelize.CoordinatesToCellName(col, row)
	if err != nil {
		return "", stderr.ErrRuntime("invalid coordinates")
	}
	return name, nil
}
