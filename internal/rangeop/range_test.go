package rangeop

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

func TestGetCell(t *testing.T) {
	f := newTestFile(t)
	f.SetCellValue("Sheet1", "A1", "hello")

	val, err := GetCell(f, "Sheet1", "A1")
	require.NoError(t, err)
	require.Equal(t, "hello", val)
}

func TestSetCell(t *testing.T) {
	f := newTestFile(t)

	err := SetCell(f, "Sheet1", "A1", "world")
	require.NoError(t, err)

	val, _ := f.GetCellValue("Sheet1", "A1")
	require.Equal(t, "world", val)
}

func TestClearCell(t *testing.T) {
	f := newTestFile(t)
	f.SetCellValue("Sheet1", "A1", "to clear")

	err := ClearCell(f, "Sheet1", "A1")
	require.NoError(t, err)

	val, _ := f.GetCellValue("Sheet1", "A1")
	require.Empty(t, val)
}

func TestGetRange(t *testing.T) {
	f := newTestFile(t)
	f.SetCellValue("Sheet1", "A1", "a")
	f.SetCellValue("Sheet1", "B1", "b")
	f.SetCellValue("Sheet1", "A2", "c")
	f.SetCellValue("Sheet1", "B2", "d")

	data, err := GetRange(f, "Sheet1", "A1:B2")
	require.NoError(t, err)
	require.Len(t, data, 2)
	require.Len(t, data[0], 2)
	require.Equal(t, "a", data[0][0])
	require.Equal(t, "b", data[0][1])
}

func TestSetRange(t *testing.T) {
	f := newTestFile(t)

	data := [][]any{
		{"x", "y"},
		{1, 2},
	}
	err := SetRange(f, "Sheet1", "A1", data)
	require.NoError(t, err)

	val, _ := f.GetCellValue("Sheet1", "A1")
	require.Equal(t, "x", val)
}

func TestAppendRows(t *testing.T) {
	f := newTestFile(t)
	f.SetCellValue("Sheet1", "A1", "header")

	data := [][]any{
		{"row1"},
		{"row2"},
	}
	err := AppendRows(f, "Sheet1", data)
	require.NoError(t, err)

	val, _ := f.GetCellValue("Sheet1", "A2")
	require.Equal(t, "row1", val)
}

func TestInsertRows(t *testing.T) {
	f := newTestFile(t)
	f.SetCellValue("Sheet1", "A1", "first")
	f.SetCellValue("Sheet1", "A2", "second")

	err := InsertRows(f, "Sheet1", 2, 1)
	require.NoError(t, err)

	// "second" should now be at A3
	val, _ := f.GetCellValue("Sheet1", "A3")
	require.Equal(t, "second", val)
}

func TestDeleteRows(t *testing.T) {
	f := newTestFile(t)
	f.SetCellValue("Sheet1", "A1", "first")
	f.SetCellValue("Sheet1", "A2", "second")
	f.SetCellValue("Sheet1", "A3", "third")

	err := DeleteRows(f, "Sheet1", 2, 1)
	require.NoError(t, err)

	// "third" should now be at A2
	val, _ := f.GetCellValue("Sheet1", "A2")
	require.Equal(t, "third", val)
}

func TestInsertCols(t *testing.T) {
	f := newTestFile(t)
	f.SetCellValue("Sheet1", "A1", "col1")
	f.SetCellValue("Sheet1", "B1", "col2")

	err := InsertCols(f, "Sheet1", "B", 1)
	require.NoError(t, err)

	// "col2" should now be at C1
	val, _ := f.GetCellValue("Sheet1", "C1")
	require.Equal(t, "col2", val)
}

func TestDeleteCols(t *testing.T) {
	f := newTestFile(t)
	f.SetCellValue("Sheet1", "A1", "col1")
	f.SetCellValue("Sheet1", "B1", "col2")
	f.SetCellValue("Sheet1", "C1", "col3")

	err := DeleteCols(f, "Sheet1", "B", 1)
	require.NoError(t, err)

	// "col3" should now be at B1
	val, _ := f.GetCellValue("Sheet1", "B1")
	require.Equal(t, "col3", val)
}
