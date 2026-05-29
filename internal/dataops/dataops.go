package dataops

import (
	"fmt"
	"sort"
	"strings"

	"github.com/xuri/excelize/v2"
)

// AddComputedColumn adds a new computed column with a formula.
func AddComputedColumn(f *excelize.File, sheet, colName, formulaTmpl string) error {
	rows, err := f.GetRows(sheet)
	if err != nil {
		return fmt.Errorf("get rows: %w", err)
	}

	if len(rows) == 0 {
		return fmt.Errorf("sheet is empty")
	}

	// Find next empty column
	colIdx := len(rows[0]) + 1
	colName2, err := excelize.ColumnNumberToName(colIdx)
	if err != nil {
		return fmt.Errorf("column number to name: %w", err)
	}

	// Set header
	headerCell := fmt.Sprintf("%s1", colName2)
	f.SetCellValue(sheet, headerCell, colName)

	// Fill formulas for data rows
	for row := 2; row <= len(rows); row++ {
		cell := fmt.Sprintf("%s%d", colName2, row)
		formula := replaceColumnRefs(formulaTmpl, rows[0], row)
		f.SetCellFormula(sheet, cell, formula)
	}

	return nil
}

// FillMissingValues fills empty cells with a default value.
func FillMissingValues(f *excelize.File, sheet, col, defaultValue string) error {
	rows, err := f.GetRows(sheet)
	if err != nil {
		return fmt.Errorf("get rows: %w", err)
	}

	for i := range rows {
		cell := fmt.Sprintf("%s%d", col, i+1)
		val, _ := f.GetCellValue(sheet, cell)
		if strings.TrimSpace(val) == "" {
			f.SetCellValue(sheet, cell, defaultValue)
		}
	}

	return nil
}

// ReplaceValues replaces all occurrences of old value with new value in a column.
func ReplaceValues(f *excelize.File, sheet, col, oldValue, newValue string) error {
	rows, err := f.GetRows(sheet)
	if err != nil {
		return fmt.Errorf("get rows: %w", err)
	}

	colIdx, err := excelize.ColumnNameToNumber(col)
	if err != nil {
		return fmt.Errorf("column name to number: %w", err)
	}

	for i, row := range rows {
		if colIdx-1 < len(row) && row[colIdx-1] == oldValue {
			cell := fmt.Sprintf("%s%d", col, i+1)
			f.SetCellValue(sheet, cell, newValue)
		}
	}

	return nil
}

// CleanupSheet performs basic data cleanup operations.
func CleanupSheet(f *excelize.File, sheet string) error {
	// Remove empty rows
	rows, err := f.GetRows(sheet)
	if err != nil {
		return fmt.Errorf("get rows: %w", err)
	}

	// Write back non-empty rows
	newRow := 1
	for _, row := range rows {
		if !isEmptyRow(row) {
			for j, cell := range row {
				col, _ := excelize.ColumnNumberToName(j + 1)
				f.SetCellValue(sheet, fmt.Sprintf("%s%d", col, newRow), strings.TrimSpace(cell))
			}
			newRow++
		}
	}

	return nil
}

// Deduplicate removes duplicate rows based on specified columns.
func Deduplicate(f *excelize.File, sheet string, cols []string) error {
	rows, err := f.GetRows(sheet)
	if err != nil {
		return fmt.Errorf("get rows: %w", err)
	}

	if len(rows) < 2 {
		return nil
	}

	// Build column index map
	colIndices := make([]int, len(cols))
	for i, col := range cols {
		idx, err := excelize.ColumnNameToNumber(col)
		if err != nil {
			return fmt.Errorf("column %s not found: %w", col, err)
		}
		colIndices[i] = idx - 1
	}

	// Track seen values
	seen := make(map[string]bool)
	uniqueRows := [][]string{rows[0]} // Keep header

	for _, row := range rows[1:] {
		key := buildRowKey(row, colIndices)
		if !seen[key] {
			seen[key] = true
			uniqueRows = append(uniqueRows, row)
		}
	}

	// Write back unique rows
	for i, row := range uniqueRows {
		for j, cell := range row {
			col, _ := excelize.ColumnNumberToName(j + 1)
			f.SetCellValue(sheet, fmt.Sprintf("%s%d", col, i+1), cell)
		}
	}

	// Clear remaining rows
	for i := len(uniqueRows); i < len(rows); i++ {
		for j := range rows[0] {
			col, _ := excelize.ColumnNumberToName(j + 1)
			f.SetCellValue(sheet, fmt.Sprintf("%s%d", col, i+1), nil)
		}
	}

	return nil
}

