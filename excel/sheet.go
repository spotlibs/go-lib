package excel

import "github.com/spotlibs/go-lib/stderr"

// NewSheet creates a new sheet with the given name.
// Returns the sheet index (0-based) and error.
func (e *ExcelFile) NewSheet(name string) (int, error) {
	idx, err := e.f.NewSheet(name)
	if err != nil {
		return 0, stderr.ErrRuntime("failed to create sheet: " + name)
	}
	return idx, nil
}

// DeleteSheet deletes the sheet with the given name.
// Returns error if deletion fails.
func (e *ExcelFile) DeleteSheet(name string) error {
	if err := e.f.DeleteSheet(name); err != nil {
		return stderr.ErrRuntime("failed to delete sheet: " + name)
	}
	return nil
}

// CopySheet copies a sheet by index from srcIndex to destIndex.
// Returns error if copy fails.
func (e *ExcelFile) CopySheet(srcIndex, destIndex int) error {
	if err := e.f.CopySheet(srcIndex, destIndex); err != nil {
		return stderr.ErrRuntime("failed to copy sheet")
	}
	return nil
}

// SetActiveSheet sets the active (selected) sheet by its 0-based index.
func (e *ExcelFile) SetActiveSheet(index int) {
	e.f.SetActiveSheet(index)
}

// GetSheetIndex returns the 0-based index of the sheet by name.
// Returns -1 and error if not found.
func (e *ExcelFile) GetSheetIndex(name string) (int, error) {
	idx, err := e.f.GetSheetIndex(name)
	if err != nil {
		return -1, stderr.ErrRuntime("failed to get sheet index: " + name)
	}
	return idx, nil
}

// GetSheetList returns all sheet names in workbook order.
func (e *ExcelFile) GetSheetList() []string {
	return e.f.GetSheetList()
}
