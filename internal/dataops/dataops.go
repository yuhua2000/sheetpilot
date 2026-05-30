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
	if err := f.SetCellValue(sheet, headerCell, colName); err != nil {
		return fmt.Errorf("set header: %w", err)
	}

	// Fill formulas for data rows
	for row := 2; row <= len(rows); row++ {
		cell := fmt.Sprintf("%s%d", colName2, row)
		formula := replaceColumnRefs(formulaTmpl, rows[0], row)
		if err := f.SetCellFormula(sheet, cell, formula); err != nil {
			return fmt.Errorf("set formula at %s: %w", cell, err)
		}
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
			if err := f.SetCellValue(sheet, cell, defaultValue); err != nil {
				return fmt.Errorf("set cell %s: %w", cell, err)
			}
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
			if err := f.SetCellValue(sheet, cell, newValue); err != nil {
				return fmt.Errorf("set cell %s: %w", cell, err)
			}
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
				if err := f.SetCellValue(sheet, fmt.Sprintf("%s%d", col, newRow), strings.TrimSpace(cell)); err != nil {
					return fmt.Errorf("set cell: %w", err)
				}
			}
			newRow++
		}
	}

	return nil
}

// Deduplicate removes duplicate rows based on specified columns.
// cols can be column letters (A, B) or column names (部门, 工资).
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
		colIdx, err := resolveColumnIndex(f, sheet, col)
		if err != nil {
			return fmt.Errorf("column %s not found: %w", col, err)
		}
		colIndices[i] = colIdx
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
			if err := f.SetCellValue(sheet, fmt.Sprintf("%s%d", col, i+1), cell); err != nil {
				return fmt.Errorf("set cell: %w", err)
			}
		}
	}

	// Clear remaining rows
	for i := len(uniqueRows); i < len(rows); i++ {
		for j := range rows[0] {
			col, _ := excelize.ColumnNumberToName(j + 1)
			if err := f.SetCellValue(sheet, fmt.Sprintf("%s%d", col, i+1), nil); err != nil {
				return fmt.Errorf("clear cell: %w", err)
			}
		}
	}

	return nil
}

// SortTable sorts the table by specified columns.
// sortCols can be column letters (A, B) or column names (部门, 工资).
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
		colIdx, err := resolveColumnIndex(f, sheet, col)
		if err != nil {
			return fmt.Errorf("column %s not found: %w", col, err)
		}
		colIndices[i] = colIdx
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
				if err := f.SetCellValue(sheet, fmt.Sprintf("%s%d", col, i+1), cell); err != nil {
					return fmt.Errorf("set header: %w", err)
				}
			} else {
				if err := f.SetCellValue(sheet, fmt.Sprintf("%s%d", col, i+1), dataRows[i-1][j]); err != nil {
					return fmt.Errorf("set cell: %w", err)
				}
			}
		}
	}

	return nil
}

// FilterRows filters rows based on a condition.
// col can be column letter (A, B) or column name (部门, 工资).
func FilterRows(f *excelize.File, sheet, col, op, value string) ([][]string, error) {
	rows, err := f.GetRows(sheet)
	if err != nil {
		return nil, fmt.Errorf("get rows: %w", err)
	}

	if len(rows) < 2 {
		return [][]string{}, nil
	}

	colIdx, err := resolveColumnIndex(f, sheet, col)
	if err != nil {
		return nil, fmt.Errorf("column %s not found: %w", col, err)
	}

	result := [][]string{rows[0]} // Include header
	for _, row := range rows[1:] {
		if colIdx < len(row) && matchCondition(row[colIdx], op, value) {
			result = append(result, row)
		}
	}

	return result, nil
}

