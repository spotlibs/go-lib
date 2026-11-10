package excel

import (
	"strings"

	"github.com/spotlibs/go-lib/stderr"
	"github.com/xuri/excelize/v2"
)

// SheetDimension is the used range of a sheet as declared in the file.
// All coordinates are 1-based. A zero value (Ref == "") means no dimension is declared.
type SheetDimension struct {
	Ref      string
	StartCol int
	StartRow int
	EndCol   int
	EndRow   int
}

// RowCount returns the number of rows in the declared range, or 0 if none is declared.
func (d SheetDimension) RowCount() int {
	if d.Ref == "" {
		return 0
	}
	return d.EndRow - d.StartRow + 1
}

// ColCount returns the number of columns in the declared range, or 0 if none is declared.
func (d SheetDimension) ColCount() int {
	if d.Ref == "" {
		return 0
	}
	return d.EndCol - d.StartCol + 1
}

// GetSheetDimension returns the used range declared for the sheet.
// The value comes from the file metadata and may be empty or inaccurate; use it as a hint only.
// Returns stderr.ErrNotFound if the sheet does not exist.
func (e *ExcelFile) GetSheetDimension(sheet string) (SheetDimension, error) {
	ref, err := e.f.GetSheetDimension(sheet)
	if err != nil {
		return SheetDimension{}, readErr(err, sheet, "failed to get sheet dimension: "+sheet)
	}
	if ref == "" {
		return SheetDimension{}, nil
	}

	parts := strings.Split(ref, ":")
	startCol, startRow, err := excelize.CellNameToCoordinates(parts[0])
	if err != nil {
		return SheetDimension{}, stderr.ErrRuntime("invalid sheet dimension: " + ref)
	}
	endCol, endRow := startCol, startRow
	if len(parts) > 1 {
		endCol, endRow, err = excelize.CellNameToCoordinates(parts[1])
		if err != nil {
			return SheetDimension{}, stderr.ErrRuntime("invalid sheet dimension: " + ref)
		}
	}

	return SheetDimension{
		Ref:      ref,
		StartCol: startCol,
		StartRow: startRow,
		EndCol:   endCol,
		EndRow:   endRow,
	}, nil
}
