package worksheet

import (
	"fmt"

	"github.com/xuri/excelize/v2"
)

// SheetInfo contains metadata about a sheet.
type SheetInfo struct {
	Name     string
	Index    int
	RowCount int
	ColCount int
}

// ListSheets returns all sheet names in the workbook.
func ListSheets(f *excelize.File) []string {
	return f.GetSheetList()
}

// CreateSheet creates a new sheet with the given name.
func CreateSheet(f *excelize.File, name string) (int, error) {
	idx, err := f.NewSheet(name)
	if err != nil {
		return 0, fmt.Errorf("create sheet: %w", err)
	}
	return idx, nil
}

// DeleteSheet deletes a sheet by name.
func DeleteSheet(f *excelize.File, name string) error {
	if err := f.DeleteSheet(name); err != nil {
		return fmt.Errorf("delete sheet: %w", err)
	}
	return nil
}

// RenameSheet renames a sheet.
func RenameSheet(f *excelize.File, oldName, newName string) error {
	if err := f.SetSheetName(oldName, newName); err != nil {
		return fmt.Errorf("rename sheet: %w", err)
	}
	return nil
}

// CopySheet copies a sheet and returns the new sheet name.
func CopySheet(f *excelize.File, src string) (string, error) {
	idx, err := f.GetSheetIndex(src)
	if err != nil {
		return "", fmt.Errorf("get sheet index: %w", err)
	}
	// CopySheet in excelize copies within the same workbook
	// We need to create a new sheet and copy content
	newSheet := src + " (Copy)"
	newIdx, err := f.NewSheet(newSheet)
	if err != nil {
		return "", fmt.Errorf("create new sheet: %w", err)
	}
	if err := f.CopySheet(idx, newIdx); err != nil {
		return "", fmt.Errorf("copy sheet: %w", err)
	}
	return newSheet, nil
}

// SetActiveSheet sets the active sheet.
func SetActiveSheet(f *excelize.File, name string) error {
	idx, err := f.GetSheetIndex(name)
	if err != nil {
		return fmt.Errorf("get sheet index: %w", err)
	}
	if idx == -1 {
		return fmt.Errorf("sheet %s not found", name)
	}
	f.SetActiveSheet(idx)
	return nil
}

// GetSheetInfo returns metadata about a sheet.
func GetSheetInfo(f *excelize.File, name string) (*SheetInfo, error) {
	idx, err := f.GetSheetIndex(name)
	if err != nil {
		return nil, fmt.Errorf("get sheet index: %w", err)
	}
	if idx == -1 {
		return nil, fmt.Errorf("sheet %s not found", name)
	}

	rows, err := f.GetRows(name)
	if err != nil {
		return nil, fmt.Errorf("get rows: %w", err)
	}

	rowCount := len(rows)
	colCount := 0
	if rowCount > 0 {
		colCount = len(rows[0])
	}

	return &SheetInfo{
		Name:     name,
		Index:    idx,
		RowCount: rowCount,
		ColCount: colCount,
	}, nil
}
