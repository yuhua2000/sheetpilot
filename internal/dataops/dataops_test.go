package dataops

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

func TestAddComputedColumn(t *testing.T) {
	f := newTestFile(t)

	// Set up test data
	f.SetCellValue("Sheet1", "A1", "Revenue")
	f.SetCellValue("Sheet1", "B1", "Cost")
	f.SetCellValue("Sheet1", "A2", 1000)
	f.SetCellValue("Sheet1", "B2", 600)
	f.SetCellValue("Sheet1", "A3", 2000)
	f.SetCellValue("Sheet1", "B3", 1200)

	err := AddComputedColumn(f, "Sheet1", "Profit", "{Revenue}-{Cost}")
	require.NoError(t, err)

	// Verify header was added
	val, _ := f.GetCellValue("Sheet1", "C1")
	require.Equal(t, "Profit", val)

	// Verify formula was added
	formula, _ := f.GetCellFormula("Sheet1", "C2")
	require.NotEmpty(t, formula)
}

func TestFillMissingValues(t *testing.T) {
	f := newTestFile(t)

	f.SetCellValue("Sheet1", "A1", "Name")
	f.SetCellValue("Sheet1", "A2", "Alice")
	f.SetCellValue("Sheet1", "A3", "")
	f.SetCellValue("Sheet1", "A4", "Bob")

	err := FillMissingValues(f, "Sheet1", "A", "Unknown")
	require.NoError(t, err)

	val, _ := f.GetCellValue("Sheet1", "A3")
	require.Equal(t, "Unknown", val)
}

func TestReplaceValues(t *testing.T) {
	f := newTestFile(t)

	f.SetCellValue("Sheet1", "A1", "Country")
	f.SetCellValue("Sheet1", "A2", "USA")
	f.SetCellValue("Sheet1", "A3", "UK")
	f.SetCellValue("Sheet1", "A4", "USA")

	err := ReplaceValues(f, "Sheet1", "A", "USA", "United States")
	require.NoError(t, err)

	val, _ := f.GetCellValue("Sheet1", "A2")
	require.Equal(t, "United States", val)

	val, _ = f.GetCellValue("Sheet1", "A4")
	require.Equal(t, "United States", val)
}

func TestCleanupSheet(t *testing.T) {
	f := newTestFile(t)

	f.SetCellValue("Sheet1", "A1", " Name ")
	f.SetCellValue("Sheet1", "B1", " Age")
	f.SetCellValue("Sheet1", "A2", "Alice")
	f.SetCellValue("Sheet1", "B2", 25)

	err := CleanupSheet(f, "Sheet1")
	require.NoError(t, err)

	val, _ := f.GetCellValue("Sheet1", "A1")
	require.Equal(t, "Name", val)
}

func TestDeduplicate(t *testing.T) {
	f := newTestFile(t)

	f.SetCellValue("Sheet1", "A1", "Name")
	f.SetCellValue("Sheet1", "B1", "City")
	f.SetCellValue("Sheet1", "A2", "Alice")
	f.SetCellValue("Sheet1", "B2", "Beijing")
	f.SetCellValue("Sheet1", "A3", "Bob")
	f.SetCellValue("Sheet1", "B3", "Shanghai")
	f.SetCellValue("Sheet1", "A4", "Alice")
	f.SetCellValue("Sheet1", "B4", "Beijing")

	err := Deduplicate(f, "Sheet1", []string{"A", "B"})
	require.NoError(t, err)

	// Verify duplicate removed
	val, _ := f.GetCellValue("Sheet1", "A3")
	require.NotEqual(t, "Alice", val)
}

func TestSortTable(t *testing.T) {
	f := newTestFile(t)

	f.SetCellValue("Sheet1", "A1", "Name")
	f.SetCellValue("Sheet1", "B1", "Score")
	f.SetCellValue("Sheet1", "A2", "Charlie")
	f.SetCellValue("Sheet1", "B2", 70)
	f.SetCellValue("Sheet1", "A3", "Alice")
	f.SetCellValue("Sheet1", "B3", 90)
	f.SetCellValue("Sheet1", "A4", "Bob")
	f.SetCellValue("Sheet1", "B4", 80)

	err := SortTable(f, "Sheet1", []string{"B"}, true)
	require.NoError(t, err)

	// Verify sorted ascending by score
	val, _ := f.GetCellValue("Sheet1", "A2")
	require.Equal(t, "Charlie", val) // 70
}

func TestFilterRows(t *testing.T) {
	f := newTestFile(t)

	f.SetCellValue("Sheet1", "A1", "Name")
	f.SetCellValue("Sheet1", "B1", "Age")
	f.SetCellValue("Sheet1", "A2", "Alice")
	f.SetCellValue("Sheet1", "B2", 25)
	f.SetCellValue("Sheet1", "A3", "Bob")
	f.SetCellValue("Sheet1", "B3", 30)
	f.SetCellValue("Sheet1", "A4", "Charlie")
	f.SetCellValue("Sheet1", "B4", 20)

	result, err := FilterRows(f, "Sheet1", "B", ">", "22")
	require.NoError(t, err)
	require.Len(t, result, 3) // header + 2 rows
}

func TestGroupBy(t *testing.T) {
	f := newTestFile(t)

	f.SetCellValue("Sheet1", "A1", "Region")
	f.SetCellValue("Sheet1", "B1", "Sales")
	f.SetCellValue("Sheet1", "A2", "North")
	f.SetCellValue("Sheet1", "B2", 100)
	f.SetCellValue("Sheet1", "A3", "South")
	f.SetCellValue("Sheet1", "B3", 200)
	f.SetCellValue("Sheet1", "A4", "North")
	f.SetCellValue("Sheet1", "B4", 150)

	resultSheet, err := GroupBy(f, "Sheet1", "A", "B", "sum")
	require.NoError(t, err)
	require.NotEmpty(t, resultSheet)

	// Verify result sheet exists
	sheets := f.GetSheetList()
	require.Contains(t, sheets, resultSheet)

	// Verify headers
	headerA, _ := f.GetCellValue(resultSheet, "A1")
	require.Equal(t, "Region", headerA)

	headerB, _ := f.GetCellValue(resultSheet, "B1")
	require.Equal(t, "sum(Sales)", headerB)

	// Verify group values exist
	valA2, _ := f.GetCellValue(resultSheet, "A2")
	require.NotEmpty(t, valA2)

	// Verify formula exists
	formulaB2, _ := f.GetCellFormula(resultSheet, "B2")
	require.NotEmpty(t, formulaB2)
}
