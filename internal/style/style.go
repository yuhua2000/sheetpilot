package style

import (
	"fmt"

	"github.com/xuri/excelize/v2"
)

// AutoFitColumns adjusts column widths to fit content.
func AutoFitColumns(f *excelize.File, sheet string) error {
	cols, err := f.GetCols(sheet)
	if err != nil {
		return fmt.Errorf("get cols: %w", err)
	}

	for i, col := range cols {
		maxWidth := 10 // minimum width
		for _, cell := range col {
			width := len(cell) + 2
			if width > maxWidth {
				maxWidth = width
			}
		}
		if maxWidth > 50 {
			maxWidth = 50 // maximum width
		}

		colName, err := excelize.ColumnNumberToName(i + 1)
		if err != nil {
			return fmt.Errorf("column number to name: %w", err)
		}
		if err := f.SetColWidth(sheet, colName, colName, float64(maxWidth)); err != nil {
			return fmt.Errorf("set col width: %w", err)
		}
	}

	return nil
}

// AutoFitRows adjusts row heights to fit content.
func AutoFitRows(f *excelize.File, sheet string) error {
	rows, err := f.GetRows(sheet)
	if err != nil {
		return fmt.Errorf("get rows: %w", err)
	}

	for i := range rows {
		// Default row height is 15, adjust based on content
		height := 15.0
		for _, cell := range rows[i] {
			lines := countLines(cell)
			if float64(lines)*15 > height {
				height = float64(lines) * 15
			}
		}
		if err := f.SetRowHeight(sheet, i+1, height); err != nil {
			return fmt.Errorf("set row height: %w", err)
		}
	}

	return nil
}

// SetNumberFormat sets the number format for a cell.
func SetNumberFormat(f *excelize.File, sheet, cell, format string) error {
	style, err := f.NewStyle(&excelize.Style{
		NumFmt: getNumFmtID(format),
	})
	if err != nil {
		return fmt.Errorf("create style: %w", err)
	}
	if err := f.SetCellStyle(sheet, cell, cell, style); err != nil {
		return fmt.Errorf("set cell style: %w", err)
	}
	return nil
}

// getNumFmtID returns the number format ID for common formats.
func getNumFmtID(format string) int {
	formats := map[string]int{
		"General":    0,
		"0":          1,
		"0.00":       2,
		"#,##0":      3,
		"#,##0.00":   4,
		"0%":         9,
		"0.00%":      10,
		"yyyy-mm-dd": 14,
		"hh:mm:ss":   16,
		"#,##0.00_);(#,##0.00)": 4,
	}

	if id, ok := formats[format]; ok {
		return id
	}
	return 0 // General format
}

func countLines(s string) int {
	count := 1
	for _, c := range s {
		if c == '\n' {
			count++
		}
	}
	return count
}
