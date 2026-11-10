package excel_test

import (
	"testing"

	"github.com/spotlibs/go-lib/stderr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetCellValue_ThenGetCellValue(t *testing.T) {
	f := newWorkbook(t)

	require.NoError(t, f.SetCellValue("Sheet1", "A1", "hello"))

	value, err := f.GetCellValue("Sheet1", "A1")
	require.NoError(t, err)
	assert.Equal(t, "hello", value)
}

func TestSetCellValue_MissingSheetReturnsError(t *testing.T) {
	f := newWorkbook(t)

	err := f.SetCellValue("NoSuchSheet", "A1", "x")

	assert.Error(t, err)
}

func TestGetCellValue_MissingSheetReturnsNotFound(t *testing.T) {
	f := newWorkbook(t)

	_, err := f.GetCellValue("NoSuchSheet", "A1")

	assert.True(t, stderr.IsErrNotFound(err))
}

func TestGetCellValue_EmptyCellReturnsEmptyString(t *testing.T) {
	f := newWorkbook(t)

	value, err := f.GetCellValue("Sheet1", "Z99")

	require.NoError(t, err)
	assert.Equal(t, "", value)
}

func TestSetCellFormula_Success(t *testing.T) {
	f := newWorkbook(t)
	require.NoError(t, f.SetCellValue("Sheet1", "A1", 1))
	require.NoError(t, f.SetCellValue("Sheet1", "A2", 2))

	err := f.SetCellFormula("Sheet1", "A3", "=SUM(A1:A2)")

	assert.NoError(t, err)
}

func TestSetCellFormula_MissingSheetReturnsError(t *testing.T) {
	f := newWorkbook(t)

	err := f.SetCellFormula("NoSuchSheet", "A1", "=1+1")

	assert.Error(t, err)
}

func TestGetRows_ReturnsRowsInOrder(t *testing.T) {
	f := newWorkbook(t)
	require.NoError(t, f.SetCellValue("Sheet1", "A1", "x"))
	require.NoError(t, f.SetCellValue("Sheet1", "B1", "y"))
	require.NoError(t, f.SetCellValue("Sheet1", "A2", "z"))

	rows, err := f.GetRows("Sheet1")

	require.NoError(t, err)
	assert.Equal(t, [][]string{{"x", "y"}, {"z"}}, rows)
}

func TestGetRows_MissingSheetReturnsNotFound(t *testing.T) {
	f := newWorkbook(t)

	_, err := f.GetRows("NoSuchSheet")

	assert.True(t, stderr.IsErrNotFound(err))
}

func TestGetCellValueRaw_ReturnsStoredNumber(t *testing.T) {
	f := newWorkbook(t)
	require.NoError(t, f.SetCellValue("Sheet1", "A1", 1234.5))

	value, err := f.GetCellValueRaw("Sheet1", "A1")

	require.NoError(t, err)
	assert.Equal(t, "1234.5", value)
}

func TestGetCellValueRaw_MissingSheetReturnsNotFound(t *testing.T) {
	f := newWorkbook(t)

	_, err := f.GetCellValueRaw("NoSuchSheet", "A1")

	assert.True(t, stderr.IsErrNotFound(err))
}

func TestGetRowsRaw_ReturnsStoredValues(t *testing.T) {
	f := newWorkbook(t)
	require.NoError(t, f.SetCellValue("Sheet1", "A1", 0.5))
	require.NoError(t, f.SetCellValue("Sheet1", "B1", "text"))

	rows, err := f.GetRowsRaw("Sheet1")

	require.NoError(t, err)
	assert.Equal(t, [][]string{{"0.5", "text"}}, rows)
}

func TestGetRowsRaw_MissingSheetReturnsNotFound(t *testing.T) {
	f := newWorkbook(t)

	_, err := f.GetRowsRaw("NoSuchSheet")

	assert.True(t, stderr.IsErrNotFound(err))
}

func TestGetCellInt_WholeNumber(t *testing.T) {
	f := newWorkbook(t)
	require.NoError(t, f.SetCellValue("Sheet1", "A1", 42))

	value, err := f.GetCellInt("Sheet1", "A1")

	require.NoError(t, err)
	assert.Equal(t, 42, value)
}