// GroupBy groups rows by a column and applies aggregation.
// GroupBy groups rows by a column and applies aggregation using Excel formulas.
// Result is written to a new sheet with formulas that auto-update.
// groupCol and aggCol can be column letters (A, B) or column names (部门, 工资).
func GroupBy(f *excelize.File, sheet, groupCol, aggCol, aggFunc string) (string, error) {
	rows, err := f.GetRows(sheet)
	if err != nil {
		return "", fmt.Errorf("get rows: %w", err)
	}

	if len(rows) < 2 {
		return "", fmt.Errorf("sheet has no data rows")
	}

	// Resolve column references (support both letter and name)
	groupColLetter, err := resolveColumnRef(f, sheet, groupCol)
	if err != nil {
		return "", fmt.Errorf("resolve group column: %w", err)
	}

	aggColLetter, err := resolveColumnRef(f, sheet, aggCol)
	if err != nil {
		return "", fmt.Errorf("resolve agg column: %w", err)
	}

	groupIdx, _ := excelize.ColumnNameToNumber(groupColLetter)

	// Get actual column names from header row
	groupColName := groupColLetter
	aggColName := aggColLetter
	if groupIdx-1 < len(rows[0]) {
		groupColName = rows[0][groupIdx-1]
	}
	aggIdx, _ := excelize.ColumnNameToNumber(aggColLetter)
	if aggIdx-1 < len(rows[0]) {
		aggColName = rows[0][aggIdx-1]
	}

	// Get unique group values
	uniqueGroups := getUniqueValues(rows[1:], groupIdx-1)

	// Create result sheet
	resultSheet := sheet + "_groupby"
	if _, err := f.NewSheet(resultSheet); err != nil {
		return "", fmt.Errorf("create result sheet: %w", err)
	}

	// Write headers
	if err := f.SetCellValue(resultSheet, "A1", groupColName); err != nil {
		return "", fmt.Errorf("set header: %w", err)
	}
	if err := f.SetCellValue(resultSheet, "B1", fmt.Sprintf("%s(%s)", aggFunc, aggColName)); err != nil {
		return "", fmt.Errorf("set header: %w", err)
	}

	// Write group values and formulas
	for i, group := range uniqueGroups {
		row := i + 2
		if err := f.SetCellValue(resultSheet, fmt.Sprintf("A%d", row), group); err != nil {
			return "", fmt.Errorf("set group value: %w", err)
		}

		// Build Excel formula (SUMIFS, COUNTIFS, etc.)
		formula := buildGroupByFormula(aggFunc, sheet, groupColLetter, aggColLetter, group, rows)
		if err := f.SetCellFormula(resultSheet, fmt.Sprintf("B%d", row), formula); err != nil {
			return "", fmt.Errorf("set formula: %w", err)
		}
	}

	return resultSheet, nil
}

// resolveColumnRef resolves a column reference (letter or name) to a column letter.
// If the input is already a column letter (A, B, AA, etc.), it returns as-is.
// If the input is a column name (部门, 工资, etc.), it finds the corresponding letter.
func resolveColumnRef(f *excelize.File, sheet, colRef string) (string, error) {
	// Check if it's already a column letter (1-3 characters, all uppercase ASCII)
	if len(colRef) >= 1 && len(colRef) <= 3 {
		allUpper := true
		for _, c := range colRef {
			if c < 'A' || c > 'Z' {
				allUpper = false
				break
			}
		}
		if allUpper {
			_, err := excelize.ColumnNameToNumber(colRef)
			if err == nil {
				return colRef, nil
			}
		}
	}

	// Try to find column by name in header row
	rows, err := f.GetRows(sheet)
	if err != nil {
		return "", fmt.Errorf("get rows: %w", err)
	}

	if len(rows) == 0 {
		return "", fmt.Errorf("sheet has no data")
	}

	for i, cell := range rows[0] {
		if cell == colRef {
			colLetter, _ := excelize.ColumnNumberToName(i + 1)
			return colLetter, nil
		}
	}

	return "", fmt.Errorf("column '%s' not found in header row", colRef)
}

