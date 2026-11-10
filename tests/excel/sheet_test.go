package excel_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSheet_ReturnsIndex(t *testing.T) {
	f := newWorkbook(t)

	idx, err := f.NewSheet("Data")

	require.NoError(t, err)
	assert.Equal(t, 1, idx)
	assert.Equal(t, []string{"Sheet1", "Data"}, f.GetSheetList())
}

func TestNewSheet_InvalidNameReturnsError(t *testing.T) {
	f := newWorkbook(t)

	_, err := f.NewSheet("this sheet name is far too long to be valid in excel")

	assert.Error(t, err)
}

func TestDeleteSheet_RemovesSheet(t *testing.T) {
	f := newWorkbook(t)
	_, err := f.NewSheet("Data")
	require.NoError(t, err)

	err = f.DeleteSheet("Sheet1")

	require.NoError(t, err)
	assert.Equal(t, []string{"Data"}, f.GetSheetList())
}

func TestDeleteSheet_MissingSheetIsNoOp(t *testing.T) {
	f := newWorkbook(t)

	err := f.DeleteSheet("NoSuchSheet")

	assert.NoError(t, err)
	assert.Equal(t, []string{"Sheet1"}, f.GetSheetList())
}

func TestCopySheet_OverwritesDestinationWithSource(t *testing.T) {
	f := newWorkbook(t)
	require.NoError(t, f.SetCellValue("Sheet1", "A1", "copy-me"))
	_, err := f.NewSheet("Data")
	require.NoError(t, err)

	err = f.CopySheet(0, 1)

	require.NoError(t, err)
	assert.Len(t, f.GetSheetList(), 2)
	value, err := f.GetCellValue("Data", "A1")
	require.NoError(t, err)
	assert.Equal(t, "copy-me", value)
}

func TestCopySheet_SameIndexReturnsError(t *testing.T) {
	f := newWorkbook(t)

	err := f.CopySheet(0, 0)

	assert.Error(t, err)
}

func TestCopySheet_OutOfRangeIndexReturnsError(t *testing.T) {
	f := newWorkbook(t)

	err := f.CopySheet(99, 0)

	assert.Error(t, err)
}

func TestSetActiveSheet_KeepsSheetList(t *testing.T) {
	f := newWorkbook(t)
	_, err := f.NewSheet("Data")
	require.NoError(t, err)

	f.SetActiveSheet(1)

	assert.Equal(t, []string{"Sheet1", "Data"}, f.GetSheetList())
}

func TestGetSheetIndex_Found(t *testing.T) {
	f := newWorkbook(t)

	idx, err := f.GetSheetIndex("Sheet1")

	require.NoError(t, err)
	assert.Equal(t, 0, idx)
}

func TestGetSheetIndex_MissingSheetReturnsMinusOne(t *testing.T) {
	f := newWorkbook(t)

	idx, _ := f.GetSheetIndex("NoSuchSheet")

	assert.Equal(t, -1, idx)
}

func TestGetSheetIndex_InvalidNameReturnsError(t *testing.T) {
	f := newWorkbook(t)

	idx, err := f.GetSheetIndex("this sheet name is far too long to be valid in excel")

	assert.Error(t, err)
	assert.Equal(t, -1, idx)
}

func TestGetSheetList_ReturnsNamesInOrder(t *testing.T) {
	f := newWorkbook(t)
	_, err := f.NewSheet("Second")
	require.NoError(t, err)

	assert.Equal(t, []string{"Sheet1", "Second"}, f.GetSheetList())
}

func TestHasSheet_ExistingAndMissing(t *testing.T) {
	f := newWorkbook(t)

	assert.True(t, f.HasSheet("Sheet1"))
	assert.False(t, f.HasSheet("NoSuchSheet"))
}

func TestGetFirstSheet_ReturnsFirstName(t *testing.T) {
	f := newWorkbook(t)
	_, err := f.NewSheet("Second")
	require.NoError(t, err)

	name, err := f.GetFirstSheet()

	require.NoError(t, err)
	assert.Equal(t, "Sheet1", name)
}
