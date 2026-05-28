package worksheet

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"
)

func newTestFile(t *testing.T) *excelize.File {
	t.Helper()
	f := excelize.NewFile()
	t.Cleanup(func() { f.Close() })
	return f
}

func TestListSheets(t *testing.T) {
	f := newTestFile(t)

	sheets := ListSheets(f)
	require.Len(t, sheets, 1)
	require.Equal(t, "Sheet1", sheets[0])
}

func TestCreateSheet(t *testing.T) {
	f := newTestFile(t)

	idx, err := CreateSheet(f, "TestSheet")
	require.NoError(t, err)
	require.NotZero(t, idx)

	sheets := ListSheets(f)
	require.Len(t, sheets, 2)
}

func TestDeleteSheet(t *testing.T) {
	f := newTestFile(t)

	_, err := CreateSheet(f, "ToDelete")
	require.NoError(t, err)

	err = DeleteSheet(f, "ToDelete")
	require.NoError(t, err)

	sheets := ListSheets(f)
	require.Len(t, sheets, 1)
}

func TestRenameSheet(t *testing.T) {
	f := newTestFile(t)

	err := RenameSheet(f, "Sheet1", "NewName")
	require.NoError(t, err)

	sheets := ListSheets(f)
	require.Equal(t, "NewName", sheets[0])
}

func TestCopySheet(t *testing.T) {
	f := newTestFile(t)

	newName, err := CopySheet(f, "Sheet1")
	require.NoError(t, err)
	require.NotEmpty(t, newName)

	sheets := ListSheets(f)
	require.Len(t, sheets, 2)
}

func TestSetActiveSheet(t *testing.T) {
	f := newTestFile(t)

	_, err := CreateSheet(f, "Active")
	require.NoError(t, err)

	err = SetActiveSheet(f, "Active")
	require.NoError(t, err)

	idx := f.GetActiveSheetIndex()
	sheetName := f.GetSheetName(idx)
	require.Equal(t, "Active", sheetName)

	// Non-existent sheet
	err = SetActiveSheet(f, "NonExistent")
	require.Error(t, err)
}

func TestGetSheetInfo(t *testing.T) {
	f := newTestFile(t)

	// Add some data
	f.SetCellValue("Sheet1", "A1", "Hello")
	f.SetCellValue("Sheet1", "B1", "World")
	f.SetCellValue("Sheet1", "A2", 123)

	info, err := GetSheetInfo(f, "Sheet1")
	require.NoError(t, err)
	require.Equal(t, "Sheet1", info.Name)
	require.GreaterOrEqual(t, info.RowCount, 2)
	require.GreaterOrEqual(t, info.ColCount, 2)

	// Non-existent sheet
	_, err = GetSheetInfo(f, "NonExistent")
	require.Error(t, err)
}
