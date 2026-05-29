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

func TestSplitSheet(t *testing.T) {
	f := newTestFile(t)

	f.SetCellValue("Sheet1", "A1", "Region")
	f.SetCellValue("Sheet1", "B1", "Sales")
	f.SetCellValue("Sheet1", "A2", "North")
	f.SetCellValue("Sheet1", "B2", 100)
	f.SetCellValue("Sheet1", "A3", "South")
	f.SetCellValue("Sheet1", "B3", 200)
	f.SetCellValue("Sheet1", "A4", "North")
	f.SetCellValue("Sheet1", "B4", 150)

	newSheets, err := SplitSheet(f, "Sheet1", "A")
	require.NoError(t, err)
	require.Len(t, newSheets, 2)

	// Verify new sheets exist
	sheets := f.GetSheetList()
	require.Contains(t, sheets, "Sheet1_North")
	require.Contains(t, sheets, "Sheet1_South")

	// Verify data in North sheet
	val, _ := f.GetCellValue("Sheet1_North", "B2")
	require.Equal(t, "100", val)
}

func TestMergeSheets(t *testing.T) {
	f := newTestFile(t)

	// Create test sheets
	f.NewSheet("Sheet2")
	f.SetCellValue("Sheet1", "A1", "Name")
	f.SetCellValue("Sheet1", "B1", "Value")
	f.SetCellValue("Sheet1", "A2", "A")
	f.SetCellValue("Sheet1", "B2", 10)

	f.SetCellValue("Sheet2", "A1", "Name")
	f.SetCellValue("Sheet2", "B1", "Value")
	f.SetCellValue("Sheet2", "A2", "B")
	f.SetCellValue("Sheet2", "B2", 20)

	err := MergeSheets(f, []string{"Sheet1", "Sheet2"}, "Merged")
	require.NoError(t, err)

	// Verify merged sheet
	sheets := f.GetSheetList()
	require.Contains(t, sheets, "Merged")

	// Verify header
	header, _ := f.GetCellValue("Merged", "A1")
	require.Equal(t, "Name", header)

	// Verify data
	valA2, _ := f.GetCellValue("Merged", "A2")
	require.Equal(t, "A", valA2)

	valA3, _ := f.GetCellValue("Merged", "A3")
	require.Equal(t, "B", valA3)
}

func TestSortTableWithColumnNames(t *testing.T) {
	f := newTestFile(t)

	// Set up test data with Chinese column names (use string values for sorting)
	f.SetCellValue("Sheet1", "A1", "姓名")
	f.SetCellValue("Sheet1", "B1", "部门")
	f.SetCellValue("Sheet1", "C1", "等级")
	f.SetCellValue("Sheet1", "A2", "张三")
	f.SetCellValue("Sheet1", "B2", "销售部")
	f.SetCellValue("Sheet1", "C2", "B")
	f.SetCellValue("Sheet1", "A3", "李四")
	f.SetCellValue("Sheet1", "B3", "技术部")
	f.SetCellValue("Sheet1", "C3", "A")
	f.SetCellValue("Sheet1", "A4", "王五")
	f.SetCellValue("Sheet1", "B4", "销售部")
	f.SetCellValue("Sheet1", "C4", "C")

	// Sort by column name "等级" ascending
	err := SortTable(f, "Sheet1", []string{"等级"}, true)
	require.NoError(t, err)

	// Verify sorted order: A < B < C
	val, _ := f.GetCellValue("Sheet1", "A2")
	require.Equal(t, "李四", val) // A is smallest
}

func TestSortTableWithColumnLetters(t *testing.T) {
	f := newTestFile(t)

	// Set up test data
	f.SetCellValue("Sheet1", "A1", "Name")
	f.SetCellValue("Sheet1", "B1", "Score")
	f.SetCellValue("Sheet1", "A2", "Alice")
	f.SetCellValue("Sheet1", "B2", 90)
	f.SetCellValue("Sheet1", "A3", "Bob")
	f.SetCellValue("Sheet1", "B3", 70)
	f.SetCellValue("Sheet1", "A4", "Charlie")
	f.SetCellValue("Sheet1", "B4", 85)

	// Sort by column letter "B" ascending
	err := SortTable(f, "Sheet1", []string{"B"}, true)
	require.NoError(t, err)

	// Verify sorted order
	val, _ := f.GetCellValue("Sheet1", "A2")
	require.Equal(t, "Bob", val) // 70 is smallest
}

