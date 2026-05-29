package chart

import (
	"fmt"
	"strings"

	"github.com/xuri/excelize/v2"
)

// ChartType represents the type of chart.
type ChartType string

const (
	ChartTypeBar    ChartType = "bar"
	ChartTypeLine   ChartType = "line"
	ChartTypePie    ChartType = "pie"
	ChartTypeScatter ChartType = "scatter"
)

// ChartInfo contains information about a chart to create.
type ChartInfo struct {
	Type    ChartType `json:"type"`
	Title   string    `json:"title"`
	XAxis   string    `json:"x_axis"`
	YAxis   string    `json:"y_axis"`
	Series  []Series  `json:"series"`
}

// Series represents a data series in a chart.
type Series struct {
	Name string `json:"name"`
	XValues string `json:"x_values"`
	YValues string `json:"y_values"`
}

// ChartRecommendation contains chart type recommendations based on data analysis.
type ChartRecommendation struct {
	RecommendedType ChartType `json:"recommended_type"`
	Reason          string    `json:"reason"`
	Alternatives    []ChartType `json:"alternatives"`
	DataSummary     DataSummary `json:"data_summary"`
}

// DataSummary contains summary information about the data.
type DataSummary struct {
	RowCount    int        `json:"row_count"`
	ColCount    int        `json:"col_count"`
	ColumnTypes []ColType  `json:"column_types"`
	HasDate     bool       `json:"has_date"`
	HasCategory bool       `json:"has_category"`
	HasNumeric  bool       `json:"has_numeric"`
}

// ColType represents column type information.
type ColType struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// CreateChart creates a chart in the workbook.
func CreateChart(f *excelize.File, sheet string, info ChartInfo) error {
	// Map chart type to excelize type
	chartType := mapChartType(info.Type)

	// Build series
	series := make([]excelize.ChartSeries, len(info.Series))
	for i, s := range info.Series {
		series[i] = excelize.ChartSeries{
			Name:       s.Name,
			Categories: fmt.Sprintf("'%s'!%s", sheet, s.XValues),
			Values:     fmt.Sprintf("'%s'!%s", sheet, s.YValues),
		}
	}

	// Create chart
	return f.AddChart(sheet, "E1", &excelize.Chart{
		Type: chartType,
		Title: []excelize.RichTextRun{
			{
				Text: info.Title,
			},
		},
		XAxis: excelize.ChartAxis{
			Title: []excelize.RichTextRun{
				{
					Text: info.XAxis,
				},
			},
		},
		YAxis: excelize.ChartAxis{
			Title: []excelize.RichTextRun{
				{
					Text: info.YAxis,
				},
			},
		},
		Series: series,
	})
}

// AutoChart analyzes data and returns chart recommendations.
func AutoChart(f *excelize.File, sheet string) (*ChartRecommendation, error) {
	rows, err := f.GetRows(sheet)
	if err != nil {
		return nil, fmt.Errorf("get rows: %w", err)
	}

	if len(rows) < 2 {
		return nil, fmt.Errorf("insufficient data for chart recommendation")
	}

	// Analyze column types
	colTypes := analyzeColumnTypes(rows)
	summary := buildDataSummary(rows, colTypes)

	// Determine recommended chart type
	rec := recommendChartType(summary)

	return &ChartRecommendation{
		RecommendedType: rec,
		Reason:          getRecommendationReason(rec, summary),
		Alternatives:    getAlternativeCharts(rec),
		DataSummary:     summary,
	}, nil
}

// mapChartType maps our chart type to excelize chart type.
func mapChartType(t ChartType) excelize.ChartType {
	switch t {
	case ChartTypeBar:
		return excelize.Bar
	case ChartTypeLine:
		return excelize.Line
	case ChartTypePie:
		return excelize.Pie
	case ChartTypeScatter:
		return excelize.Scatter
	default:
		return excelize.Bar
	}
}

