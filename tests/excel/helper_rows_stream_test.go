package excel_test

import (
	"testing"

	"github.com/spotlibs/go-lib/excel"
	"github.com/spotlibs/go-lib/stderr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCellNameToCoordinates_Valid(t *testing.T) {
	col, row, err := excel.CellNameToCoordinates("C3")

	require.NoError(t, err)
	assert.Equal(t, 3, col)
	assert.Equal(t, 3, row)
}

func TestCellNameToCoordinates_InvalidReturnsError(t *testing.T) {
	_, _, err := excel.CellNameToCoordinates("1A")

	assert.Error(t, err)
}

func TestCoordinatesToCellName_Valid(t *testing.T) {
	name, err := excel.CoordinatesToCellName(3, 3)

	require.NoError(t, err)
	assert.Equal(t, "C3", name)
}

func TestCoordinatesToCellName_InvalidReturnsError(t *testing.T) {
	name, err := excel.CoordinatesToCellName(0, 1)

	assert.Error(t, err)
	assert.Equal(t, "", name)
}

func TestPadRow_ShorterRowIsExtended(t *testing.T) {
	out := excel.PadRow([]string{"a"}, 3)

	assert.Equal(t, []string{"a", "", ""}, out)
}

func TestPadRow_LongerOrEqualRowIsUnchanged(t *testing.T) {
	in := []string{"a", "b"}

	assert.Equal(t, in, excel.PadRow(in, 2))
	assert.Equal(t, in, excel.PadRow(in, 1))
}

func TestPadRows_PadsEveryRow(t *testing.T) {
	out := excel.PadRows([][]string{{"a"}, {"b", "c"}}, 2)

	assert.Equal(t, [][]string{{"a", ""}, {"b", "c"}}, out)
}

func TestIsBlankRow_Cases(t *testing.T) {
	assert.True(t, excel.IsBlankRow(nil))
	assert.True(t, excel.IsBlankRow([]string{}))
	assert.True(t, excel.IsBlankRow([]string{"", "   ", "\t"}))
	assert.False(t, excel.IsBlankRow([]string{"", "x"}))
}

func TestRemoveBlankRows_KeepsOrder(t *testing.T) {
	out := excel.RemoveBlankRows([][]string{{"a"}, {"", " "}, {"b"}, nil})

	assert.Equal(t, [][]string{{"a"}, {"b"}}, out)
}

func TestGetSheetDimension_ReturnsDeclaredRef(t *testing.T) {
	path := newSavedWorkbook(t, map[string]interface{}{"A1": "x", "C5": "y"})
	f, err := excel.Open(path)
	require.NoError(t, err)
	defer f.Close()

	// Excelize reports the dimension stored in the file metadata, which is "A1" for this workbook.
	dim, err := f.GetSheetDimension("Sheet1")

	require.NoError(t, err)
	assert.Equal(t, "A1", dim.Ref)
	assert.Equal(t, 1, dim.StartRow)
	assert.Equal(t, 1, dim.RowCount())
	assert.Equal(t, 1, dim.ColCount())
}

func TestGetSheetDimension_MissingSheetReturnsNotFound(t *testing.T) {
	f := newWorkbook(t)

	_, err := f.GetSheetDimension("NoSuchSheet")

	assert.True(t, stderr.IsErrNotFound(err))
}

func TestSheetDimension_ZeroValueHasNoRowsOrColumns(t *testing.T) {
	var dim excel.SheetDimension

	assert.Equal(t, 0, dim.RowCount())
	assert.Equal(t, 0, dim.ColCount())
}

func TestRowReader_IteratesRowsIncludingGaps(t *testing.T) {
	f := newWorkbook(t)
	require.NoError(t, f.SetCellValue("Sheet1", "A1", "a"))
	require.NoError(t, f.SetCellValue("Sheet1", "B1", "b"))
	require.NoError(t, f.SetCellValue("Sheet1", "A2", "c"))
	require.NoError(t, f.SetCellValue("Sheet1", "A4", "d"))

	r, err := f.NewRowReader("Sheet1")
	require.NoError(t, err)
	defer r.Close()

	assert.Equal(t, 0, r.RowNum())
	var rowNums []int
	var rows [][]string
	for r.Next() {
		cols, err := r.Columns()
		require.NoError(t, err)
		rowNums = append(rowNums, r.RowNum())
		rows = append(rows, cols)
	}

	require.NoError(t, r.Error())
	assert.Equal(t, []int{1, 2, 3, 4}, rowNums)
	assert.Equal(t, []string{"a", "b"}, rows[0])
	assert.Equal(t, []string{"c"}, rows[1])
	assert.Empty(t, rows[2])
	assert.Equal(t, []string{"d"}, rows[3])
}

func TestRowReader_ColumnsRawReturnsStoredValues(t *testing.T) {
	f := newWorkbook(t)
	require.NoError(t, f.SetCellValue("Sheet1", "A1", 0.5))

	r, err := f.NewRowReader("Sheet1")
	require.NoError(t, err)
	defer r.Close()

	require.True(t, r.Next())
	cols, err := r.ColumnsRaw()

	require.NoError(t, err)
	assert.Equal(t, []string{"0.5"}, cols)
}

func TestRowReader_MissingSheetReturnsNotFound(t *testing.T) {
	f := newWorkbook(t)

	_, err := f.NewRowReader("NoSuchSheet")

	assert.True(t, stderr.IsErrNotFound(err))
}

func TestStreamWriter_WritesRows(t *testing.T) {
	f := newWorkbook(t)
	sw, err := f.NewStreamWriter("Sheet1")
	require.NoError(t, err)

	require.NoError(t, sw.SetRow("A1", []interface{}{"name", "qty"}))
	require.NoError(t, sw.SetRow("A2", []interface{}{"apple", 3}))
	require.NoError(t, sw.Flush())

	rows, err := f.GetRows("Sheet1")
	require.NoError(t, err)
	assert.Equal(t, [][]string{{"name", "qty"}, {"apple", "3"}}, rows)
}

func TestStreamWriter_MissingSheetReturnsError(t *testing.T) {
	f := newWorkbook(t)

	_, err := f.NewStreamWriter("NoSuchSheet")

	assert.Error(t, err)
}

func TestStreamWriter_InvalidCellReturnsError(t *testing.T) {
	f := newWorkbook(t)
	sw, err := f.NewStreamWriter("Sheet1")
	require.NoError(t, err)

	err = sw.SetRow("!!", []interface{}{"x"})

	assert.Error(t, err)
}