// SortTable sorts the table by specified columns.
func SortTable(f *excelize.File, sheet string, sortCols []string, ascending bool) error {
	rows, err := f.GetRows(sheet)
	if err != nil {
		return fmt.Errorf("get rows: %w", err)
	}

	if len(rows) < 2 {
		return nil
	}

	// Build column index map
	colIndices := make([]int, len(sortCols))
	for i, col := range sortCols {
		idx, err := excelize.ColumnNameToNumber(col)
		if err != nil {
			return fmt.Errorf("column %s not found: %w", col, err)
		}
		colIndices[i] = idx - 1
	}

	// Sort data rows (skip header)
	dataRows := rows[1:]
	sort.Slice(dataRows, func(i, j int) bool {
		for _, colIdx := range colIndices {
			valI := ""
			valJ := ""
			if colIdx < len(dataRows[i]) {
				valI = dataRows[i][colIdx]
			}
			if colIdx < len(dataRows[j]) {
				valJ = dataRows[j][colIdx]
			}
			if valI != valJ {
				if ascending {
					return valI < valJ
				}
				return valI > valJ
			}
		}
		return false
	})

	// Write back
	for i, row := range rows {
		for j, cell := range row {
			col, _ := excelize.ColumnNumberToName(j + 1)
			if i == 0 {
				f.SetCellValue(sheet, fmt.Sprintf("%s%d", col, i+1), cell)
			} else {
				f.SetCellValue(sheet, fmt.Sprintf("%s%d", col, i+1), dataRows[i-1][j])
			}
		}
	}

	return nil
}

// FilterRows filters rows based on a condition.
func FilterRows(f *excelize.File, sheet, col, op, value string) ([][]string, error) {
	rows, err := f.GetRows(sheet)
	if err != nil {
		return nil, fmt.Errorf("get rows: %w", err)
	}

	if len(rows) < 2 {
		return [][]string{}, nil
	}

	colIdx, err := excelize.ColumnNameToNumber(col)
	if err != nil {
		return nil, fmt.Errorf("column name to number: %w", err)
	}

	result := [][]string{rows[0]} // Include header
	for _, row := range rows[1:] {
		if colIdx-1 < len(row) && matchCondition(row[colIdx-1], op, value) {
			result = append(result, row)
		}
	}

	return result, nil
}

