package excel

import (
	"github.com/spotlibs/go-lib/stderr"
	"github.com/xuri/excelize/v2"
)

// Style is a re-export of excelize.Style.
// Callers build a Style value and pass a pointer to NewStyle.
type Style = excelize.Style

// NewStyle creates a new cell style from the given Style definition.
// Returns the style ID (int) to be used with SetCellStyle.
// Returns error if style creation fails.
func (e *ExcelFile) NewStyle(style *Style) (int, error) {
	id, err := e.f.NewStyle(style)
	if err != nil {
		return 0, stderr.ErrRuntime("failed to create style")
	}
	return id, nil
}

// SetCellStyle applies a style (by ID from NewStyle) to a range of cells.
// sheet: sheet name. topLeftCell: top-left cell e.g. "A1".
// bottomRightCell: bottom-right cell e.g. "D10". styleID: ID returned by NewStyle.
// Returns error if applying fails.
func (e *ExcelFile) SetCellStyle(sheet, topLeftCell, bottomRightCell string, styleID int) error {
	if err := e.f.SetCellStyle(sheet, topLeftCell, bottomRightCell, styleID); err != nil {
		return stderr.ErrRuntime("failed to set cell style on " + sheet + "!" + topLeftCell + ":" + bottomRightCell)
	}
	return nil
}
