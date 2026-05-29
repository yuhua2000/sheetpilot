package io

import (
	"os"
	"path/filepath"
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

func TestExportCSV(t *testing.T) {
	f := newTestFile(t)
	tmpDir := t.TempDir()

	// Set up test data
	f.SetCellValue("Sheet1", "A1", "Name")
	f.SetCellValue("Sheet1", "B1", "Age")
	f.SetCellValue("Sheet1", "A2", "Alice")
	f.SetCellValue("Sheet1", "B2", 25)
	f.SetCellValue("Sheet1", "A3", "Bob")
	f.SetCellValue("Sheet1", "B3", 30)

	outputPath := filepath.Join(tmpDir, "test.csv")
	err := ExportCSV(f, "Sheet1", outputPath)
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(outputPath)
	require.NoError(t, err)

	// Read and verify content
	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	require.Contains(t, string(content), "Name")
	require.Contains(t, string(content), "Alice")
}

func TestImportCSV(t *testing.T) {
	f := newTestFile(t)
	tmpDir := t.TempDir()

	// Create test CSV file
	csvPath := filepath.Join(tmpDir, "test.csv")
	err := os.WriteFile(csvPath, []byte("Name,Age\nAlice,25\nBob,30"), 0644)
	require.NoError(t, err)

	// Import CSV
	err = ImportCSV(f, csvPath, "Imported")
	require.NoError(t, err)

	// Verify data
	sheets := f.GetSheetList()
	require.Contains(t, sheets, "Imported")

	val, _ := f.GetCellValue("Imported", "A1")
	require.Equal(t, "Name", val)

	val, _ = f.GetCellValue("Imported", "A2")
	require.Equal(t, "Alice", val)
}

func TestExportJSON(t *testing.T) {
	f := newTestFile(t)

	// Set up test data
	f.SetCellValue("Sheet1", "A1", "Name")
	f.SetCellValue("Sheet1", "B1", "Age")
	f.SetCellValue("Sheet1", "A2", "Alice")
	f.SetCellValue("Sheet1", "B2", "25")

	result, err := ExportJSON(f, "Sheet1")
	require.NoError(t, err)
	require.Contains(t, result, "Name")
	require.Contains(t, result, "Alice")
}

func TestExportImportRoundTrip(t *testing.T) {
	f := newTestFile(t)
	tmpDir := t.TempDir()

	// Set up test data
	f.SetCellValue("Sheet1", "A1", "Product")
	f.SetCellValue("Sheet1", "B1", "Price")
	f.SetCellValue("Sheet1", "A2", "Apple")
	f.SetCellValue("Sheet1", "B2", "1.5")
	f.SetCellValue("Sheet1", "A3", "Banana")
	f.SetCellValue("Sheet1", "B3", "0.8")

	// Export to CSV
	csvPath := filepath.Join(tmpDir, "roundtrip.csv")
	err := ExportCSV(f, "Sheet1", csvPath)
	require.NoError(t, err)

	// Import into new sheet
	err = ImportCSV(f, csvPath, "RoundTrip")
	require.NoError(t, err)

	// Verify data matches
	val1, _ := f.GetCellValue("Sheet1", "A2")
	val2, _ := f.GetCellValue("RoundTrip", "A2")
	require.Equal(t, val1, val2)

	val1, _ = f.GetCellValue("Sheet1", "B3")
	val2, _ = f.GetCellValue("RoundTrip", "B3")
	require.Equal(t, val1, val2)
}
