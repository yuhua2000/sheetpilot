package io

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	"github.com/xuri/excelize/v2"
)

// ExportCSV exports a sheet to CSV file.
func ExportCSV(f *excelize.File, sheet, outputPath string) error {
	rows, err := f.GetRows(sheet)
	if err != nil {
		return fmt.Errorf("get rows: %w", err)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer file.Close()

	// Write BOM for UTF-8
	file.Write([]byte{0xEF, 0xBB, 0xBF})

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("write row: %w", err)
		}
	}

	return nil
}

// ImportCSV imports a CSV file into a sheet.
func ImportCSV(f *excelize.File, csvPath, sheet string) error {
	file, err := os.Open(csvPath)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	// Skip BOM if present
	bom := make([]byte, 3)
	n, _ := file.Read(bom)
	if n == 3 && bom[0] == 0xEF && bom[1] == 0xBB && bom[2] == 0xBF {
		// BOM found, continue reading
	} else {
		// No BOM, seek back
		file.Seek(0, 0)
	}

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("read csv: %w", err)
	}

	// Create sheet if it doesn't exist
	sheets := f.GetSheetList()
	exists := false
	for _, s := range sheets {
		if s == sheet {
			exists = true
			break
		}
	}
	if !exists {
		f.NewSheet(sheet)
	}

	// Write data
	for i, row := range records {
		for j, cell := range row {
			colName, _ := excelize.ColumnNumberToName(j + 1)
			cellRef := fmt.Sprintf("%s%d", colName, i+1)
			f.SetCellValue(sheet, cellRef, cell)
		}
	}

	return nil
}

// ExportJSON exports a sheet to JSON format (returns as string).
func ExportJSON(f *excelize.File, sheet string) (string, error) {
	rows, err := f.GetRows(sheet)
	if err != nil {
		return "", fmt.Errorf("get rows: %w", err)
	}

	if len(rows) == 0 {
		return "[]", nil
	}

	// Use first row as headers
	headers := rows[0]
	var result strings.Builder
	result.WriteString("[\n")

	for i := 1; i < len(rows); i++ {
		result.WriteString("  {")
		for j, cell := range rows[i] {
			if j < len(headers) {
				if j > 0 {
					result.WriteString(", ")
				}
				result.WriteString(fmt.Sprintf("\"%s\": \"%s\"", headers[j], escapeJSON(cell)))
			}
		}
		result.WriteString("}")
		if i < len(rows)-1 {
			result.WriteString(",")
		}
		result.WriteString("\n")
	}

	result.WriteString("]")
	return result.String(), nil
}

// escapeJSON escapes special characters for JSON.
func escapeJSON(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\t", "\\t")
	return s
}
