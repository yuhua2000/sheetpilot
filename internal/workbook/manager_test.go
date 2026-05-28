package workbook

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestManager_Open(t *testing.T) {
	m := NewManager()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.xlsx")

	// Open new file (should create)
	wb, err := m.Open(path)
	require.NoError(t, err)
	require.NotEmpty(t, wb.ID)
	require.Equal(t, path, wb.Path)

	// Verify file exists
	_, err = os.Stat(path)
	require.NoError(t, err)

	// Open existing file
	wb2, err := m.Open(path)
	require.NoError(t, err)
	require.NotEqual(t, wb.ID, wb2.ID)
}

func TestManager_Save(t *testing.T) {
	m := NewManager()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "save_test.xlsx")

	wb, err := m.Open(path)
	require.NoError(t, err)

	// Modify file
	wb.File.SetCellValue("Sheet1", "A1", "test")

	// Save
	err = m.Save(wb.ID)
	require.NoError(t, err)

	// Save non-existent
	err = m.Save("invalid")
	require.Error(t, err)
}

func TestManager_SaveAs(t *testing.T) {
	m := NewManager()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "original.xlsx")
	newPath := filepath.Join(tmpDir, "copy.xlsx")

	wb, err := m.Open(path)
	require.NoError(t, err)

	// SaveAs
	err = m.SaveAs(wb.ID, newPath)
	require.NoError(t, err)

	// Verify new file exists
	_, err = os.Stat(newPath)
	require.NoError(t, err)

	// SaveAs non-existent
	err = m.SaveAs("invalid", newPath)
	require.Error(t, err)
}

func TestManager_Close(t *testing.T) {
	m := NewManager()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "close_test.xlsx")

	wb, err := m.Open(path)
	require.NoError(t, err)

	// Close
	err = m.Close(wb.ID)
	require.NoError(t, err)

	// Verify removed from list
	_, err = m.Get(wb.ID)
	require.Error(t, err)

	// Close non-existent
	err = m.Close("invalid")
	require.Error(t, err)
}

func TestManager_Get(t *testing.T) {
	m := NewManager()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "get_test.xlsx")

	wb, err := m.Open(path)
	require.NoError(t, err)

	// Get existing
	got, err := m.Get(wb.ID)
	require.NoError(t, err)
	require.Equal(t, wb.ID, got.ID)

	// Get non-existent
	_, err = m.Get("invalid")
	require.Error(t, err)
}

func TestManager_List(t *testing.T) {
	m := NewManager()
	tmpDir := t.TempDir()

	// Empty list
	list := m.List()
	require.Empty(t, list)

	// Open two files
	m.Open(filepath.Join(tmpDir, "1.xlsx"))
	m.Open(filepath.Join(tmpDir, "2.xlsx"))

	list = m.List()
	require.Len(t, list, 2)
}

func TestManager_GetInfo(t *testing.T) {
	m := NewManager()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "info_test.xlsx")

	wb, err := m.Open(path)
	require.NoError(t, err)

	// Get info
	info, err := m.GetInfo(wb.ID)
	require.NoError(t, err)
	require.Equal(t, wb.ID, info.ID)
	require.Equal(t, path, info.Path)
	require.GreaterOrEqual(t, info.SheetCount, 1)

	// Get info non-existent
	_, err = m.GetInfo("invalid")
	require.Error(t, err)
}