func TestFilterRowsWithColumnName(t *testing.T) {
	f := newTestFile(t)

	// Set up test data with Chinese column names (use string values for comparison)
	f.SetCellValue("Sheet1", "A1", "姓名")
	f.SetCellValue("Sheet1", "B1", "部门")
	f.SetCellValue("Sheet1", "C1", "等级")
	f.SetCellValue("Sheet1", "A2", "张三")
	f.SetCellValue("Sheet1", "B2", "销售部")
	f.SetCellValue("Sheet1", "C2", "B")
	f.SetCellValue("Sheet1", "A3", "李四")
	f.SetCellValue("Sheet1", "B3", "技术部")
	f.SetCellValue("Sheet1", "C3", "A")
	f.SetCellValue("Sheet1", "A4", "王五")
	f.SetCellValue("Sheet1", "B4", "销售部")
	f.SetCellValue("Sheet1", "C4", "C")

	// Filter by column name "等级" > "A"
	result, err := FilterRows(f, "Sheet1", "等级", ">", "A")
	require.NoError(t, err)
	require.Len(t, result, 3) // header + 2 rows (B, C)
}

func TestFilterRowsWithColumnLetter(t *testing.T) {
	f := newTestFile(t)

	// Set up test data
	f.SetCellValue("Sheet1", "A1", "Name")
	f.SetCellValue("Sheet1", "B1", "Age")
	f.SetCellValue("Sheet1", "A2", "Alice")
	f.SetCellValue("Sheet1", "B2", 25)
	f.SetCellValue("Sheet1", "A3", "Bob")
	f.SetCellValue("Sheet1", "B3", 30)
	f.SetCellValue("Sheet1", "A4", "Charlie")
	f.SetCellValue("Sheet1", "B4", 20)

	// Filter by column letter "B" > 22
	result, err := FilterRows(f, "Sheet1", "B", ">", "22")
	require.NoError(t, err)
	require.Len(t, result, 3) // header + 2 rows (25, 30)
}

func TestDeduplicateWithColumnNames(t *testing.T) {
	f := newTestFile(t)

	// Set up test data with Chinese column names
	f.SetCellValue("Sheet1", "A1", "姓名")
	f.SetCellValue("Sheet1", "B1", "部门")
	f.SetCellValue("Sheet1", "A2", "张三")
	f.SetCellValue("Sheet1", "B2", "销售部")
	f.SetCellValue("Sheet1", "A3", "李四")
	f.SetCellValue("Sheet1", "B3", "技术部")
	f.SetCellValue("Sheet1", "A4", "王五")
	f.SetCellValue("Sheet1", "B4", "销售部")

	// Deduplicate by column name "部门"
	err := Deduplicate(f, "Sheet1", []string{"部门"})
	require.NoError(t, err)

	// Verify only unique departments remain
	val, _ := f.GetCellValue("Sheet1", "A3")
	require.NotEqual(t, "王五", val) // 王五 should be removed (duplicate 部门)
}

func TestGroupByWithColumnNames(t *testing.T) {
	f := newTestFile(t)

	// Set up test data with Chinese column names
	f.SetCellValue("Sheet1", "A1", "部门")
	f.SetCellValue("Sheet1", "B1", "工资")
	f.SetCellValue("Sheet1", "A2", "销售部")
	f.SetCellValue("Sheet1", "B2", 8000)
	f.SetCellValue("Sheet1", "A3", "技术部")
	f.SetCellValue("Sheet1", "B3", 12000)
	f.SetCellValue("Sheet1", "A4", "销售部")
	f.SetCellValue("Sheet1", "B4", 9000)

	// Group by column name "部门", sum of "工资"
	resultSheet, err := GroupBy(f, "Sheet1", "部门", "工资", "sum")
	require.NoError(t, err)
	require.NotEmpty(t, resultSheet)

	// Verify result sheet exists
	sheets := f.GetSheetList()
	require.Contains(t, sheets, resultSheet)
}

func TestResolveColumnRef(t *testing.T) {
	f := newTestFile(t)

	// Set up header row
	f.SetCellValue("Sheet1", "A1", "姓名")
	f.SetCellValue("Sheet1", "B1", "部门")
	f.SetCellValue("Sheet1", "C1", "工资")

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"column letter A", "A", "A"},
		{"column letter B", "B", "B"},
		{"column letter AB", "AB", "AB"},
		{"column name 姓名", "姓名", "A"},
		{"column name 部门", "部门", "B"},
		{"column name 工资", "工资", "C"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := resolveColumnRef(f, "Sheet1", tt.input)
			require.NoError(t, err)
			require.Equal(t, tt.expected, result)
		})
	}

	// Test column not found
	_, err := resolveColumnRef(f, "Sheet1", "不存在")
	require.Error(t, err)
}
