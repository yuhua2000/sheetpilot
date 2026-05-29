package view

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

func TestHideShowSheet(t *testing.T) {
	f := newTestFile(t)

	// Create a second sheet
	f.NewSheet("Sheet2")

	// Hide sheet
	err := HideSheet(f, "Sheet2")
	require.NoError(t, err)

	// Show sheet
	err = ShowSheet(f, "Sheet2")
	require.NoError(t, err)
}

func TestHideShowRows(t *testing.T) {
	f := newTestFile(t)

	// Hide rows 1-3
	err := HideRows(f, "Sheet1", "1:3")
	require.NoError(t, err)

	// Show rows 1-3
	err = ShowRows(f, "Sheet1", "1:3")
	require.NoError(t, err)
}

func TestHideShowColumns(t *testing.T) {
	f := newTestFile(t)

	// Hide columns A-C
	err := HideColumns(f, "Sheet1", "A:C")
	require.NoError(t, err)

	// Show columns A-C
	err = ShowColumns(f, "Sheet1", "A:C")
	require.NoError(t, err)
}

func TestSetRowHeight(t *testing.T) {
	f := newTestFile(t)

	// Set row height for rows 1-5
	err := SetRowHeight(f, "Sheet1", "1:5", 30.0)
	require.NoError(t, err)
}

func TestSetColWidth(t *testing.T) {
	f := newTestFile(t)

	// Set column width for A-C
	err := SetColWidth(f, "Sheet1", "A:C", 15.0)
	require.NoError(t, err)
}

func TestProtectUnprotectSheet(t *testing.T) {
	f := newTestFile(t)

	// Protect sheet
	err := ProtectSheet(f, "Sheet1", "password123")
	require.NoError(t, err)

	// Unprotect sheet
	err = UnprotectSheet(f, "Sheet1", "password123")
	require.NoError(t, err)
}

func TestSetPrintArea(t *testing.T) {
	f := newTestFile(t)

	// Set print area
	err := SetPrintArea(f, "Sheet1", "A1:G50")
	require.NoError(t, err)
}

func TestSetHeaderFooter(t *testing.T) {
	f := newTestFile(t)

	// Set header and footer
	err := SetHeaderFooter(f, "Sheet1", &excelize.HeaderFooterOptions{
		OddHeader: "Test Header",
		OddFooter: "Test Footer",
	})
	require.NoError(t, err)
}

func TestSetDefinedName(t *testing.T) {
	f := newTestFile(t)

	// Set defined name
	err := SetDefinedName(f, "MyRange", "Sheet1!$A$1:$C$10", "")
	require.NoError(t, err)
}

func TestParseRange(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantStart int
		wantEnd   int
		wantErr   bool
	}{
		{"valid range", "1:5", 1, 5, false},
		{"single number", "1:1", 1, 1, false},
		{"invalid format", "1-5", 0, 0, true},
		{"invalid start", "a:5", 0, 0, true},
		{"invalid end", "1:b", 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end, err := parseRange(tt.input)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.wantStart, start)
				require.Equal(t, tt.wantEnd, end)
			}
		})
	}
}

func TestParseColRange(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantStart string
		wantEnd   string
		wantErr   bool
	}{
		{"valid range", "A:C", "A", "C", false},
		{"single column", "A:A", "A", "A", false},
		{"invalid format", "A-C", "", "", true},
		{"invalid column", "1:C", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end, err := parseColRange(tt.input)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.wantStart, start)
				require.Equal(t, tt.wantEnd, end)
			}
		})
	}
}
