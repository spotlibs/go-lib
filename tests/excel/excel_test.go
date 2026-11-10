package excel_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/spotlibs/go-lib/excel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew_HasDefaultSheet(t *testing.T) {
	f := newWorkbook(t)

	assert.Equal(t, []string{"Sheet1"}, f.GetSheetList())
}

func TestOpen_Success(t *testing.T) {
	path := newSavedWorkbook(t, map[string]interface{}{"A1": "hello"})

	f, err := excel.Open(path)
	require.NoError(t, err)
	defer f.Close()

	value, err := f.GetCellValue("Sheet1", "A1")
	require.NoError(t, err)
	assert.Equal(t, "hello", value)
}

func TestOpen_MissingFileReturnsError(t *testing.T) {
	f, err := excel.Open(filepath.Join(t.TempDir(), "missing.xlsx"))

	assert.Error(t, err)
	assert.Nil(t, f)
}

func TestOpenReader_Success(t *testing.T) {
	src := newWorkbook(t)
	require.NoError(t, src.SetCellValue("Sheet1", "B2", "reader"))
	buf, err := src.WriteToBuffer()
	require.NoError(t, err)

	f, err := excel.OpenReader(bytes.NewReader(buf.Bytes()))
	require.NoError(t, err)
	defer f.Close()

	value, err := f.GetCellValue("Sheet1", "B2")
	require.NoError(t, err)
	assert.Equal(t, "reader", value)
}

func TestOpenReader_InvalidDataReturnsError(t *testing.T) {
	f, err := excel.OpenReader(bytes.NewReader([]byte("not an xlsx file")))

	assert.Error(t, err)
	assert.Nil(t, f)
}

func TestOpenWithOptions_Success(t *testing.T) {
	path := newSavedWorkbook(t, map[string]interface{}{"A1": "with-options"})

	f, err := excel.OpenWithOptions(path, excel.OpenOptions{})
	require.NoError(t, err)
	defer f.Close()

	value, err := f.GetCellValue("Sheet1", "A1")
	require.NoError(t, err)
	assert.Equal(t, "with-options", value)
}

func TestOpenWithOptions_MissingFileReturnsError(t *testing.T) {
	f, err := excel.OpenWithOptions(filepath.Join(t.TempDir(), "missing.xlsx"), excel.OpenOptions{})

	assert.Error(t, err)
	assert.Nil(t, f)
}

func TestOpenWithOptions_SizeLimitExceededReturnsError(t *testing.T) {
	path := newSavedWorkbook(t, map[string]interface{}{"A1": "big enough to exceed the limit"})

	f, err := excel.OpenWithOptions(path, excel.OpenOptions{
		UnzipSizeLimit:    1024,
		UnzipXMLSizeLimit: 1024,
	})

	assert.Error(t, err)
	assert.Nil(t, f)
}

func TestOpenReaderWithOptions_Success(t *testing.T) {
	src := newWorkbook(t)
	require.NoError(t, src.SetCellValue("Sheet1", "C3", "opt-reader"))
	buf, err := src.WriteToBuffer()
	require.NoError(t, err)

	f, err := excel.OpenReaderWithOptions(bytes.NewReader(buf.Bytes()), excel.OpenOptions{})
	require.NoError(t, err)
	defer f.Close()

	value, err := f.GetCellValue("Sheet1", "C3")
	require.NoError(t, err)
	assert.Equal(t, "opt-reader", value)
}

func TestOpenReaderWithOptions_InvalidDataReturnsError(t *testing.T) {
	f, err := excel.OpenReaderWithOptions(bytes.NewReader([]byte("garbage")), excel.OpenOptions{})

	assert.Error(t, err)
	assert.Nil(t, f)
}

func TestWriteToBuffer_ProducesZipArchive(t *testing.T) {
	f := newWorkbook(t)
	require.NoError(t, f.SetCellValue("Sheet1", "A1", "buffer"))

	buf, err := f.WriteToBuffer()

	require.NoError(t, err)
	// XLSX files are zip archives, which start with the "PK" signature.
	assert.True(t, bytes.HasPrefix(buf.Bytes(), []byte("PK")))
}

func TestSaveAs_WritesFileToDisk(t *testing.T) {
	f := newWorkbook(t)
	require.NoError(t, f.SetCellValue("Sheet1", "A1", "saved"))
	path := filepath.Join(t.TempDir(), "saved.xlsx")

	err := f.SaveAs(path)

	require.NoError(t, err)
	_, statErr := os.Stat(path)
	assert.NoError(t, statErr)
}

func TestSaveAs_InvalidPathReturnsError(t *testing.T) {
	f := newWorkbook(t)
	path := filepath.Join(t.TempDir(), "missing-dir", "saved.xlsx")

	err := f.SaveAs(path)

	assert.Error(t, err)
}

func TestClose_NewWorkbookReturnsNil(t *testing.T) {
	f := excel.New()

	assert.NoError(t, f.Close())
}
