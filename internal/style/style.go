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

// StyleOptions contains style configuration options.
type StyleOptions struct {
	Bold      string
	Italic    string
	FontSize  string
	BgColor   string
	FontColor string
	Align     string
	Border    string
}

// SetStyle sets cell style with various options.
func SetStyle(f *excelize.File, sheet, cell string, opts StyleOptions) error {
	style := excelize.Style{}

	// Font settings
	font := &excelize.Font{}
	hasFont := false

	if opts.Bold == "true" {
		font.Bold = true
		hasFont = true
	}
	if opts.Italic == "true" {
		font.Italic = true
		hasFont = true
	}
	if opts.FontSize != "" {
		size := 12.0
		fmt.Sscanf(opts.FontSize, "%f", &size)
		font.Size = size
		hasFont = true
	}
	if opts.FontColor != "" {
		font.Color = opts.FontColor
		hasFont = true
	}
	if hasFont {
		style.Font = font
	}

	// Fill settings
	if opts.BgColor != "" {
		style.Fill = excelize.Fill{
			Type:    "pattern",
			Pattern: 1,
			Color:   []string{opts.BgColor},
		}
	}

	// Alignment settings
	if opts.Align != "" {
		hAlign := "left"
		switch opts.Align {
		case "center":
			hAlign = "center"
		case "right":
			hAlign = "right"
		}
		style.Alignment = &excelize.Alignment{
			Horizontal: hAlign,
			Vertical:   "center",
		}
	}

	// Border settings
	if opts.Border != "" {
		borderStyle := 1
		switch opts.Border {
		case "medium":
			borderStyle = 2
		case "thick":
			borderStyle = 3
		}
		style.Border = []excelize.Border{
			{Type: "left", Color: "000000", Style: borderStyle},
			{Type: "right", Color: "000000", Style: borderStyle},
			{Type: "top", Color: "000000", Style: borderStyle},
			{Type: "bottom", Color: "000000", Style: borderStyle},
		}
	}

	styleID, err := f.NewStyle(&style)
	if err != nil {
		return fmt.Errorf("create style: %w", err)
	}
	return f.SetCellStyle(sheet, cell, cell, styleID)
}

// SetConditionalFormatting adds conditional formatting to a range.
func SetConditionalFormatting(f *excelize.File, sheet, rangeRef, operator, value, bgColor, fontColor string) error {
	// Create style for conditional formatting
	style := excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Pattern: 1,
			Color:   []string{bgColor},
		},
	}
	if fontColor != "" {
		style.Font = &excelize.Font{
			Color: fontColor,
		}
	}

	styleID, err := f.NewStyle(&style)
	if err != nil {
		return fmt.Errorf("create style: %w", err)
	}

	format := excelize.ConditionalFormatOptions{
		Type:     "cell",
		Criteria: mapOperator(operator),
		Value:    value,
		Format:   &styleID,
	}

	return f.SetConditionalFormat(sheet, rangeRef, []excelize.ConditionalFormatOptions{format})
}

// mapOperator maps string operator to excelize operator.
func mapOperator(op string) string {
	operators := map[string]string{
		"greater_than": ">",
		"less_than":    "<",
		"equal":        "=",
		"between":      "between",
	}
	if v, ok := operators[op]; ok {
		return v
	}
	return op
}

// FreezePanes freezes panes in a sheet.
func FreezePanes(f *excelize.File, sheet, row, col string) error {
	return f.SetPanes(sheet, &excelize.Panes{
		Freeze:      true,
		Split:       false,
		XSplit:      parseInt(col),
		YSplit:      parseInt(row),
		TopLeftCell: fmt.Sprintf("%s%d", nextCol(col), parseInt(row)+1),
		ActivePane:  "bottomRight",
	})
}

// AddFilter adds auto filter to a range.
func AddFilter(f *excelize.File, sheet, rangeRef string) error {
	return f.AutoFilter(sheet, rangeRef, nil)
}

// FormatAsTable formats a range as a table with styling.
func FormatAsTable(f *excelize.File, sheet, rangeRef, headerBg string, stripeRows bool) error {
	// Create header style
	headerStyle := excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Pattern: 1,
			Color:   []string{headerBg},
		},
		Font: &excelize.Font{
			Bold:  true,
			Color: "FFFFFF",
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
	}

	headerStyleID, err := f.NewStyle(&headerStyle)
	if err != nil {
		return fmt.Errorf("create header style: %w", err)
	}

	// Parse range to get header row
	startCell, endCell, err := parseRangeRef(rangeRef)
	if err != nil {
		return fmt.Errorf("parse range: %w", err)
	}

	// Apply header style
	headerRange := fmt.Sprintf("%s:%s", startCell, endCell[:1]+"1")
	f.SetCellStyle(sheet, headerRange, headerRange, headerStyleID)

	// Create data style with borders
	dataStyle := excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Alignment: &excelize.Alignment{
			Vertical: "center",
		},
	}

	dataStyleID, err := f.NewStyle(&dataStyle)
	if err != nil {
		return fmt.Errorf("create data style: %w", err)
	}

	// Apply data style to entire range
	f.SetCellStyle(sheet, rangeRef, rangeRef, dataStyleID)

	// Add stripe rows if enabled
	if stripeRows {
		stripeStyle := excelize.Style{
			Fill: excelize.Fill{
				Type:    "pattern",
				Pattern: 1,
				Color:   []string{"F2F2F2"},
			},
			Border: dataStyle.Border,
			Alignment: dataStyle.Alignment,
		}
		stripeStyleID, err := f.NewStyle(&stripeStyle)
		if err == nil {
			// Apply stripe to alternating rows
			rows, _ := f.GetRows(sheet)
			for i := 2; i < len(rows); i += 2 {
				rowRange := fmt.Sprintf("%s%d:%s%d", startCell, i, endCell[:1], i)
				f.SetCellStyle(sheet, rowRange, rowRange, stripeStyleID)
			}
		}
	}

	return nil
}

// parseRangeRef parses a range reference into start and end cells.
func parseRangeRef(rangeRef string) (string, string, error) {
	parts := splitRange(rangeRef)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid range: %s", rangeRef)
	}
	return parts[0], parts[1], nil
}

// splitRange splits a range string.
func splitRange(s string) []string {
	for i, c := range s {
		if c == ':' {
			return []string{s[:i], s[i+1:]}
		}
	}
	return []string{s}
}

// parseInt parses a string to int.
func parseInt(s string) int {
	n := 0
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		}
	}
	return n
}

// nextCol returns the next column letter.
func nextCol(col string) string {
	if col == "" || col == "0" {
		return "A"
	}
	return string(rune(col[0]) + 1)
}
