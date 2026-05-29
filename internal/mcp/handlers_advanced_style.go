package mcp

import (
	"context"
	"log/slog"

	"github.com/mark3labs/mcp-go/mcp"
	stylePkg "github.com/yuhua2000/sheetpilot/internal/style"
)

func (s *Server) registerAdvancedStyleTools() {
	s.mcpSrv.AddTool(
		mcp.NewTool("set_style",
			mcp.WithDescription("Set cell style (font, background, border, alignment)"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
			mcp.WithString("sheet", mcp.Required(), mcp.Description("Sheet name")),
			mcp.WithString("cell", mcp.Required(), mcp.Description("Cell reference (e.g., A1)")),
			mcp.WithString("bold", mcp.Description("Bold: true/false")),
			mcp.WithString("italic", mcp.Description("Italic: true/false")),
			mcp.WithString("font_size", mcp.Description("Font size (e.g., 12)")),
			mcp.WithString("bg_color", mcp.Description("Background color (e.g., #FF0000)")),
			mcp.WithString("font_color", mcp.Description("Font color (e.g., #FFFFFF)")),
			mcp.WithString("align", mcp.Description("Alignment: left, center, right")),
			mcp.WithString("border", mcp.Description("Border: thin, medium, thick")),
		),
		s.handleSetStyle,
	)

	s.mcpSrv.AddTool(
		mcp.NewTool("conditional_formatting",
			mcp.WithDescription("Add conditional formatting (e.g., highlight cells)"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
			mcp.WithString("sheet", mcp.Required(), mcp.Description("Sheet name")),
			mcp.WithString("range", mcp.Required(), mcp.Description("Cell range (e.g., A1:A10)")),
			mcp.WithString("operator", mcp.Required(), mcp.Description("Operator: greater_than, less_than, equal, between")),
			mcp.WithString("value", mcp.Required(), mcp.Description("Threshold value")),
			mcp.WithString("bg_color", mcp.Required(), mcp.Description("Background color when condition met (e.g., #FF0000)")),
			mcp.WithString("font_color", mcp.Description("Font color when condition met")),
		),
		s.handleConditionalFormatting,
	)

	s.mcpSrv.AddTool(
		mcp.NewTool("freeze_panes",
			mcp.WithDescription("Freeze panes (header row and/or first column)"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
			mcp.WithString("sheet", mcp.Required(), mcp.Description("Sheet name")),
			mcp.WithString("row", mcp.Required(), mcp.Description("Freeze below this row (e.g., 1 for header)")),
			mcp.WithString("col", mcp.Required(), mcp.Description("Freeze right of this column (e.g., 0 for none)")),
		),
		s.handleFreezePanes,
	)

	s.mcpSrv.AddTool(
		mcp.NewTool("add_filter",
			mcp.WithDescription("Add auto filter to header row"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
			mcp.WithString("sheet", mcp.Required(), mcp.Description("Sheet name")),
			mcp.WithString("range", mcp.Required(), mcp.Description("Data range including header (e.g., A1:D100)")),
		),
		s.handleAddFilter,
	)

	s.mcpSrv.AddTool(
		mcp.NewTool("format_as_table",
			mcp.WithDescription("Format range as a table with header styling"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
			mcp.WithString("sheet", mcp.Required(), mcp.Description("Sheet name")),
			mcp.WithString("range", mcp.Required(), mcp.Description("Table range (e.g., A1:D100)")),
			mcp.WithString("header_bg", mcp.Description("Header background color (default: #4472C4)")),
			mcp.WithString("stripe_rows", mcp.Description("Stripe rows: true/false (default: true)")),
		),
		s.handleFormatAsTable,
	)
}

func (s *Server) handleSetStyle(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("workbook_id")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	sheet, err := req.RequireString("sheet")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	cell, err := req.RequireString("cell")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	slog.Debug("setting style", "workbook_id", id, "sheet", sheet, "cell", cell)
	wb, err := s.manager.Get(id)
	if err != nil {
		return mcpError(err.Error()), nil
	}

	opts := stylePkg.StyleOptions{
		Bold:      req.GetString("bold", ""),
		Italic:    req.GetString("italic", ""),
		FontSize:  req.GetString("font_size", ""),
		BgColor:   req.GetString("bg_color", ""),
		FontColor: req.GetString("font_color", ""),
		Align:     req.GetString("align", ""),
		Border:    req.GetString("border", ""),
	}

	if err := stylePkg.SetStyle(wb.File, sheet, cell, opts); err != nil {
		slog.Error("set style failed", "workbook_id", id, "sheet", sheet, "cell", cell, "error", err)
		return mcpError(err.Error()), nil
	}

	return mcpResult("ok"), nil
}

func (s *Server) handleConditionalFormatting(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("workbook_id")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	sheet, err := req.RequireString("sheet")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	rangeRef, err := req.RequireString("range")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	operator, err := req.RequireString("operator")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	value, err := req.RequireString("value")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	bgColor, err := req.RequireString("bg_color")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	fontColor := req.GetString("font_color", "")

	slog.Debug("setting conditional formatting", "workbook_id", id, "sheet", sheet, "range", rangeRef)
	wb, err := s.manager.Get(id)
	if err != nil {
		return mcpError(err.Error()), nil
	}

	if err := stylePkg.SetConditionalFormatting(wb.File, sheet, rangeRef, operator, value, bgColor, fontColor); err != nil {
		slog.Error("conditional formatting failed", "workbook_id", id, "sheet", sheet, "error", err)
		return mcpError(err.Error()), nil
	}

	return mcpResult("ok"), nil
}

func (s *Server) handleFreezePanes(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("workbook_id")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	sheet, err := req.RequireString("sheet")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	row, err := req.RequireString("row")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	col, err := req.RequireString("col")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	slog.Debug("freezing panes", "workbook_id", id, "sheet", sheet, "row", row, "col", col)
	wb, err := s.manager.Get(id)
	if err != nil {
		return mcpError(err.Error()), nil
	}

	if err := stylePkg.FreezePanes(wb.File, sheet, row, col); err != nil {
		slog.Error("freeze panes failed", "workbook_id", id, "sheet", sheet, "error", err)
		return mcpError(err.Error()), nil
	}

	return mcpResult("ok"), nil
}

func (s *Server) handleAddFilter(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("workbook_id")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	sheet, err := req.RequireString("sheet")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	rangeRef, err := req.RequireString("range")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	slog.Debug("adding filter", "workbook_id", id, "sheet", sheet, "range", rangeRef)
	wb, err := s.manager.Get(id)
	if err != nil {
		return mcpError(err.Error()), nil
	}

	if err := stylePkg.AddFilter(wb.File, sheet, rangeRef); err != nil {
		slog.Error("add filter failed", "workbook_id", id, "sheet", sheet, "error", err)
		return mcpError(err.Error()), nil
	}

	return mcpResult("ok"), nil
}

func (s *Server) handleFormatAsTable(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("workbook_id")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	sheet, err := req.RequireString("sheet")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	rangeRef, err := req.RequireString("range")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	headerBg := req.GetString("header_bg", "#4472C4")
	stripeRows := req.GetString("stripe_rows", "true") == "true"

	slog.Debug("formatting as table", "workbook_id", id, "sheet", sheet, "range", rangeRef)
	wb, err := s.manager.Get(id)
	if err != nil {
		return mcpError(err.Error()), nil
	}

	if err := stylePkg.FormatAsTable(wb.File, sheet, rangeRef, headerBg, stripeRows); err != nil {
		slog.Error("format as table failed", "workbook_id", id, "sheet", sheet, "error", err)
		return mcpError(err.Error()), nil
	}

	return mcpResult("ok"), nil
}