// GroupBy groups rows by a column and applies aggregation.
// GroupBy groups rows by a column and applies aggregation using Excel formulas.
// Result is written to a new sheet with formulas that auto-update.
func GroupBy(f *excelize.File, sheet, groupCol, aggCol, aggFunc string) (string, error) {
	rows, err := f.GetRows(sheet)
	if err != nil {
		return "", fmt.Errorf("get rows: %w", err)
	}

	if len(rows) < 2 {
		return "", fmt.Errorf("sheet has no data rows")
	}

	groupIdx, err := excelize.ColumnNameToNumber(groupCol)
	if err != nil {
		return "", fmt.Errorf("column name to number: %w", err)
	}

	// Get actual column names from header row
	groupColName := groupCol
	aggColName := aggCol
	if groupIdx-1 < len(rows[0]) {
		groupColName = rows[0][groupIdx-1]
	}
	aggIdx, _ := excelize.ColumnNameToNumber(aggCol)
	if aggIdx-1 < len(rows[0]) {
		aggColName = rows[0][aggIdx-1]
	}

	// Get unique group values
	uniqueGroups := getUniqueValues(rows[1:], groupIdx-1)

	// Create result sheet
	resultSheet := sheet + "_groupby"
	f.NewSheet(resultSheet)

	// Write headers
	f.SetCellValue(resultSheet, "A1", groupColName)
	f.SetCellValue(resultSheet, "B1", fmt.Sprintf("%s(%s)", aggFunc, aggColName))

	// Write group values and formulas
	for i, group := range uniqueGroups {
		row := i + 2
		f.SetCellValue(resultSheet, fmt.Sprintf("A%d", row), group)

		// Build Excel formula (SUMIFS, COUNTIFS, etc.)
		formula := buildGroupByFormula(aggFunc, sheet, groupCol, aggCol, group, rows)
		f.SetCellFormula(resultSheet, fmt.Sprintf("B%d", row), formula)
	}

	return resultSheet, nil
}

// getUniqueValues returns unique values from a column.
func getUniqueValues(rows [][]string, colIdx int) []string {
	seen := make(map[string]bool)
	var result []string
	for _, row := range rows {
		if colIdx < len(row) {
			val := row[colIdx]
			if !seen[val] {
				seen[val] = true
				result = append(result, val)
			}
		}
	}
	return result
}

// buildGroupByFormula creates an Excel formula for group by aggregation.
func buildGroupByFormula(aggFunc, sheet, groupCol, aggCol, groupValue string, rows [][]string) string {
	// Get data range (skip header)
	dataRange := fmt.Sprintf("%s2:%s%d", groupCol, groupCol, len(rows))

	// Build criteria - need to handle the group value
	criteria := fmt.Sprintf("\"%s\"", groupValue)

	switch strings.ToLower(aggFunc) {
	case "sum":
		return fmt.Sprintf("=SUMIFS(%s!%s:%s,%s!%s:%s,%s)",
			sheet, aggCol, aggCol, sheet, dataRange, dataRange, criteria)
	case "count":
		return fmt.Sprintf("=COUNTIF(%s!%s:%s,%s)",
			sheet, dataRange, dataRange, criteria)
	case "avg", "average":
		return fmt.Sprintf("=AVERAGEIF(%s!%s:%s,%s,%s!%s:%s)",
			sheet, dataRange, dataRange, criteria, sheet, aggCol, aggCol)
	case "min":
		return fmt.Sprintf("=MINIFS(%s!%s:%s,%s!%s:%s,%s)",
			sheet, aggCol, aggCol, sheet, dataRange, dataRange, criteria)
	case "max":
		return fmt.Sprintf("=MAXIFS(%s!%s:%s,%s!%s:%s,%s)",
			sheet, aggCol, aggCol, sheet, dataRange, dataRange, criteria)
	default:
		return fmt.Sprintf("=SUMIFS(%s!%s:%s,%s!%s:%s,%s)",
			sheet, aggCol, aggCol, sheet, dataRange, dataRange, criteria)
	}
}

// Helper functions

func replaceColumnRefs(formula string, headers []string, row int) string {
	result := formula
	for i, header := range headers {
		col, _ := excelize.ColumnNumberToName(i + 1)
		placeholder := fmt.Sprintf("{%s}", header)
		replacement := fmt.Sprintf("%s%d", col, row)
		result = strings.ReplaceAll(result, placeholder, replacement)
	}
	return result
}

func isEmptyRow(row []string) bool {
	for _, cell := range row {
		if strings.TrimSpace(cell) != "" {
			return false
		}
	}
	return true
}

func buildRowKey(row []string, colIndices []int) string {
	parts := make([]string, len(colIndices))
	for i, idx := range colIndices {
		if idx < len(row) {
			parts[i] = row[idx]
		}
	}
	return strings.Join(parts, "|")
}

