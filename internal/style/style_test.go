package style

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

func TestAutoFitColumns(t *testing.T) {
	f := newTestFile(t)

	// Set up test data with varying lengths
	f.SetCellValue("Sheet1", "A1", "Name")
	f.SetCellValue("Sheet1", "B1", "Description")
	f.SetCellValue("Sheet1", "A2", "Alice")
	f.SetCellValue("Sheet1", "B2", "A very long description that should make the column wider")

	err := AutoFitColumns(f, "Sheet1")
	require.NoError(t, err)

	// Verify columns were adjusted (no easy way to check exact width, just no error)
}

func TestAutoFitRows(t *testing.T) {
	f := newTestFile(t)

	f.SetCellValue("Sheet1", "A1", "Name")
	f.SetCellValue("Sheet1", "A2", "Alice")

	err := AutoFitRows(f, "Sheet1")
	require.NoError(t, err)
}

func TestSetNumberFormat(t *testing.T) {
	f := newTestFile(t)

	f.SetCellValue("Sheet1", "A1", 1234.5678)

	err := SetNumberFormat(f, "Sheet1", "A1", "#,##0.00")
	require.NoError(t, err)

	// Verify style was applied (no easy way to check format, just no error)
}

func TestSetNumberFormat_Percent(t *testing.T) {
	f := newTestFile(t)

	f.SetCellValue("Sheet1", "A1", 0.85)

	err := SetNumberFormat(f, "Sheet1", "A1", "0.00%")
	require.NoError(t, err)
}

func TestSetNumberFormat_Date(t *testing.T) {
	f := newTestFile(t)

	f.SetCellValue("Sheet1", "A1", "2024-01-15")

	err := SetNumberFormat(f, "Sheet1", "A1", "yyyy-mm-dd")
	require.NoError(t, err)
}

func TestGetNumFmtID(t *testing.T) {
	tests := []struct {
		format string
		want   int
	}{
		{"General", 0},
		{"0", 1},
		{"0.00", 2},
		{"#,##0", 3},
		{"#,##0.00", 4},
		{"0%", 9},
		{"0.00%", 10},
		{"yyyy-mm-dd", 14},
		{"unknown", 0},
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			got := getNumFmtID(tt.format)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestSetStyle(t *testing.T) {
	f := newTestFile(t)

	opts := StyleOptions{
		Bold:      "true",
		BgColor:   "#FF0000",
		FontColor: "#FFFFFF",
		Align:     "center",
		Border:    "thin",
	}

	err := SetStyle(f, "Sheet1", "A1", opts)
	require.NoError(t, err)
}

func TestSetConditionalFormatting(t *testing.T) {
	f := newTestFile(t)
	f.SetCellValue("Sheet1", "A1", 10)
	f.SetCellValue("Sheet1", "A2", 20)
	f.SetCellValue("Sheet1", "A3", 5)

	err := SetConditionalFormatting(f, "Sheet1", "A1:A3", "less_than", "10", "#FF0000", "#FFFFFF")
	require.NoError(t, err)
}

func TestFreezePanes(t *testing.T) {
	f := newTestFile(t)

	err := FreezePanes(f, "Sheet1", "1", "0")
	require.NoError(t, err)
}

func TestAddFilter(t *testing.T) {
	f := newTestFile(t)
	f.SetCellValue("Sheet1", "A1", "Name")
	f.SetCellValue("Sheet1", "B1", "Value")
	f.SetCellValue("Sheet1", "A2", "A")
	f.SetCellValue("Sheet1", "B2", 10)

	err := AddFilter(f, "Sheet1", "A1:B2")
	require.NoError(t, err)
}

func TestFormatAsTable(t *testing.T) {
	f := newTestFile(t)
	f.SetCellValue("Sheet1", "A1", "Name")
	f.SetCellValue("Sheet1", "B1", "Value")
	f.SetCellValue("Sheet1", "A2", "A")
	f.SetCellValue("Sheet1", "B2", 10)
	f.SetCellValue("Sheet1", "A3", "B")
	f.SetCellValue("Sheet1", "B3", 20)

	err := FormatAsTable(f, "Sheet1", "A1:B3", "#4472C4", true)
	require.NoError(t, err)
}
