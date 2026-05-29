package chart

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

func TestAutoChart(t *testing.T) {
	f := newTestFile(t)

	// Set up test data with date and numeric columns
	f.SetCellValue("Sheet1", "A1", "Date")
	f.SetCellValue("Sheet1", "B1", "Sales")
	f.SetCellValue("Sheet1", "A2", "2024-01-01")
	f.SetCellValue("Sheet1", "B2", 100)
	f.SetCellValue("Sheet1", "A3", "2024-01-02")
	f.SetCellValue("Sheet1", "B3", 200)

	rec, err := AutoChart(f, "Sheet1")
	require.NoError(t, err)
	require.NotNil(t, rec)
	require.NotEmpty(t, rec.RecommendedType)
	require.NotEmpty(t, rec.Reason)
	require.NotNil(t, rec.DataSummary)
}

func TestAutoChart_CategoryNumeric(t *testing.T) {
	f := newTestFile(t)

	// Set up test data with category and numeric columns
	f.SetCellValue("Sheet1", "A1", "Product")
	f.SetCellValue("Sheet1", "B1", "Revenue")
	f.SetCellValue("Sheet1", "A2", "Apple")
	f.SetCellValue("Sheet1", "B2", 1000)
	f.SetCellValue("Sheet1", "A3", "Banana")
	f.SetCellValue("Sheet1", "B3", 2000)
	f.SetCellValue("Sheet1", "A4", "Cherry")
	f.SetCellValue("Sheet1", "B4", 1500)

	rec, err := AutoChart(f, "Sheet1")
	require.NoError(t, err)
	require.NotNil(t, rec)
	// Should recommend pie or bar for few categories
	require.Contains(t, []ChartType{ChartTypePie, ChartTypeBar}, rec.RecommendedType)
}

func TestCreateChart(t *testing.T) {
	f := newTestFile(t)

	// Set up test data
	f.SetCellValue("Sheet1", "A1", "Category")
	f.SetCellValue("Sheet1", "B1", "Value")
	f.SetCellValue("Sheet1", "A2", "A")
	f.SetCellValue("Sheet1", "B2", 10)
	f.SetCellValue("Sheet1", "A3", "B")
	f.SetCellValue("Sheet1", "B3", 20)

	info := ChartInfo{
		Type:  ChartTypeBar,
		Title: "Test Chart",
		XAxis: "Category",
		YAxis: "Value",
		Series: []Series{
			{
				Name:    "Series 1",
				XValues: "A2:A3",
				YValues: "B2:B3",
			},
		},
	}

	err := CreateChart(f, "Sheet1", info)
	require.NoError(t, err)
}

func TestDetectValueType(t *testing.T) {
	tests := []struct {
		value    string
		expected string
	}{
		{"2024-01-01", "date"},
		{"123", "number"},
		{"12.34", "number"},
		{"-5", "number"},
		{"hello", "string"},
		{"", "empty"},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			result := detectValueType(tt.value)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestIsNumeric(t *testing.T) {
	tests := []struct {
		value    string
		expected bool
	}{
		{"123", true},
		{"12.34", true},
		{"-5", true},
		{"-5.5", true},
		{"abc", false},
		{"", false},
		{"12a", false},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			result := isNumeric(tt.value)
			require.Equal(t, tt.expected, result)
		})
	}
}