// analyzeColumnTypes analyzes the types of each column.
func analyzeColumnTypes(rows [][]string) []ColType {
	if len(rows) == 0 {
		return nil
	}

	colCount := len(rows[0])
	result := make([]ColType, colCount)

	for col := 0; col < colCount; col++ {
		name := fmt.Sprintf("Column%d", col+1)
		if col < len(rows[0]) {
			name = rows[0][col]
		}

		// Sample data to determine type
		types := make(map[string]int)
		sampleSize := 10
		if len(rows)-1 < sampleSize {
			sampleSize = len(rows) - 1
		}

		for i := 1; i <= sampleSize; i++ {
			if col < len(rows[i]) {
				types[detectValueType(rows[i][col])]++
			}
		}

		// Determine dominant type
		colType := "string"
		maxCount := 0
		for t, count := range types {
			if count > maxCount {
				maxCount = count
				colType = t
			}
		}

		result[col] = ColType{Name: name, Type: colType}
	}

	return result
}

// detectValueType detects the type of a value.
func detectValueType(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return "empty"
	}

	// Check date
	dateFormats := []string{"2006-01-02", "2006/01/02", "01/02/2006"}
	for _, f := range dateFormats {
		if _, err := parseDate(v, f); err == nil {
			return "date"
		}
	}

	// Check number
	if isNumeric(v) {
		return "number"
	}

	return "string"
}

// parseDate tries to parse a date string.
func parseDate(v, format string) (interface{}, error) {
	// Simple date parsing check
	if len(v) == 10 && (v[4] == '-' || v[4] == '/') {
		return v, nil
	}
	return nil, fmt.Errorf("not a date")
}

// isNumeric checks if a string is numeric.
func isNumeric(v string) bool {
	if v == "" {
		return false
	}
	dotCount := 0
	for i, c := range v {
		if c == '.' {
			dotCount++
			if dotCount > 1 {
				return false
			}
		} else if c == '-' && i == 0 {
			continue
		} else if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// buildDataSummary builds a summary of the data.
func buildDataSummary(rows [][]string, colTypes []ColType) DataSummary {
	summary := DataSummary{
		RowCount:    len(rows) - 1,
		ColCount:    len(rows[0]),
		ColumnTypes: colTypes,
	}

	for _, ct := range colTypes {
		switch ct.Type {
		case "date":
			summary.HasDate = true
		case "number":
			summary.HasNumeric = true
		case "string":
			summary.HasCategory = true
		}
	}

	return summary
}

// recommendChartType recommends a chart type based on data summary.
func recommendChartType(summary DataSummary) ChartType {
	// Date + Number → Line chart
	if summary.HasDate && summary.HasNumeric {
		return ChartTypeLine
	}

	// Category + Number → Bar chart
	if summary.HasCategory && summary.HasNumeric {
		// If few categories, pie chart might be better
		if summary.RowCount <= 10 {
			return ChartTypePie
		}
		return ChartTypeBar
	}

	// Two numeric columns → Scatter
	if summary.ColCount == 2 && summary.HasNumeric {
		return ChartTypeScatter
	}

	// Default to bar chart
	return ChartTypeBar
}

// getRecommendationReason returns the reason for the recommendation.
func getRecommendationReason(chartType ChartType, summary DataSummary) string {
	switch chartType {
	case ChartTypeLine:
		return "Time series data detected - line chart shows trends over time"
	case ChartTypeBar:
		if summary.HasCategory {
			return "Category and numeric data - bar chart for comparison"
		}
		return "Numeric data suitable for bar chart comparison"
	case ChartTypePie:
		return "Few categories with numeric values - pie chart shows proportions"
	case ChartTypeScatter:
		return "Two numeric columns - scatter chart shows correlation"
	default:
		return "Default chart type for the data"
	}
}

// getAlternativeCharts returns alternative chart types.
func getAlternativeCharts(primary ChartType) []ChartType {
	switch primary {
	case ChartTypeLine:
		return []ChartType{ChartTypeBar, ChartTypeScatter}
	case ChartTypeBar:
		return []ChartType{ChartTypeLine, ChartTypePie}
	case ChartTypePie:
		return []ChartType{ChartTypeBar}
	case ChartTypeScatter:
		return []ChartType{ChartTypeLine}
	default:
		return []ChartType{ChartTypeBar, ChartTypeLine}
	}
}
