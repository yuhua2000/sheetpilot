package mcp

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/xuri/excelize/v2"
	"github.com/yuhua2000/sheetpilot/internal/view"
)

func (s *Server) registerViewTools() {
	// Sheet visibility
	s.mcpSrv.AddTool(
		mcp.NewTool("hide_sheet",
			mcp.WithDescription("Hide a worksheet"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
			mcp.WithString("sheet", mcp.Required(), mcp.Description("Sheet name")),
		),
		s.handleHideSheet,
	)

	s.mcpSrv.AddTool(
		mcp.NewTool("show_sheet",
			mcp.WithDescription("Show a hidden worksheet"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
			mcp.WithString("sheet", mcp.Required(), mcp.Description("Sheet name")),
		),
		s.handleShowSheet,
	)

	// Row/Column visibility
	s.mcpSrv.AddTool(
		mcp.NewTool("hide_rows",
			mcp.WithDescription("Hide rows (e.g., '1:5' hides rows 1-5)"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
			mcp.WithString("sheet", mcp.Required(), mcp.Description("Sheet name")),
			mcp.WithString("range", mcp.Required(), mcp.Description("Row range (e.g., 1:5)")),
		),
		s.handleHideRows,
	)

	s.mcpSrv.AddTool(
		mcp.NewTool("show_rows",
			mcp.WithDescription("Show hidden rows"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
			mcp.WithString("sheet", mcp.Required(), mcp.Description("Sheet name")),
			mcp.WithString("range", mcp.Required(), mcp.Description("Row range (e.g., 1:5)")),
		),
		s.handleShowRows,
	)

	s.mcpSrv.AddTool(
		mcp.NewTool("hide_columns",
			mcp.WithDescription("Hide columns (e.g., 'A:C' hides columns A-C)"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
			mcp.WithString("sheet", mcp.Required(), mcp.Description("Sheet name")),
			mcp.WithString("range", mcp.Required(), mcp.Description("Column range (e.g., A:C)")),
		),
		s.handleHideColumns,
	)

	s.mcpSrv.AddTool(
		mcp.NewTool("show_columns",
			mcp.WithDescription("Show hidden columns"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
			mcp.WithString("sheet", mcp.Required(), mcp.Description("Sheet name")),
			mcp.WithString("range", mcp.Required(), mcp.Description("Column range (e.g., A:C)")),
		),
		s.handleShowColumns,
	)

	// Row/Column size
	s.mcpSrv.AddTool(
		mcp.NewTool("set_row_height",
			mcp.WithDescription("Set row height"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
			mcp.WithString("sheet", mcp.Required(), mcp.Description("Sheet name")),
			mcp.WithString("range", mcp.Required(), mcp.Description("Row range (e.g., 1:5)")),
			mcp.WithString("height", mcp.Required(), mcp.Description("Height in points")),
		),
		s.handleSetRowHeight,
	)

	s.mcpSrv.AddTool(
		mcp.NewTool("set_col_width",
			mcp.WithDescription("Set column width"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
			mcp.WithString("sheet", mcp.Required(), mcp.Description("Sheet name")),
			mcp.WithString("range", mcp.Required(), mcp.Description("Column range (e.g., A:C)")),
			mcp.WithString("width", mcp.Required(), mcp.Description("Width in characters")),
		),
		s.handleSetColWidth,
	)

	// Sheet protection
	s.mcpSrv.AddTool(
		mcp.NewTool("protect_sheet",
			mcp.WithDescription("Protect worksheet with password"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
			mcp.WithString("sheet", mcp.Required(), mcp.Description("Sheet name")),
			mcp.WithString("password", mcp.Description("Protection password (optional)")),
		),
		s.handleProtectSheet,
	)

	s.mcpSrv.AddTool(
		mcp.NewTool("unprotect_sheet",
			mcp.WithDescription("Remove worksheet protection"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
			mcp.WithString("sheet", mcp.Required(), mcp.Description("Sheet name")),
			mcp.WithString("password", mcp.Description("Protection password (if set)")),
		),
		s.handleUnprotectSheet,
	)

	// Print settings
	s.mcpSrv.AddTool(
		mcp.NewTool("set_print_area",
			mcp.WithDescription("Set print area"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
			mcp.WithString("sheet", mcp.Required(), mcp.Description("Sheet name")),
			mcp.WithString("area", mcp.Required(), mcp.Description("Print area (e.g., A1:G50)")),
		),
		s.handleSetPrintArea,
	)

	s.mcpSrv.AddTool(
		mcp.NewTool("set_header_footer",
			mcp.WithDescription("Set page header and footer"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
			mcp.WithString("sheet", mcp.Required(), mcp.Description("Sheet name")),
			mcp.WithString("header", mcp.Description("Header text")),
			mcp.WithString("footer", mcp.Description("Footer text")),
		),
		s.handleSetHeaderFooter,
	)

	// Named ranges
	s.mcpSrv.AddTool(
		mcp.NewTool("set_defined_name",
			mcp.WithDescription("Create a named range"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
			mcp.WithString("name", mcp.Required(), mcp.Description("Name for the range")),
			mcp.WithString("refers_to", mcp.Required(), mcp.Description("Cell reference (e.g., Sheet1!$A$1:$C$10)")),
			mcp.WithString("scope", mcp.Description("Scope (sheet name or empty for workbook)")),
		),
		s.handleSetDefinedName,
	)

	// Picture
	s.mcpSrv.AddTool(
		mcp.NewTool("insert_image",
			mcp.WithDescription("Insert an image into a cell"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
			mcp.WithString("sheet", mcp.Required(), mcp.Description("Sheet name")),
			mcp.WithString("cell", mcp.Required(), mcp.Description("Cell reference (e.g., A1)")),
			mcp.WithString("image_path", mcp.Required(), mcp.Description("Path to image file")),
		),
		s.handleInsertImage,
	)
}

func (s *Server) handleHideSheet(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("workbook_id")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	sheet, err := req.RequireString("sheet")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	slog.Debug("hiding sheet", "workbook_id", id, "sheet", sheet)
	wb, err := s.manager.Get(id)
	if err != nil {
		return mcpError(err.Error()), nil
	}

	if err := view.HideSheet(wb.File, sheet); err != nil {
		slog.Error("hide sheet failed", "workbook_id", id, "sheet", sheet, "error", err)
		return mcpError(err.Error()), nil
	}

	return mcpResult("ok"), nil
}

func (s *Server) handleShowSheet(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("workbook_id")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	sheet, err := req.RequireString("sheet")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	slog.Debug("showing sheet", "workbook_id", id, "sheet", sheet)
	wb, err := s.manager.Get(id)
	if err != nil {
		return mcpError(err.Error()), nil
	}

	if err := view.ShowSheet(wb.File, sheet); err != nil {
		slog.Error("show sheet failed", "workbook_id", id, "sheet", sheet, "error", err)
		return mcpError(err.Error()), nil
	}

	return mcpResult("ok"), nil
}

func (s *Server) handleHideRows(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("workbook_id")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	sheet, err := req.RequireString("sheet")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	rowRange, err := req.RequireString("range")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	slog.Debug("hiding rows", "workbook_id", id, "sheet", sheet, "range", rowRange)
	wb, err := s.manager.Get(id)
	if err != nil {
		return mcpError(err.Error()), nil
	}

	if err := view.HideRows(wb.File, sheet, rowRange); err != nil {
		slog.Error("hide rows failed", "workbook_id", id, "sheet", sheet, "error", err)
		return mcpError(err.Error()), nil
	}

	return mcpResult("ok"), nil
}

func (s *Server) handleShowRows(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("workbook_id")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	sheet, err := req.RequireString("sheet")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	rowRange, err := req.RequireString("range")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	slog.Debug("showing rows", "workbook_id", id, "sheet", sheet, "range", rowRange)
	wb, err := s.manager.Get(id)
	if err != nil {
		return mcpError(err.Error()), nil
	}

	if err := view.ShowRows(wb.File, sheet, rowRange); err != nil {
		slog.Error("show rows failed", "workbook_id", id, "sheet", sheet, "error", err)
		return mcpError(err.Error()), nil
	}

	return mcpResult("ok"), nil
}

func (s *Server) handleHideColumns(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("workbook_id")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	sheet, err := req.RequireString("sheet")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	colRange, err := req.RequireString("range")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	slog.Debug("hiding columns", "workbook_id", id, "sheet", sheet, "range", colRange)
	wb, err := s.manager.Get(id)
	if err != nil {
		return mcpError(err.Error()), nil
	}

	if err := view.HideColumns(wb.File, sheet, colRange); err != nil {
		slog.Error("hide columns failed", "workbook_id", id, "sheet", sheet, "error", err)
		return mcpError(err.Error()), nil
	}

	return mcpResult("ok"), nil
}

func (s *Server) handleShowColumns(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("workbook_id")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	sheet, err := req.RequireString("sheet")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	colRange, err := req.RequireString("range")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	slog.Debug("showing columns", "workbook_id", id, "sheet", sheet, "range", colRange)
	wb, err := s.manager.Get(id)
	if err != nil {
		return mcpError(err.Error()), nil
	}

	if err := view.ShowColumns(wb.File, sheet, colRange); err != nil {
		slog.Error("show columns failed", "workbook_id", id, "sheet", sheet, "error", err)
		return mcpError(err.Error()), nil
	}

	return mcpResult("ok"), nil
}

func (s *Server) handleSetRowHeight(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("workbook_id")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	sheet, err := req.RequireString("sheet")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	rowRange, err := req.RequireString("range")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	heightStr, err := req.RequireString("height")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	height, err := strconv.ParseFloat(heightStr, 64)
	if err != nil {
		return mcpError(fmt.Sprintf("invalid height: %v", err)), nil
	}

	slog.Debug("setting row height", "workbook_id", id, "sheet", sheet, "range", rowRange, "height", height)
	wb, err := s.manager.Get(id)
	if err != nil {
		return mcpError(err.Error()), nil
	}

	if err := view.SetRowHeight(wb.File, sheet, rowRange, height); err != nil {
		slog.Error("set row height failed", "workbook_id", id, "sheet", sheet, "error", err)
		return mcpError(err.Error()), nil
	}

	return mcpResult("ok"), nil
}

func (s *Server) handleSetColWidth(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("workbook_id")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	sheet, err := req.RequireString("sheet")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	colRange, err := req.RequireString("range")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	widthStr, err := req.RequireString("width")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	width, err := strconv.ParseFloat(widthStr, 64)
	if err != nil {
		return mcpError(fmt.Sprintf("invalid width: %v", err)), nil
	}

	slog.Debug("setting column width", "workbook_id", id, "sheet", sheet, "range", colRange, "width", width)
	wb, err := s.manager.Get(id)
	if err != nil {
		return mcpError(err.Error()), nil
	}

	if err := view.SetColWidth(wb.File, sheet, colRange, width); err != nil {
		slog.Error("set column width failed", "workbook_id", id, "sheet", sheet, "error", err)
		return mcpError(err.Error()), nil
	}

	return mcpResult("ok"), nil
}

func (s *Server) handleProtectSheet(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("workbook_id")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	sheet, err := req.RequireString("sheet")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	password := req.GetString("password", "")

	slog.Debug("protecting sheet", "workbook_id", id, "sheet", sheet)
	wb, err := s.manager.Get(id)
	if err != nil {
		return mcpError(err.Error()), nil
	}

	if err := view.ProtectSheet(wb.File, sheet, password); err != nil {
		slog.Error("protect sheet failed", "workbook_id", id, "sheet", sheet, "error", err)
		return mcpError(err.Error()), nil
	}

	return mcpResult("ok"), nil
}

func (s *Server) handleUnprotectSheet(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("workbook_id")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	sheet, err := req.RequireString("sheet")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	password := req.GetString("password", "")

	slog.Debug("unprotecting sheet", "workbook_id", id, "sheet", sheet)
	wb, err := s.manager.Get(id)
	if err != nil {
		return mcpError(err.Error()), nil
	}

	if err := view.UnprotectSheet(wb.File, sheet, password); err != nil {
		slog.Error("unprotect sheet failed", "workbook_id", id, "sheet", sheet, "error", err)
		return mcpError(err.Error()), nil
	}

	return mcpResult("ok"), nil
}

func (s *Server) handleSetPrintArea(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("workbook_id")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	sheet, err := req.RequireString("sheet")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	area, err := req.RequireString("area")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	slog.Debug("setting print area", "workbook_id", id, "sheet", sheet, "area", area)
	wb, err := s.manager.Get(id)
	if err != nil {
		return mcpError(err.Error()), nil
	}

	if err := view.SetPrintArea(wb.File, sheet, area); err != nil {
		slog.Error("set print area failed", "workbook_id", id, "sheet", sheet, "error", err)
		return mcpError(err.Error()), nil
	}

	return mcpResult("ok"), nil
}

func (s *Server) handleSetHeaderFooter(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("workbook_id")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	sheet, err := req.RequireString("sheet")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	header := req.GetString("header", "")
	footer := req.GetString("footer", "")

	slog.Debug("setting header footer", "workbook_id", id, "sheet", sheet)
	wb, err := s.manager.Get(id)
	if err != nil {
		return mcpError(err.Error()), nil
	}

	alignWithMargins := true
	opts := &excelize.HeaderFooterOptions{
		AlignWithMargins: &alignWithMargins,
	}
	if header != "" {
		opts.OddHeader = header
	}
	if footer != "" {
		opts.OddFooter = footer
	}

	if err := view.SetHeaderFooter(wb.File, sheet, opts); err != nil {
		slog.Error("set header footer failed", "workbook_id", id, "sheet", sheet, "error", err)
		return mcpError(err.Error()), nil
	}

	return mcpResult("ok"), nil
}

func (s *Server) handleSetDefinedName(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("workbook_id")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	name, err := req.RequireString("name")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	refersTo, err := req.RequireString("refers_to")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	scope := req.GetString("scope", "")

	slog.Debug("setting defined name", "workbook_id", id, "name", name, "refers_to", refersTo)
	wb, err := s.manager.Get(id)
	if err != nil {
		return mcpError(err.Error()), nil
	}

	if err := view.SetDefinedName(wb.File, name, refersTo, scope); err != nil {
		slog.Error("set defined name failed", "workbook_id", id, "name", name, "error", err)
		return mcpError(err.Error()), nil
	}

	return mcpResult("ok"), nil
}

func (s *Server) handleInsertImage(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	imagePath, err := req.RequireString("image_path")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	slog.Debug("inserting image", "workbook_id", id, "sheet", sheet, "cell", cell, "image", imagePath)
	wb, err := s.manager.Get(id)
	if err != nil {
		return mcpError(err.Error()), nil
	}

	if err := wb.File.AddPicture(sheet, cell, imagePath, nil); err != nil {
		slog.Error("insert image failed", "workbook_id", id, "sheet", sheet, "error", err)
		return mcpError(err.Error()), nil
	}

	return mcpResult("ok"), nil
}
