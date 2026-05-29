package view

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

// HideSheet hides a worksheet.
func HideSheet(f *excelize.File, sheet string) error {
	return f.SetSheetVisible(sheet, false)
}

// ShowSheet shows a hidden worksheet.
func ShowSheet(f *excelize.File, sheet string) error {
	return f.SetSheetVisible(sheet, true)
}

// HideRows hides rows by range (e.g., "1:5" hides rows 1-5).
func HideRows(f *excelize.File, sheet, rowRange string) error {
	start, end, err := parseRange(rowRange)
	if err != nil {
		return fmt.Errorf("parse row range: %w", err)
	}

	for row := start; row <= end; row++ {
		if err := f.SetRowVisible(sheet, row, false); err != nil {
			return fmt.Errorf("hide row %d: %w", row, err)
		}
	}
	return nil
}

// ShowRows shows hidden rows by range.
func ShowRows(f *excelize.File, sheet, rowRange string) error {
	start, end, err := parseRange(rowRange)
	if err != nil {
		return fmt.Errorf("parse row range: %w", err)
	}

	for row := start; row <= end; row++ {
		if err := f.SetRowVisible(sheet, row, true); err != nil {
			return fmt.Errorf("show row %d: %w", row, err)
		}
	}
	return nil
}

// HideColumns hides columns by range (e.g., "A:C" hides columns A-C).
func HideColumns(f *excelize.File, sheet, colRange string) error {
	startCol, endCol, err := parseColRange(colRange)
	if err != nil {
		return fmt.Errorf("parse column range: %w", err)
	}

	startNum, _ := excelize.ColumnNameToNumber(startCol)
	endNum, _ := excelize.ColumnNameToNumber(endCol)

	for i := startNum; i <= endNum; i++ {
		col, _ := excelize.ColumnNumberToName(i)
		if err := f.SetColVisible(sheet, col, false); err != nil {
			return fmt.Errorf("hide column %s: %w", col, err)
		}
	}
	return nil
}

// ShowColumns shows hidden columns by range.
func ShowColumns(f *excelize.File, sheet, colRange string) error {
	startCol, endCol, err := parseColRange(colRange)
	if err != nil {
		return fmt.Errorf("parse column range: %w", err)
	}

	startNum, _ := excelize.ColumnNameToNumber(startCol)
	endNum, _ := excelize.ColumnNameToNumber(endCol)

	for i := startNum; i <= endNum; i++ {
		col, _ := excelize.ColumnNumberToName(i)
		if err := f.SetColVisible(sheet, col, true); err != nil {
			return fmt.Errorf("show column %s: %w", col, err)
		}
	}
	return nil
}

// SetRowHeight sets the height for specified rows.
func SetRowHeight(f *excelize.File, sheet, rowRange string, height float64) error {
	start, end, err := parseRange(rowRange)
	if err != nil {
		return fmt.Errorf("parse row range: %w", err)
	}

	for row := start; row <= end; row++ {
		if err := f.SetRowHeight(sheet, row, height); err != nil {
			return fmt.Errorf("set row %d height: %w", row, err)
		}
	}
	return nil
}

// SetColWidth sets the width for specified columns.
func SetColWidth(f *excelize.File, sheet, colRange string, width float64) error {
	startCol, endCol, err := parseColRange(colRange)
	if err != nil {
		return fmt.Errorf("parse column range: %w", err)
	}

	return f.SetColWidth(sheet, startCol, endCol, width)
}

// ProtectSheet protects a worksheet with optional password.
func ProtectSheet(f *excelize.File, sheet, password string) error {
	opts := &excelize.SheetProtectionOptions{
		Password: password,
	}
	return f.ProtectSheet(sheet, opts)
}

// UnprotectSheet removes protection from a worksheet.
func UnprotectSheet(f *excelize.File, sheet, password string) error {
	return f.UnprotectSheet(sheet, password)
}

// SetPrintArea sets the print area for a worksheet.
func SetPrintArea(f *excelize.File, sheet, area string) error {
	// SetPrintArea in excelize is done through DefinedName
	definedName := &excelize.DefinedName{
		Name:     "_xlnm.Print_Area",
		RefersTo: fmt.Sprintf("'%s'!%s", sheet, area),
		Scope:    sheet,
	}
	return f.SetDefinedName(definedName)
}

// SetHeaderFooter sets header and footer for a worksheet.
func SetHeaderFooter(f *excelize.File, sheet string, opts *excelize.HeaderFooterOptions) error {
	return f.SetHeaderFooter(sheet, opts)
}

// SetDefinedName creates a named range.
func SetDefinedName(f *excelize.File, name, refersTo, scope string) error {
	return f.SetDefinedName(&excelize.DefinedName{
		Name:     name,
		RefersTo: refersTo,
		Scope:    scope,
	})
}

// parseRange parses a range like "1:5" into start and end integers.
func parseRange(s string) (int, int, error) {
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid range format: %s", s)
	}

	start, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid start: %w", err)
	}

	end, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid end: %w", err)
	}

	return start, end, nil
}

// parseColRange parses a column range like "A:C" into start and end columns.
func parseColRange(s string) (string, string, error) {
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid column range format: %s", s)
	}

	startCol := strings.TrimSpace(parts[0])
	endCol := strings.TrimSpace(parts[1])

	// Validate columns
	_, err := excelize.ColumnNameToNumber(startCol)
	if err != nil {
		return "", "", fmt.Errorf("invalid start column: %w", err)
	}

	_, err = excelize.ColumnNameToNumber(endCol)
	if err != nil {
		return "", "", fmt.Errorf("invalid end column: %w", err)
	}

	return startCol, endCol, nil
}
