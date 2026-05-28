package analysis

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

func TestGetTableInfo(t *testing.T) {
	f := newTestFile(t)

	// Set up test data
	f.SetCellValue("Sheet1", "A1", "Name")
	f.SetCellValue("Sheet1", "B1", "Age")
	f.SetCellValue("Sheet1", "C1", "City")
	f.SetCellValue("Sheet1", "A2", "Alice")
	f.SetCellValue("Sheet1", "B2", 25)
	f.SetCellValue("Sheet1", "C2", "Beijing")
	f.SetCellValue("Sheet1", "A3", "Bob")
	f.SetCellValue("Sheet1", "B3", 30)
	f.SetCellValue("Sheet1", "C3", "Shanghai")

	info, err := GetTableInfo(f, "Sheet1")
	require.NoError(t, err)
	require.Equal(t, "A1", info.StartCell)
	require.Equal(t, 3, info.RowCount)
	require.Equal(t, 3, info.ColCount)
	require.Len(t, info.Columns, 3)
	require.Equal(t, "Name", info.Columns[0])
	require.Equal(t, "Age", info.Columns[1])
	require.Equal(t, "City", info.Columns[2])
}

func TestGetTableInfo_EmptySheet(t *testing.T) {
	f := newTestFile(t)

	info, err := GetTableInfo(f, "Sheet1")
	require.NoError(t, err)
	require.Equal(t, 0, info.RowCount)
	require.Equal(t, 0, info.ColCount)
}

func TestGetColumnTypes(t *testing.T) {
	f := newTestFile(t)

	// Set up test data with different types
	f.SetCellValue("Sheet1", "A1", "Name")
	f.SetCellValue("Sheet1", "B1", "Age")
	f.SetCellValue("Sheet1", "C1", "Salary")
	f.SetCellValue("Sheet1", "D1", "Active")
	f.SetCellValue("Sheet1", "A2", "Alice")
	f.SetCellValue("Sheet1", "B2", 25)
	f.SetCellValue("Sheet1", "C2", "$5000")
	f.SetCellValue("Sheet1", "D2", "true")
	f.SetCellValue("Sheet1", "A3", "Bob")
	f.SetCellValue("Sheet1", "B3", 30)
	f.SetCellValue("Sheet1", "C3", "$6000")
	f.SetCellValue("Sheet1", "D3", "false")

	types, err := GetColumnTypes(f, "Sheet1", 100)
	require.NoError(t, err)
	require.Len(t, types, 4)
	require.Equal(t, "string", types[0].Type)
	require.Equal(t, "number", types[1].Type)
	require.Equal(t, "currency", types[2].Type)
	require.Equal(t, "bool", types[3].Type)
}

func TestGetSheetOverview(t *testing.T) {
	f := newTestFile(t)

	// Set up test data
	f.SetCellValue("Sheet1", "A1", "Product")
	f.SetCellValue("Sheet1", "B1", "Price")
	f.SetCellValue("Sheet1", "A2", "Apple")
	f.SetCellValue("Sheet1", "B2", 1.5)
	f.SetCellValue("Sheet1", "A3", "Banana")
	f.SetCellValue("Sheet1", "B3", 0.8)

	overview, err := GetSheetOverview(f, "Sheet1")
	require.NoError(t, err)
	require.Equal(t, "Sheet1", overview.SheetName)
	require.Equal(t, 3, overview.RowCount)
	require.Equal(t, 2, overview.ColCount)
	require.Len(t, overview.Columns, 2)
	require.Len(t, overview.Samples, 2) // 2 data rows
}