func TestGetCellInt_EmptyCellReturnsZero(t *testing.T) {
	f := newWorkbook(t)

	value, err := f.GetCellInt("Sheet1", "A1")

	require.NoError(t, err)
	assert.Equal(t, 0, value)
}

func TestGetCellInt_FractionReturnsError(t *testing.T) {
	f := newWorkbook(t)
	require.NoError(t, f.SetCellValue("Sheet1", "A1", 1.5))

	_, err := f.GetCellInt("Sheet1", "A1")

	assert.Error(t, err)
}

func TestGetCellInt_TextReturnsError(t *testing.T) {
	f := newWorkbook(t)
	require.NoError(t, f.SetCellValue("Sheet1", "A1", "abc"))

	_, err := f.GetCellInt("Sheet1", "A1")

	assert.Error(t, err)
}

func TestGetCellInt_NaNTextReturnsError(t *testing.T) {
	f := newWorkbook(t)
	require.NoError(t, f.SetCellValue("Sheet1", "A1", "NaN"))

	_, err := f.GetCellInt("Sheet1", "A1")

	assert.Error(t, err)
}

func TestGetCellInt_MissingSheetReturnsError(t *testing.T) {
	f := newWorkbook(t)

	_, err := f.GetCellInt("NoSuchSheet", "A1")

	assert.Error(t, err)
}

func TestGetCellFloat_Decimal(t *testing.T) {
	f := newWorkbook(t)
	require.NoError(t, f.SetCellValue("Sheet1", "A1", 3.25))

	value, err := f.GetCellFloat("Sheet1", "A1")

	require.NoError(t, err)
	assert.Equal(t, 3.25, value)
}

func TestGetCellFloat_EmptyCellReturnsZero(t *testing.T) {
	f := newWorkbook(t)

	value, err := f.GetCellFloat("Sheet1", "A1")

	require.NoError(t, err)
	assert.Equal(t, 0.0, value)
}

func TestGetCellFloat_InfinityTextReturnsError(t *testing.T) {
	f := newWorkbook(t)
	require.NoError(t, f.SetCellValue("Sheet1", "A1", "Inf"))

	_, err := f.GetCellFloat("Sheet1", "A1")

	assert.Error(t, err)
}

func TestGetCellFloat_MissingSheetReturnsError(t *testing.T) {
	f := newWorkbook(t)

	_, err := f.GetCellFloat("NoSuchSheet", "A1")

	assert.Error(t, err)
}

func TestGetCellTime_DateSerial(t *testing.T) {
	f := newWorkbook(t)
	require.NoError(t, f.SetCellValue("Sheet1", "A1", 45306.0))

	value, err := f.GetCellTime("Sheet1", "A1")

	require.NoError(t, err)
	assert.Equal(t, 2024, value.Year())
	assert.Equal(t, 1, int(value.Month()))
	assert.Equal(t, 15, value.Day())
}

func TestGetCellTime_EmptyCellReturnsZeroTime(t *testing.T) {
	f := newWorkbook(t)

	value, err := f.GetCellTime("Sheet1", "A1")

	require.NoError(t, err)
	assert.True(t, value.IsZero())
}

func TestGetCellTime_NegativeSerialReturnsError(t *testing.T) {
	f := newWorkbook(t)
	require.NoError(t, f.SetCellValue("Sheet1", "A1", -1.0))

	_, err := f.GetCellTime("Sheet1", "A1")

	assert.Error(t, err)
}

func TestGetCellTime_TextReturnsError(t *testing.T) {
	f := newWorkbook(t)
	require.NoError(t, f.SetCellValue("Sheet1", "A1", "not a date"))

	_, err := f.GetCellTime("Sheet1", "A1")

	assert.Error(t, err)
}

func TestGetCellTime_MissingSheetReturnsError(t *testing.T) {
	f := newWorkbook(t)

	_, err := f.GetCellTime("NoSuchSheet", "A1")

	assert.Error(t, err)
}