func matchCondition(cell, op, value string) bool {
	switch op {
	case "=":
		return cell == value
	case "!=":
		return cell != value
	case ">":
		return cell > value
	case "<":
		return cell < value
	case ">=":
		return cell >= value
	case "<=":
		return cell <= value
	case "contains":
		return strings.Contains(cell, value)
	default:
		return false
	}
}

func parseNumber(s string) float64 {
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}

func formatNumber(f float64) string {
	return fmt.Sprintf("%.2f", f)
}

func aggregate(values []float64, fn string) float64 {
	if len(values) == 0 {
		return 0
	}

	switch fn {
	case "sum":
		sum := 0.0
		for _, v := range values {
			sum += v
		}
		return sum
	case "avg":
		return aggregate(values, "sum") / float64(len(values))
	case "count":
		return float64(len(values))
	case "min":
		min := values[0]
		for _, v := range values[1:] {
			if v < min {
				min = v
			}
		}
		return min
	case "max":
		max := values[0]
		for _, v := range values[1:] {
			if v > max {
				max = v
			}
		}
		return max
	default:
		return 0
	}
}

// SplitSheet splits a sheet into multiple sheets based on a column value.
func SplitSheet(f *excelize.File, sheet, col string) ([]string, error) {
	rows, err := f.GetRows(sheet)
	if err != nil {
		return nil, fmt.Errorf("get rows: %w", err)
	}

	if len(rows) < 2 {
		return nil, fmt.Errorf("sheet has no data rows")
	}

	colIdx, err := excelize.ColumnNameToNumber(col)
	if err != nil {
		return nil, fmt.Errorf("column name to number: %w", err)
	}

	// Group rows by column value
	groups := make(map[string][][]string)
	for _, row := range rows[1:] {
		if colIdx-1 < len(row) {
			key := row[colIdx-1]
			groups[key] = append(groups[key], row)
		}
	}

	// Create new sheets for each group
	var result []string
	for key, groupRows := range groups {
		newSheet := fmt.Sprintf("%s_%s", sheet, key)
		f.NewSheet(newSheet)

		// Write header
		for j, cell := range rows[0] {
			colName, _ := excelize.ColumnNumberToName(j + 1)
			f.SetCellValue(newSheet, fmt.Sprintf("%s1", colName), cell)
		}

		// Write data rows
		for i, row := range groupRows {
			for j, cell := range row {
				colName, _ := excelize.ColumnNumberToName(j + 1)
				f.SetCellValue(newSheet, fmt.Sprintf("%s%d", colName, i+2), cell)
			}
		}

		result = append(result, newSheet)
	}

	return result, nil
}

// MergeSheets merges multiple sheets into one.
func MergeSheets(f *excelize.File, sheets []string, destSheet string) error {
	if len(sheets) == 0 {
		return fmt.Errorf("no sheets to merge")
	}

	// Create destination sheet
	f.NewSheet(destSheet)

	// Track if header is written
	headerWritten := false
	destRow := 1

	for _, sheet := range sheets {
		rows, err := f.GetRows(sheet)
		if err != nil {
			return fmt.Errorf("get rows from %s: %w", sheet, err)
		}

		if len(rows) == 0 {
			continue
		}

		startIdx := 0
		if !headerWritten {
			// Write header from first sheet
			for j, cell := range rows[0] {
				colName, _ := excelize.ColumnNumberToName(j + 1)
				f.SetCellValue(destSheet, fmt.Sprintf("%s%d", colName, destRow), cell)
			}
			headerWritten = true
			destRow++
			startIdx = 1
		} else {
			// Skip header for subsequent sheets
			startIdx = 1
		}

		// Write data rows
		for i := startIdx; i < len(rows); i++ {
			for j, cell := range rows[i] {
				colName, _ := excelize.ColumnNumberToName(j + 1)
				f.SetCellValue(destSheet, fmt.Sprintf("%s%d", colName, destRow), cell)
			}
			destRow++
		}
	}

	return nil
}