// resolveColumnIndex resolves a column reference (letter or name) to a 0-based column index.
func resolveColumnIndex(f *excelize.File, sheet, colRef string) (int, error) {
	// Try as column letter first
	if len(colRef) >= 1 && len(colRef) <= 3 {
		allUpper := true
		for _, c := range colRef {
			if c < 'A' || c > 'Z' {
				allUpper = false
				break
			}
		}
		if allUpper {
			idx, err := excelize.ColumnNameToNumber(colRef)
			if err == nil {
				return idx - 1, nil
			}
		}
	}

	// Try to find column by name in header row
	rows, err := f.GetRows(sheet)
	if err != nil {
		return -1, fmt.Errorf("get rows: %w", err)
	}

	if len(rows) == 0 {
		return -1, fmt.Errorf("sheet has no data")
	}

	for i, cell := range rows[0] {
		if cell == colRef {
			return i, nil
		}
	}

	return -1, fmt.Errorf("column '%s' not found in header row", colRef)
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
	groupRange := fmt.Sprintf("%s2:%s%d", groupCol, groupCol, len(rows))
	aggRange := fmt.Sprintf("%s2:%s%d", aggCol, aggCol, len(rows))

	// Build criteria - need to handle the group value
	criteria := fmt.Sprintf("\"%s\"", groupValue)

	switch strings.ToLower(aggFunc) {
	case "sum":
		return fmt.Sprintf("=SUMIFS(%s!%s,%s!%s,%s)",
			sheet, aggRange, sheet, groupRange, criteria)
	case "count":
		return fmt.Sprintf("=COUNTIF(%s!%s,%s)",
			sheet, groupRange, criteria)
	case "avg", "average":
		return fmt.Sprintf("=AVERAGEIF(%s!%s,%s,%s!%s)",
			sheet, groupRange, criteria, sheet, aggRange)
	case "min":
		return fmt.Sprintf("=MINIFS(%s!%s,%s!%s,%s)",
			sheet, aggRange, sheet, groupRange, criteria)
	case "max":
		return fmt.Sprintf("=MAXIFS(%s!%s,%s!%s,%s)",
			sheet, aggRange, sheet, groupRange, criteria)
	default:
		return fmt.Sprintf("=SUMIFS(%s!%s,%s!%s,%s)",
			sheet, aggRange, sheet, groupRange, criteria)
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
		if _, err := f.NewSheet(newSheet); err != nil {
			return nil, fmt.Errorf("create sheet %s: %w", newSheet, err)
		}

		// Write header
		for j, cell := range rows[0] {
			colName, _ := excelize.ColumnNumberToName(j + 1)
			if err := f.SetCellValue(newSheet, fmt.Sprintf("%s1", colName), cell); err != nil {
				return nil, fmt.Errorf("set header: %w", err)
			}
		}

		// Write data rows
		for i, row := range groupRows {
			for j, cell := range row {
				colName, _ := excelize.ColumnNumberToName(j + 1)
				if err := f.SetCellValue(newSheet, fmt.Sprintf("%s%d", colName, i+2), cell); err != nil {
					return nil, fmt.Errorf("set cell: %w", err)
				}
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
	if _, err := f.NewSheet(destSheet); err != nil {
		return fmt.Errorf("create destination sheet: %w", err)
	}

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

		if !headerWritten {
			// Write header from first sheet
			for j, cell := range rows[0] {
				colName, _ := excelize.ColumnNumberToName(j + 1)
				if err := f.SetCellValue(destSheet, fmt.Sprintf("%s%d", colName, destRow), cell); err != nil {
					return fmt.Errorf("set header: %w", err)
				}
			}
			headerWritten = true
			destRow++
		}

		// Write data rows (skip header for all sheets)
		for i := 1; i < len(rows); i++ {
			for j, cell := range rows[i] {
				colName, _ := excelize.ColumnNumberToName(j + 1)
				if err := f.SetCellValue(destSheet, fmt.Sprintf("%s%d", colName, destRow), cell); err != nil {
					return fmt.Errorf("set cell: %w", err)
				}
			}
			destRow++
		}
	}

	return nil
}
