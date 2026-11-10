package excel_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spotlibs/go-lib/excel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenFromNFS_Success(t *testing.T) {
	path := newSavedWorkbook(t, map[string]interface{}{"A1": "from-nfs"})

	f, err := excel.OpenFromNFS(path)
	require.NoError(t, err)
	defer f.Close()

	value, err := f.GetCellValue("Sheet1", "A1")
	require.NoError(t, err)
	assert.Equal(t, "from-nfs", value)
}

func TestOpenFromNFS_MissingFileReturnsError(t *testing.T) {
	f, err := excel.OpenFromNFS(filepath.Join(t.TempDir(), "missing.xlsx"))

	assert.Error(t, err)
	assert.Nil(t, f)
}

func TestOpenFromNFS_InvalidContentReturnsError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "broken.xlsx")
	require.NoError(t, os.WriteFile(path, []byte("not a workbook"), 0600))

	f, err := excel.OpenFromNFS(path)

	assert.Error(t, err)
	assert.Nil(t, f)
}

func TestOpenFromNFSWithOptions_Success(t *testing.T) {
	path := newSavedWorkbook(t, map[string]interface{}{"B1": "nfs-opts"})

	f, err := excel.OpenFromNFSWithOptions(path, excel.OpenOptions{})
	require.NoError(t, err)
	defer f.Close()

	value, err := f.GetCellValue("Sheet1", "B1")
	require.NoError(t, err)
	assert.Equal(t, "nfs-opts", value)
}

func TestOpenFromNFSWithOptions_MissingFileReturnsError(t *testing.T) {
	f, err := excel.OpenFromNFSWithOptions(filepath.Join(t.TempDir(), "missing.xlsx"), excel.OpenOptions{})

	assert.Error(t, err)
	assert.Nil(t, f)
}

func TestOpenFromNFSWithOptions_InvalidContentReturnsError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "broken.xlsx")
	require.NoError(t, os.WriteFile(path, []byte("garbage"), 0600))

	f, err := excel.OpenFromNFSWithOptions(path, excel.OpenOptions{})

	assert.Error(t, err)
	assert.Nil(t, f)
}

func TestSaveAsNFS_WritesWorkbook(t *testing.T) {
	f := newWorkbook(t)
	require.NoError(t, f.SetCellValue("Sheet1", "A1", "saved-nfs"))
	dest := filepath.Join(t.TempDir(), "nested", "report.xlsx")

	result, err := f.SaveAsNFS(dest)

	require.NoError(t, err)
	assert.Equal(t, "NFS", result.Driver)
	assert.Equal(t, dest, result.FullPath)
	assert.Equal(t, "report.xlsx", result.PathFile)

	reopened, err := excel.OpenFromNFS(dest)
	require.NoError(t, err)
	defer reopened.Close()
	value, err := reopened.GetCellValue("Sheet1", "A1")
	require.NoError(t, err)
	assert.Equal(t, "saved-nfs", value)
}

func TestSaveAsNFS_UnwritableDestinationReturnsError(t *testing.T) {
	f := newWorkbook(t)
	blocker := filepath.Join(t.TempDir(), "blocker")
	require.NoError(t, os.WriteFile(blocker, []byte("file"), 0600))

	result, err := f.SaveAsNFS(filepath.Join(blocker, "report.xlsx"))

	assert.Error(t, err)
	assert.Nil(t, result)
}
