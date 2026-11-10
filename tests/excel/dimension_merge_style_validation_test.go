package excel_test

import (
	"testing"

	"github.com/spotlibs/go-lib/excel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"
)

func TestSetRowHeight_ThenGetRowHeight(t *testing.T) {
	f := newWorkbook(t)

	require.NoError(t, f.SetRowHeight("Sheet1", 2, 30))

	height, err := f.GetRowHeight("Sheet1", 2)
	require.NoError(t, err)
	assert.Equal(t, 30.0, height)
}

func TestSetRowHeight_MissingSheetReturnsError(t *testing.T) {
	f := newWorkbook(t)

	err := f.SetRowHeight("NoSuchSheet", 1, 20)

	assert.Error(t, err)
}

func TestSetRowHeight_InvalidRowReturnsError(t *testing.T) {
	f := newWorkbook(t)

	err := f.SetRowHeight("Sheet1", 0, 20)

	assert.Error(t, err)
}

func TestGetRowHeight_MissingSheetReturnsError(t *testing.T) {
	f := newWorkbook(t)

	_, err := f.GetRowHeight("NoSuchSheet", 1)

	assert.Error(t, err)
}

func TestSetColWidth_ThenGetColWidth(t *testing.T) {
	f := newWorkbook(t)

	require.NoError(t, f.SetColWidth("Sheet1", "A", "C", 20))

	width, err := f.GetColWidth("Sheet1", "B")
	require.NoError(t, err)
	assert.Equal(t, 20.0, width)
}

func TestSetColWidth_MissingSheetReturnsError(t *testing.T) {
	f := newWorkbook(t)

	err := f.SetColWidth("NoSuchSheet", "A", "B", 10)

	assert.Error(t, err)
}

func TestSetColWidth_InvalidColumnReturnsError(t *testing.T) {
	f := newWorkbook(t)

	err := f.SetColWidth("Sheet1", "1", "B", 10)

	assert.Error(t, err)
}

func TestGetColWidth_MissingSheetReturnsError(t *testing.T) {
	f := newWorkbook(t)

	_, err := f.GetColWidth("NoSuchSheet", "A")

	assert.Error(t, err)
}

func TestMergeCell_ThenUnmergeCell(t *testing.T) {
	f := newWorkbook(t)

	require.NoError(t, f.MergeCell("Sheet1", "A1", "B2"))

	assert.NoError(t, f.UnmergeCell("Sheet1", "A1", "B2"))
}

func TestMergeCell_MissingSheetReturnsError(t *testing.T) {
	f := newWorkbook(t)

	assert.Error(t, f.MergeCell("NoSuchSheet", "A1", "B2"))
}

func TestUnmergeCell_MissingSheetReturnsError(t *testing.T) {
	f := newWorkbook(t)

	assert.Error(t, f.UnmergeCell("NoSuchSheet", "A1", "B2"))
}

func TestNewStyle_ReturnsStyleID(t *testing.T) {
	f := newWorkbook(t)

	id, err := f.NewStyle(&excel.Style{Font: &excelize.Font{Bold: true}})

	require.NoError(t, err)
	assert.Greater(t, id, 0)
}

func TestSetCellStyle_Success(t *testing.T) {
	f := newWorkbook(t)
	id, err := f.NewStyle(&excel.Style{Font: &excelize.Font{Bold: true}})
	require.NoError(t, err)

	err = f.SetCellStyle("Sheet1", "A1", "B2", id)

	assert.NoError(t, err)
}

func TestSetCellStyle_MissingSheetReturnsError(t *testing.T) {
	f := newWorkbook(t)
	id, err := f.NewStyle(&excel.Style{Font: &excelize.Font{Bold: true}})
	require.NoError(t, err)

	err = f.SetCellStyle("NoSuchSheet", "A1", "B2", id)

	assert.Error(t, err)
}

func TestAddDataValidation_ThenDelete(t *testing.T) {
	f := newWorkbook(t)
	dv := excelize.NewDataValidation(true)
	dv.Sqref = "A1:A10"
	require.NoError(t, dv.SetDropList([]string{"yes", "no"}))

	require.NoError(t, f.AddDataValidation("Sheet1", dv))

	assert.NoError(t, f.DeleteDataValidation("Sheet1", "A1:A10"))
}

func TestAddDataValidation_MissingSheetReturnsError(t *testing.T) {
	f := newWorkbook(t)
	dv := excelize.NewDataValidation(true)
	dv.Sqref = "A1"
	require.NoError(t, dv.SetDropList([]string{"yes"}))

	err := f.AddDataValidation("NoSuchSheet", dv)

	assert.Error(t, err)
}

func TestDeleteDataValidation_AllRulesOnSheet(t *testing.T) {
	f := newWorkbook(t)

	assert.NoError(t, f.DeleteDataValidation("Sheet1"))
}

func TestDeleteDataValidation_MissingSheetReturnsError(t *testing.T) {
	f := newWorkbook(t)

	assert.Error(t, f.DeleteDataValidation("NoSuchSheet"))
}
