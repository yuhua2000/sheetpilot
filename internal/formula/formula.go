package formula

import (
	"fmt"
	"strings"

	"github.com/xuri/excelize/v2"
)

// GetFormula returns the formula in a cell.
func GetFormula(f *excelize.File, sheet, cell string) (string, error) {
	formula, err := f.GetCellFormula(sheet, cell)
	if err != nil {
		return "", fmt.Errorf("get formula: %w", err)
	}
	return formula, nil
}

// SetFormula sets a formula in a cell.
func SetFormula(f *excelize.File, sheet, cell, formula string) error {
	if err := f.SetCellFormula(sheet, cell, formula); err != nil {
		return fmt.Errorf("set formula: %w", err)
	}
	return nil
}

// FillFormulaColumn fills a formula down a column.
func FillFormulaColumn(f *excelize.File, sheet, col string, startRow, endRow int, formulaTmpl string) error {
	for row := startRow; row <= endRow; row++ {
		cell := fmt.Sprintf("%s%d", col, row)
		formula := replaceRowPlaceholder(formulaTmpl, row)
		if err := f.SetCellFormula(sheet, cell, formula); err != nil {
			return fmt.Errorf("set formula at %s: %w", cell, err)
		}
	}
	return nil
}

// replaceRowPlaceholder replaces {row} placeholder in formula template.
func replaceRowPlaceholder(tmpl string, row int) string {
	return strings.ReplaceAll(tmpl, "{row}", fmt.Sprintf("%d", row))
}
