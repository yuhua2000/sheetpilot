package rangeop

import (
	"fmt"
	"strings"

	"github.com/xuri/excelize/v2"
)

// GetCell reads a single cell value.
func GetCell(f *excelize.File, sheet, cell string) (string, error) {
	val, err := f.GetCellValue(sheet, cell)
	if err != nil {
		return "", fmt.Errorf("get cell %s: %w", cell, err)
	}
	return val, nil
}

// SetCell writes a value to a single cell.
func SetCell(f *excelize.File, sheet, cell string, value any) error {
	if err := f.SetCellValue(sheet, cell, value); err != nil {
		return fmt.Errorf("set cell %s: %w", cell, err)
	}
	return nil
}

// ClearCell clears a single cell.
func ClearCell(f *excelize.File, sheet, cell string) error {
	if err := f.SetCellValue(sheet, cell, nil); err != nil {
		return fmt.Errorf("clear cell %s: %w", cell, err)
	}
	return nil
}

// GetRange reads a range of cells and returns a 2D array.
func GetRange(f *excelize.File, sheet, rangeRef string) ([][]string, error) {
	rows, err := f.GetRows(sheet)
	if err != nil {
		return nil, fmt.Errorf("get rows: %w", err)
	}

	startCol, startRow, endCol, endRow, err := parseRange(rangeRef)
	if err != nil {
		return nil, fmt.Errorf("parse range: %w", err)
	}

	var result [][]string
	for r := startRow - 1; r < endRow && r < len(rows); r++ {
		var row []string
		for c := startCol - 1; c < endCol; c++ {
			if c < len(rows[r]) {
				row = append(row, rows[r][c])
			} else {
				row = append(row, "")
			}
		}
		result = append(result, row)
	}

	return result, nil
}

// SetRange writes a 2D array to a range starting at the given cell.
func SetRange(f *excelize.File, sheet, startCell string, data [][]any) error {
	startCol, startRow, err := excelize.CellNameToCoordinates(startCell)
	if err != nil {
		return fmt.Errorf("parse start cell: %w", err)
	}

	for r, row := range data {
		for c, val := range row {
			cell, err := excelize.CoordinatesToCellName(startCol+c, startRow+r)
			if err != nil {
				return fmt.Errorf("coordinates to cell name: %w", err)
			}
			if err := f.SetCellValue(sheet, cell, val); err != nil {
				return fmt.Errorf("set cell %s: %w", cell, err)
			}
		}
	}

	return nil
}

// AppendRows appends rows to the end of the sheet.
func AppendRows(f *excelize.File, sheet string, data [][]any) error {
	rows, err := f.GetRows(sheet)
	if err != nil {
		return fmt.Errorf("get rows: %w", err)
	}

	startRow := len(rows) + 1
	startCell := fmt.Sprintf("A%d", startRow)

	return SetRange(f, sheet, startCell, data)
}

// InsertRows inserts empty rows at the given position.
func InsertRows(f *excelize.File, sheet string, row, count int) error {
	if err := f.InsertRows(sheet, row, count); err != nil {
		return fmt.Errorf("insert rows: %w", err)
	}
	return nil
}

// DeleteRows deletes rows at the given position.
func DeleteRows(f *excelize.File, sheet string, row, count int) error {
	if err := f.RemoveRow(sheet, row); err != nil {
		return fmt.Errorf("delete rows: %w", err)
	}
	return nil
}

// InsertCols inserts empty columns at the given position.
func InsertCols(f *excelize.File, sheet, col string, count int) error {
	if err := f.InsertCols(sheet, col, count); err != nil {
		return fmt.Errorf("insert cols: %w", err)
	}
	return nil
}

// DeleteCols deletes columns at the given position.
func DeleteCols(f *excelize.File, sheet, col string, count int) error {
	if err := f.RemoveCol(sheet, col); err != nil {
		return fmt.Errorf("delete cols: %w", err)
	}
	return nil
}

// CopyRange copies a range to a destination cell.
func CopyRange(f *excelize.File, sheet, srcRange, dstCell string) error {
	data, err := GetRange(f, sheet, srcRange)
	if err != nil {
		return fmt.Errorf("get source range: %w", err)
	}

	// Convert string data to any
	anyData := make([][]any, len(data))
	for i, row := range data {
		anyData[i] = make([]any, len(row))
		for j, cell := range row {
			anyData[i][j] = cell
		}
	}

	return SetRange(f, sheet, dstCell, anyData)
}

// MoveRange moves a range to a destination cell (copy then clear source).
func MoveRange(f *excelize.File, sheet, srcRange, dstCell string) error {
	if err := CopyRange(f, sheet, srcRange, dstCell); err != nil {
		return fmt.Errorf("copy range: %w", err)
	}

	// Clear source range
	startCol, startRow, endCol, endRow, err := parseRange(srcRange)
	if err != nil {
		return fmt.Errorf("parse source range: %w", err)
	}

	for r := startRow; r <= endRow; r++ {
		for c := startCol; c <= endCol; c++ {
			cell, _ := excelize.CoordinatesToCellName(c, r)
			f.SetCellValue(sheet, cell, nil)
		}
	}

	return nil
}

// FindReplace finds and replaces text in a range.
func FindReplace(f *excelize.File, sheet, find, replace, rangeRef string) (int, error) {
	rows, err := f.GetRows(sheet)
	if err != nil {
		return 0, fmt.Errorf("get rows: %w", err)
	}

	count := 0
	for i, row := range rows {
		for j, cell := range row {
			if strings.Contains(cell, find) {
				newVal := strings.ReplaceAll(cell, find, replace)
				col, _ := excelize.ColumnNumberToName(j + 1)
				f.SetCellValue(sheet, fmt.Sprintf("%s%d", col, i+1), newVal)
				count++
			}
		}
	}

	return count, nil
}

// AddComment adds a comment to a cell.
func AddComment(f *excelize.File, sheet, cell, author, text string) error {
	return f.AddComment(sheet, excelize.Comment{
		Cell: cell,
		Paragraph: []excelize.RichTextRun{
			{
				Text: text,
				Font: &excelize.Font{
					Bold: true,
				},
			},
		},
	})
}

// AddHyperlink adds a hyperlink to a cell.
func AddHyperlink(f *excelize.File, sheet, cell, link, display string) error {
	return f.SetCellHyperLink(sheet, cell, link, "External")
}

// SetDataValidation sets data validation for a cell range with a dropdown list.
func SetDataValidation(f *excelize.File, sheet, rangeRef string, options []string) error {
	dv := excelize.DataValidation{}
	dv.SetSqref(rangeRef)
	if err := dv.SetDropList(options); err != nil {
		return fmt.Errorf("set drop list: %w", err)
	}
	return f.AddDataValidation(sheet, &dv)
}

// parseRange parses a range string like "A1:C5" into coordinates.
func parseRange(rangeRef string) (int, int, int, int, error) {
	parts := strings.Split(rangeRef, ":")
	if len(parts) != 2 {
		return 0, 0, 0, 0, fmt.Errorf("invalid range format: %s", rangeRef)
	}

	startCol, startRow, err := excelize.CellNameToCoordinates(parts[0])
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("parse start cell: %w", err)
	}

	endCol, endRow, err := excelize.CellNameToCoordinates(parts[1])
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("parse end cell: %w", err)
	}

	return startCol, startRow, endCol, endRow, nil
}
