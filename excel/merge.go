package excel

import "github.com/spotlibs/go-lib/stderr"

// MergeCell merges cells from topLeftCell to bottomRightCell.
// sheet: sheet name. topLeftCell: e.g. "A1". bottomRightCell: e.g. "D1".
// Returns error if merge fails.
func (e *ExcelFile) MergeCell(sheet, topLeftCell, bottomRightCell string) error {
	if err := e.f.MergeCell(sheet, topLeftCell, bottomRightCell); err != nil {
		return stderr.ErrRuntime("failed to merge cells " + sheet + "!" + topLeftCell + ":" + bottomRightCell)
	}
	return nil
}

// UnmergeCell unmerges cells that were previously merged in the given range.
// sheet: sheet name. topLeftCell: e.g. "A1". bottomRightCell: e.g. "D1".
// Returns error if unmerge fails.
func (e *ExcelFile) UnmergeCell(sheet, topLeftCell, bottomRightCell string) error {
	if err := e.f.UnmergeCell(sheet, topLeftCell, bottomRightCell); err != nil {
		return stderr.ErrRuntime("failed to unmerge cells " + sheet + "!" + topLeftCell + ":" + bottomRightCell)
	}
	return nil
}
