package excel_test

import (
	"path/filepath"
	"testing"

	"github.com/spotlibs/go-lib/excel"
	"github.com/stretchr/testify/require"
)

// newWorkbook creates an in-memory workbook and closes it when the test ends.
func newWorkbook(t *testing.T) *excel.ExcelFile {
	t.Helper()
	f := excel.New()
	t.Cleanup(func() { _ = f.Close() })
	return f
}

// newSavedWorkbook writes the given cells to Sheet1, saves the workbook under
// t.TempDir and returns the file path.
func newSavedWorkbook(t *testing.T, cells map[string]interface{}) string {
	t.Helper()
	f := excel.New()
	for cell, value := range cells {
		require.NoError(t, f.SetCellValue("Sheet1", cell, value))
	}
	path := filepath.Join(t.TempDir(), "book.xlsx")
	require.NoError(t, f.SaveAs(path))
	require.NoError(t, f.Close())
	return path
}
