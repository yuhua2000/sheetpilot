package analysis

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

// TableInfo contains information about a detected table.
type TableInfo struct {
	Range     string   `json:"range"`
	StartCell string   `json:"start_cell"`
	EndCell   string   `json:"end_cell"`
	RowCount  int      `json:"row_count"`
	ColCount  int      `json:"col_count"`
	Columns   []string `json:"columns"`
}

// ColumnType represents the detected type of a column.
type ColumnType struct {
	Name    string   `json:"name"`
	Type    string   `json:"type"`
	Samples []string `json:"samples"`
}

// SheetOverview contains a summary of the sheet.
type SheetOverview struct {
	SheetName string        `json:"sheet_name"`
	RowCount  int           `json:"row_count"`
	ColCount  int           `json:"col_count"`
	Columns   []ColumnType  `json:"columns"`
	Samples   [][]string    `json:"samples"`
}

// GetTableInfo detects table boundaries and returns table information.
func GetTableInfo(f *excelize.File, sheet string) (*TableInfo, error) {
	rows, err := f.GetRows(sheet)
	if err != nil {
		return nil, fmt.Errorf("get rows: %w", err)
	}

	if len(rows) == 0 {
		return &TableInfo{
			Range:     "A1",
			StartCell: "A1",
			EndCell:   "A1",
			RowCount:  0,
			ColCount:  0,
			Columns:   []string{},
		}, nil
	}

	// Find max columns in first row (header)
	colCount := len(rows[0])

	// Find last non-empty row
	rowCount := len(rows)
	for i := len(rows) - 1; i >= 0; i-- {
		if !isEmptyRow(rows[i]) {
			break
		}
		rowCount--
	}

	// Get column names from first row
	columns := make([]string, colCount)
	for i := 0; i < colCount; i++ {
		if i < len(rows[0]) {
			columns[i] = rows[0][i]
		} else {
			columns[i] = fmt.Sprintf("Column%d", i+1)
		}
	}

	startCell := "A1"
	endCell, _ := excelize.CoordinatesToCellName(colCount, rowCount)

	return &TableInfo{
		Range:     fmt.Sprintf("%s:%s", startCell, endCell),
		StartCell: startCell,
		EndCell:   endCell,
		RowCount:  rowCount,
		ColCount:  colCount,
		Columns:   columns,
	}, nil
}

// GetColumnTypes analyzes column types by sampling data.
func GetColumnTypes(f *excelize.File, sheet string, sampleSize int) ([]ColumnType, error) {
	rows, err := f.GetRows(sheet)
	if err != nil {
		return nil, fmt.Errorf("get rows: %w", err)
	}

	if len(rows) < 2 {
		return []ColumnType{}, nil
	}

	colCount := len(rows[0])
	if sampleSize <= 0 {
		sampleSize = 100
	}

	result := make([]ColumnType, colCount)
	for col := 0; col < colCount; col++ {
		name := fmt.Sprintf("Column%d", col+1)
		if col < len(rows[0]) {
			name = rows[0][col]
		}

		samples := make([]string, 0, sampleSize)
		values := make([]string, 0, sampleSize)

		for row := 1; row < len(rows) && row <= sampleSize; row++ {
			if col < len(rows[row]) {
				val := rows[row][col]
				values = append(values, val)
				if len(samples) < 5 {
					samples = append(samples, val)
				}
			}
		}

		colType := detectColumnType(values)

		result[col] = ColumnType{
			Name:    name,
			Type:    colType,
			Samples: samples,
		}
	}

	return result, nil
}

// GetSheetOverview returns a summary of the sheet.
func GetSheetOverview(f *excelize.File, sheet string) (*SheetOverview, error) {
	rows, err := f.GetRows(sheet)
	if err != nil {
		return nil, fmt.Errorf("get rows: %w", err)
	}

	if len(rows) == 0 {
		return &SheetOverview{
			SheetName: sheet,
			RowCount:  0,
			ColCount:  0,
			Columns:   []ColumnType{},
			Samples:   [][]string{},
		}, nil
	}

	colCount := len(rows[0])

	// Get column types
	columns, err := GetColumnTypes(f, sheet, 100)
	if err != nil {
		return nil, err
	}

	// Get sample rows (first 5 data rows)
	sampleRows := make([][]string, 0, 5)
	for i := 1; i < len(rows) && i <= 5; i++ {
		row := make([]string, colCount)
		for j := 0; j < colCount; j++ {
			if j < len(rows[i]) {
				row[j] = rows[i][j]
			}
		}
		sampleRows = append(sampleRows, row)
	}

	return &SheetOverview{
		SheetName: sheet,
		RowCount:  len(rows),
		ColCount:  colCount,
		Columns:   columns,
		Samples:   sampleRows,
	}, nil
}

// detectColumnType determines the type of a column based on sample values.
func detectColumnType(values []string) string {
	if len(values) == 0 {
		return "string"
	}

	boolCount := 0
	dateCount := 0
	numberCount := 0
	percentCount := 0
	currencyCount := 0

	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}

		// Check bool
		if isBool(v) {
			boolCount++
			continue
		}

		// Check percent
		if strings.HasSuffix(v, "%") {
			percentCount++
			continue
		}

		// Check currency
		if isCurrency(v) {
			currencyCount++
			continue
		}

		// Check number
		if isNumber(v) {
			numberCount++
			continue
		}

		// Check date
		if isDate(v) {
			dateCount++
			continue
		}
	}

	total := len(values)
	if total == 0 {
		return "string"
	}

	// Return type with highest count (threshold 70%)
	threshold := float64(total) * 0.7

	if float64(boolCount) >= threshold {
		return "bool"
	}
	if float64(dateCount) >= threshold {
		return "date"
	}
	if float64(percentCount) >= threshold {
		return "percent"
	}
	if float64(currencyCount) >= threshold {
		return "currency"
	}
	if float64(numberCount) >= threshold {
		return "number"
	}

	return "string"
}

func isBool(v string) bool {
	lower := strings.ToLower(v)
	return lower == "true" || lower == "false" || lower == "是" || lower == "否"
}

func isNumber(v string) bool {
	_, err := strconv.ParseFloat(v, 64)
	return err == nil
}

func isCurrency(v string) bool {
	prefixes := []string{"¥", "$", "€", "£", "￥"}
	for _, p := range prefixes {
		if strings.HasPrefix(v, p) {
			return true
		}
	}
	return false
}

func isDate(v string) bool {
	formats := []string{
		"2006-01-02",
		"2006/01/02",
		"01/02/2006",
		"02-Jan-2006",
		"2006-01-02 15:04:05",
	}
	for _, f := range formats {
		if _, err := time.Parse(f, v); err == nil {
			return true
		}
	}
	return false
}

func isEmptyRow(row []string) bool {
	for _, cell := range row {
		if strings.TrimSpace(cell) != "" {
			return false
		}
	}
	return true
}
