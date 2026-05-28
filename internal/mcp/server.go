package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/yuhua2000/sheetpilot/internal/analysis"
	"github.com/yuhua2000/sheetpilot/internal/dataops"
	formulaPkg "github.com/yuhua2000/sheetpilot/internal/formula"
	"github.com/yuhua2000/sheetpilot/internal/rangeop"
	stylePkg "github.com/yuhua2000/sheetpilot/internal/style"
	"github.com/yuhua2000/sheetpilot/internal/workbook"
	"github.com/yuhua2000/sheetpilot/internal/worksheet"
)

// Server is the Excel MCP server.
type Server struct {
	manager *workbook.Manager
	mcpSrv  *server.MCPServer
}

// NewServer creates a new Excel MCP server.
func NewServer() (*Server, error) {
	slog.Info("creating MCP server")

	s := &Server{
		manager: workbook.NewManager(),
	}

	mcpSrv := server.NewMCPServer(
		"sheetpilot",
		"0.1.0",
		server.WithToolCapabilities(true),
	)

	s.mcpSrv = mcpSrv
	s.registerTools()

	slog.Info("MCP server created", "tools", 41)
	return s, nil
}

// ServeStdio starts the server with stdio transport.
func (s *Server) ServeStdio() error {
	slog.Info("starting MCP server", "transport", "stdio")
	return server.ServeStdio(s.mcpSrv)
}

// ServeSSE starts the server with SSE transport.
func (s *Server) ServeSSE(addr string) error {
	slog.Info("starting MCP server", "transport", "sse", "addr", addr)
	return server.NewSSEServer(s.mcpSrv).Start(addr)
}

func (s *Server) registerTools() {
	slog.Debug("registering MCP tools")

	// Workbook tools
	s.mcpSrv.AddTool(
		mcp.NewTool("open_workbook",
			mcp.WithDescription("Open an existing Excel file or create a new one"),
			mcp.WithString("path", mcp.Required(), mcp.Description("Path to the Excel file")),
		),
		s.handleOpenWorkbook,
	)

	s.mcpSrv.AddTool(
		mcp.NewTool("save_workbook",
			mcp.WithDescription("Save the workbook"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
		),
		s.handleSaveWorkbook,
	)

	s.mcpSrv.AddTool(
		mcp.NewTool("save_workbook_as",
			mcp.WithDescription("Save the workbook to a new path"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
			mcp.WithString("path", mcp.Required(), mcp.Description("New file path")),
		),
		s.handleSaveWorkbookAs,
	)

	s.mcpSrv.AddTool(
		mcp.NewTool("close_workbook",
			mcp.WithDescription("Close the workbook"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
		),
		s.handleCloseWorkbook,
	)

	s.mcpSrv.AddTool(
		mcp.NewTool("list_workbooks",
			mcp.WithDescription("List all open workbooks"),
		),
		s.handleListWorkbooks,
	)

	// Sheet tools
	s.mcpSrv.AddTool(
		mcp.NewTool("list_sheets",
			mcp.WithDescription("List all sheets in the workbook"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
		),
		s.handleListSheets,
	)

	s.mcpSrv.AddTool(
		mcp.NewTool("create_sheet",
			mcp.WithDescription("Create a new sheet"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
			mcp.WithString("name", mcp.Required(), mcp.Description("Sheet name")),
		),
		s.handleCreateSheet,
	)

	s.mcpSrv.AddTool(
		mcp.NewTool("delete_sheet",
			mcp.WithDescription("Delete a sheet"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
			mcp.WithString("name", mcp.Required(), mcp.Description("Sheet name")),
		),
		s.handleDeleteSheet,
	)

	s.mcpSrv.AddTool(
		mcp.NewTool("rename_sheet",
			mcp.WithDescription("Rename a sheet"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
			mcp.WithString("old_name", mcp.Required(), mcp.Description("Current sheet name")),
			mcp.WithString("new_name", mcp.Required(), mcp.Description("New sheet name")),
		),
		s.handleRenameSheet,
	)

	s.mcpSrv.AddTool(
		mcp.NewTool("get_sheet_info",
			mcp.WithDescription("Get sheet metadata"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
			mcp.WithString("sheet", mcp.Required(), mcp.Description("Sheet name")),
		),
		s.handleGetSheetInfo,
	)

	// Cell tools
	s.mcpSrv.AddTool(
		mcp.NewTool("get_cell",
			mcp.WithDescription("Read a cell value"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
			mcp.WithString("sheet", mcp.Required(), mcp.Description("Sheet name")),
			mcp.WithString("cell", mcp.Required(), mcp.Description("Cell reference (e.g., A1)")),
		),
		s.handleGetCell,
	)

	s.mcpSrv.AddTool(
		mcp.NewTool("set_cell",
			mcp.WithDescription("Write a value to a cell"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
			mcp.WithString("sheet", mcp.Required(), mcp.Description("Sheet name")),
			mcp.WithString("cell", mcp.Required(), mcp.Description("Cell reference (e.g., A1)")),
			mcp.WithString("value", mcp.Required(), mcp.Description("Value to write")),
		),
		s.handleSetCell,
	)

	// Range tools
	s.mcpSrv.AddTool(
		mcp.NewTool("get_range",
			mcp.WithDescription("Read a range of cells"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
			mcp.WithString("sheet", mcp.Required(), mcp.Description("Sheet name")),
			mcp.WithString("range", mcp.Required(), mcp.Description("Range reference (e.g., A1:C5)")),
		),
		s.handleGetRange,
	)

	s.mcpSrv.AddTool(
		mcp.NewTool("set_range",
			mcp.WithDescription("Write data to a range"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
			mcp.WithString("sheet", mcp.Required(), mcp.Description("Sheet name")),
			mcp.WithString("start_cell", mcp.Required(), mcp.Description("Start cell (e.g., A1)")),
			mcp.WithString("data", mcp.Required(), mcp.Description("JSON array of arrays")),
		),
		s.handleSetRange,
	)

	s.mcpSrv.AddTool(
		mcp.NewTool("append_rows",
			mcp.WithDescription("Append rows to the sheet"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
			mcp.WithString("sheet", mcp.Required(), mcp.Description("Sheet name")),
			mcp.WithString("data", mcp.Required(), mcp.Description("JSON array of arrays")),
		),
		s.handleAppendRows,
	)

	// Additional workbook tools
	s.mcpSrv.AddTool(
		mcp.NewTool("get_workbook_info",
			mcp.WithDescription("Get workbook metadata"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
		),
		s.handleGetWorkbookInfo,
	)

	// Additional sheet tools
	s.mcpSrv.AddTool(
		mcp.NewTool("copy_sheet",
			mcp.WithDescription("Copy a sheet"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
			mcp.WithString("name", mcp.Required(), mcp.Description("Sheet name to copy")),
		),
		s.handleCopySheet,
	)

	s.mcpSrv.AddTool(
		mcp.NewTool("set_active_sheet",
			mcp.WithDescription("Set the active sheet"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
			mcp.WithString("name", mcp.Required(), mcp.Description("Sheet name")),
		),
		s.handleSetActiveSheet,
	)

	// Additional cell tools
	s.mcpSrv.AddTool(
		mcp.NewTool("clear_cell",
			mcp.WithDescription("Clear a cell value"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
			mcp.WithString("sheet", mcp.Required(), mcp.Description("Sheet name")),
			mcp.WithString("cell", mcp.Required(), mcp.Description("Cell reference (e.g., A1)")),
		),
		s.handleClearCell,
	)

	// Additional range tools
	s.mcpSrv.AddTool(
		mcp.NewTool("insert_rows",
			mcp.WithDescription("Insert empty rows"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
			mcp.WithString("sheet", mcp.Required(), mcp.Description("Sheet name")),
			mcp.WithString("row", mcp.Required(), mcp.Description("Row number to insert at")),
			mcp.WithString("count", mcp.Required(), mcp.Description("Number of rows to insert")),
		),
		s.handleInsertRows,
	)

	s.mcpSrv.AddTool(
		mcp.NewTool("delete_rows",
			mcp.WithDescription("Delete rows"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
			mcp.WithString("sheet", mcp.Required(), mcp.Description("Sheet name")),
			mcp.WithString("row", mcp.Required(), mcp.Description("Row number to delete from")),
			mcp.WithString("count", mcp.Required(), mcp.Description("Number of rows to delete")),
		),
		s.handleDeleteRows,
	)

	s.mcpSrv.AddTool(
		mcp.NewTool("insert_columns",
			mcp.WithDescription("Insert empty columns"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
			mcp.WithString("sheet", mcp.Required(), mcp.Description("Sheet name")),
			mcp.WithString("col", mcp.Required(), mcp.Description("Column letter to insert at (e.g., B)")),
			mcp.WithString("count", mcp.Required(), mcp.Description("Number of columns to insert")),
		),
		s.handleInsertCols,
	)

	s.mcpSrv.AddTool(
		mcp.NewTool("delete_columns",
			mcp.WithDescription("Delete columns"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
			mcp.WithString("sheet", mcp.Required(), mcp.Description("Sheet name")),
			mcp.WithString("col", mcp.Required(), mcp.Description("Column letter to delete (e.g., B)")),
			mcp.WithString("count", mcp.Required(), mcp.Description("Number of columns to delete")),
		),
		s.handleDeleteCols,
	)

	// Phase 2: Table info tools
	s.mcpSrv.AddTool(
		mcp.NewTool("table_info",
			mcp.WithDescription("Get table structure information (boundaries, headers, columns)"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
			mcp.WithString("sheet", mcp.Required(), mcp.Description("Sheet name")),
		),
		s.handleTableInfo,
	)

	s.mcpSrv.AddTool(
		mcp.NewTool("column_types",
			mcp.WithDescription("Get column types by sampling data"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
			mcp.WithString("sheet", mcp.Required(), mcp.Description("Sheet name")),
		),
		s.handleColumnTypes,
	)

	s.mcpSrv.AddTool(
		mcp.NewTool("sheet_overview",
			mcp.WithDescription("Get sheet overview (row count, col count, column types, samples)"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
			mcp.WithString("sheet", mcp.Required(), mcp.Description("Sheet name")),
		),
		s.handleSheetOverview,
	)

	// Phase 2: Data operation tools
	s.mcpSrv.AddTool(
		mcp.NewTool("add_computed_column",
			mcp.WithDescription("Add a computed column with formula"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
			mcp.WithString("sheet", mcp.Required(), mcp.Description("Sheet name")),
			mcp.WithString("col_name", mcp.Required(), mcp.Description("New column name")),
			mcp.WithString("formula", mcp.Required(), mcp.Description("Formula template, use {column_name} for column references")),
		),
		s.handleAddComputedColumn,
	)

	s.mcpSrv.AddTool(
		mcp.NewTool("sort_table",
			mcp.WithDescription("Sort table by columns"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
			mcp.WithString("sheet", mcp.Required(), mcp.Description("Sheet name")),
			mcp.WithString("columns", mcp.Required(), mcp.Description("Comma-separated column names")),
			mcp.WithString("order", mcp.Description("asc or desc (default: asc)")),
		),
		s.handleSortTable,
	)

	s.mcpSrv.AddTool(
		mcp.NewTool("filter_rows",
			mcp.WithDescription("Filter rows by condition"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
			mcp.WithString("sheet", mcp.Required(), mcp.Description("Sheet name")),
			mcp.WithString("column", mcp.Required(), mcp.Description("Column name")),
			mcp.WithString("operator", mcp.Required(), mcp.Description("Operator: =, !=, >, <, >=, <=, contains")),
			mcp.WithString("value", mcp.Required(), mcp.Description("Filter value")),
		),
		s.handleFilterRows,
	)

	s.mcpSrv.AddTool(
		mcp.NewTool("group_by",
			mcp.WithDescription("Group by column and aggregate"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
			mcp.WithString("sheet", mcp.Required(), mcp.Description("Sheet name")),
			mcp.WithString("group_column", mcp.Required(), mcp.Description("Group by column")),
			mcp.WithString("agg_column", mcp.Required(), mcp.Description("Aggregate column")),
			mcp.WithString("agg_func", mcp.Required(), mcp.Description("Aggregate function: sum, avg, count, min, max")),
		),
		s.handleGroupBy,
	)

	s.mcpSrv.AddTool(
		mcp.NewTool("fill_missing_values",
			mcp.WithDescription("Fill missing values in a column"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
			mcp.WithString("sheet", mcp.Required(), mcp.Description("Sheet name")),
			mcp.WithString("column", mcp.Required(), mcp.Description("Column name")),
			mcp.WithString("default_value", mcp.Required(), mcp.Description("Default value to fill")),
		),
		s.handleFillMissingValues,
	)

	s.mcpSrv.AddTool(
		mcp.NewTool("replace_values",
			mcp.WithDescription("Replace values in a column"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
			mcp.WithString("sheet", mcp.Required(), mcp.Description("Sheet name")),
			mcp.WithString("column", mcp.Required(), mcp.Description("Column name")),
			mcp.WithString("old_value", mcp.Required(), mcp.Description("Value to replace")),
			mcp.WithString("new_value", mcp.Required(), mcp.Description("New value")),
		),
		s.handleReplaceValues,
	)

	s.mcpSrv.AddTool(
		mcp.NewTool("cleanup_sheet",
			mcp.WithDescription("Cleanup sheet (remove empty rows, trim whitespace)"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
			mcp.WithString("sheet", mcp.Required(), mcp.Description("Sheet name")),
		),
		s.handleCleanupSheet,
	)

	s.mcpSrv.AddTool(
		mcp.NewTool("deduplicate",
			mcp.WithDescription("Remove duplicate rows"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
			mcp.WithString("sheet", mcp.Required(), mcp.Description("Sheet name")),
			mcp.WithString("columns", mcp.Required(), mcp.Description("Comma-separated column names to check")),
		),
		s.handleDeduplicate,
	)

	// Phase 2: Formula tools
	s.mcpSrv.AddTool(
		mcp.NewTool("get_formula",
			mcp.WithDescription("Get cell formula"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
			mcp.WithString("sheet", mcp.Required(), mcp.Description("Sheet name")),
			mcp.WithString("cell", mcp.Required(), mcp.Description("Cell reference (e.g., A1)")),
		),
		s.handleGetFormula,
	)

	s.mcpSrv.AddTool(
		mcp.NewTool("set_formula",
			mcp.WithDescription("Set cell formula"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
			mcp.WithString("sheet", mcp.Required(), mcp.Description("Sheet name")),
			mcp.WithString("cell", mcp.Required(), mcp.Description("Cell reference (e.g., A1)")),
			mcp.WithString("formula", mcp.Required(), mcp.Description("Formula (e.g., =SUM(A1:A10))")),
		),
		s.handleSetFormula,
	)

	s.mcpSrv.AddTool(
		mcp.NewTool("fill_formula_column",
			mcp.WithDescription("Fill a formula down a column"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
			mcp.WithString("sheet", mcp.Required(), mcp.Description("Sheet name")),
			mcp.WithString("column", mcp.Required(), mcp.Description("Column letter (e.g., C)")),
			mcp.WithString("start_row", mcp.Required(), mcp.Description("Start row number")),
			mcp.WithString("end_row", mcp.Required(), mcp.Description("End row number")),
			mcp.WithString("formula_template", mcp.Required(), mcp.Description("Formula template with {row} placeholder")),
		),
		s.handleFillFormulaColumn,
	)

	// Phase 2: Style tools
	s.mcpSrv.AddTool(
		mcp.NewTool("auto_fit_columns",
			mcp.WithDescription("Auto-fit column widths to content"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
			mcp.WithString("sheet", mcp.Required(), mcp.Description("Sheet name")),
		),
		s.handleAutoFitColumns,
	)

	s.mcpSrv.AddTool(
		mcp.NewTool("auto_fit_rows",
			mcp.WithDescription("Auto-fit row heights to content"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
			mcp.WithString("sheet", mcp.Required(), mcp.Description("Sheet name")),
		),
		s.handleAutoFitRows,
	)

	s.mcpSrv.AddTool(
		mcp.NewTool("set_number_format",
			mcp.WithDescription("Set number format for a cell"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
			mcp.WithString("sheet", mcp.Required(), mcp.Description("Sheet name")),
			mcp.WithString("cell", mcp.Required(), mcp.Description("Cell reference (e.g., A1)")),
			mcp.WithString("format", mcp.Required(), mcp.Description("Number format (e.g., #,##0.00, 0.00%, yyyy-mm-dd)")),
		),
		s.handleSetNumberFormat,
	)

	slog.Debug("MCP tools registered")
}

func (s *Server) handleOpenWorkbook(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, err := req.RequireString("path")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	slog.Info("opening workbook", "path", path)
	wb, err := s.manager.Open(path)
	if err != nil {
		slog.Error("open workbook failed", "path", path, "error", err)
		return mcpError(err.Error()), nil
	}

	slog.Info("workbook opened", "id", wb.ID, "path", path)
	return mcpResult(wb.ID), nil
}

func (s *Server) handleSaveWorkbook(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("workbook_id")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	slog.Info("saving workbook", "id", id)
	if err := s.manager.Save(id); err != nil {
		slog.Error("save workbook failed", "id", id, "error", err)
		return mcpError(err.Error()), nil
	}

	slog.Info("workbook saved", "id", id)
	return mcpResult("saved"), nil
}

func (s *Server) handleSaveWorkbookAs(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("workbook_id")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	path, err := req.RequireString("path")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	slog.Info("saving workbook as", "id", id, "path", path)
	if err := s.manager.SaveAs(id, path); err != nil {
		slog.Error("save workbook as failed", "id", id, "path", path, "error", err)
		return mcpError(err.Error()), nil
	}

	slog.Info("workbook saved", "id", id, "path", path)
	return mcpResult("saved"), nil
}

func (s *Server) handleCloseWorkbook(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("workbook_id")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	slog.Info("closing workbook", "id", id)
	if err := s.manager.Close(id); err != nil {
		slog.Error("close workbook failed", "id", id, "error", err)
		return mcpError(err.Error()), nil
	}

	slog.Info("workbook closed", "id", id)
	return mcpResult("closed"), nil
}

func (s *Server) handleListWorkbooks(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	slog.Debug("listing workbooks")
	list := s.manager.List()

	type wbInfo struct {
		ID   string `json:"id"`
		Path string `json:"path"`
	}

	infos := make([]wbInfo, len(list))
	for i, wb := range list {
		infos[i] = wbInfo{ID: wb.ID, Path: wb.Path}
	}

	data, _ := json.Marshal(infos)
	slog.Debug("workbooks listed", "count", len(list))
	return mcpResult(string(data)), nil
}

func (s *Server) handleListSheets(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("workbook_id")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	slog.Debug("listing sheets", "workbook_id", id)
	wb, err := s.manager.Get(id)
	if err != nil {
		return mcpError(err.Error()), nil
	}

	sheets := worksheet.ListSheets(wb.File)
	data, _ := json.Marshal(sheets)
	slog.Debug("sheets listed", "workbook_id", id, "count", len(sheets))
	return mcpResult(string(data)), nil
}

func (s *Server) handleCreateSheet(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("workbook_id")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	name, err := req.RequireString("name")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	slog.Info("creating sheet", "workbook_id", id, "name", name)
	wb, err := s.manager.Get(id)
	if err != nil {
		return mcpError(err.Error()), nil
	}

	if _, err := worksheet.CreateSheet(wb.File, name); err != nil {
		slog.Error("create sheet failed", "workbook_id", id, "name", name, "error", err)
		return mcpError(err.Error()), nil
	}

	slog.Info("sheet created", "workbook_id", id, "name", name)
	return mcpResult("created"), nil
}

func (s *Server) handleDeleteSheet(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("workbook_id")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	name, err := req.RequireString("name")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	slog.Info("deleting sheet", "workbook_id", id, "name", name)
	wb, err := s.manager.Get(id)
	if err != nil {
		return mcpError(err.Error()), nil
	}

	if err := worksheet.DeleteSheet(wb.File, name); err != nil {
		slog.Error("delete sheet failed", "workbook_id", id, "name", name, "error", err)
		return mcpError(err.Error()), nil
	}

	slog.Info("sheet deleted", "workbook_id", id, "name", name)
	return mcpResult("deleted"), nil
}

func (s *Server) handleRenameSheet(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("workbook_id")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	oldName, err := req.RequireString("old_name")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	newName, err := req.RequireString("new_name")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	slog.Info("renaming sheet", "workbook_id", id, "old_name", oldName, "new_name", newName)
	wb, err := s.manager.Get(id)
	if err != nil {
		return mcpError(err.Error()), nil
	}

	if err := worksheet.RenameSheet(wb.File, oldName, newName); err != nil {
		slog.Error("rename sheet failed", "workbook_id", id, "old_name", oldName, "new_name", newName, "error", err)
		return mcpError(err.Error()), nil
	}

	slog.Info("sheet renamed", "workbook_id", id, "old_name", oldName, "new_name", newName)
	return mcpResult("renamed"), nil
}

func (s *Server) handleGetSheetInfo(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("workbook_id")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	sheet, err := req.RequireString("sheet")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	slog.Debug("getting sheet info", "workbook_id", id, "sheet", sheet)
	wb, err := s.manager.Get(id)
	if err != nil {
		return mcpError(err.Error()), nil
	}

	info, err := worksheet.GetSheetInfo(wb.File, sheet)
	if err != nil {
		slog.Error("get sheet info failed", "workbook_id", id, "sheet", sheet, "error", err)
		return mcpError(err.Error()), nil
	}

	data, _ := json.Marshal(info)
	return mcpResult(string(data)), nil
}

func (s *Server) handleGetCell(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	slog.Debug("getting cell", "workbook_id", id, "sheet", sheet, "cell", cell)
	wb, err := s.manager.Get(id)
	if err != nil {
		return mcpError(err.Error()), nil
	}

	val, err := rangeop.GetCell(wb.File, sheet, cell)
	if err != nil {
		slog.Error("get cell failed", "workbook_id", id, "sheet", sheet, "cell", cell, "error", err)
		return mcpError(err.Error()), nil
	}

	return mcpResult(val), nil
}

func (s *Server) handleSetCell(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	value, err := req.RequireString("value")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	slog.Debug("setting cell", "workbook_id", id, "sheet", sheet, "cell", cell)
	wb, err := s.manager.Get(id)
	if err != nil {
		return mcpError(err.Error()), nil
	}

	if err := rangeop.SetCell(wb.File, sheet, cell, value); err != nil {
		slog.Error("set cell failed", "workbook_id", id, "sheet", sheet, "cell", cell, "error", err)
		return mcpError(err.Error()), nil
	}

	return mcpResult("ok"), nil
}

func (s *Server) handleGetRange(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("workbook_id")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	sheet, err := req.RequireString("sheet")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	r, err := req.RequireString("range")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	slog.Debug("getting range", "workbook_id", id, "sheet", sheet, "range", r)
	wb, err := s.manager.Get(id)
	if err != nil {
		return mcpError(err.Error()), nil
	}

	data, err := rangeop.GetRange(wb.File, sheet, r)
	if err != nil {
		slog.Error("get range failed", "workbook_id", id, "sheet", sheet, "range", r, "error", err)
		return mcpError(err.Error()), nil
	}

	result, _ := json.Marshal(data)
	return mcpResult(string(result)), nil
}

func (s *Server) handleSetRange(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("workbook_id")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	sheet, err := req.RequireString("sheet")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	startCell, err := req.RequireString("start_cell")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	dataStr, err := req.RequireString("data")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	var data [][]any
	if err := json.Unmarshal([]byte(dataStr), &data); err != nil {
		return mcpError(fmt.Sprintf("invalid data format: %v", err)), nil
	}

	slog.Debug("setting range", "workbook_id", id, "sheet", sheet, "start_cell", startCell, "rows", len(data))
	wb, err := s.manager.Get(id)
	if err != nil {
		return mcpError(err.Error()), nil
	}

	if err := rangeop.SetRange(wb.File, sheet, startCell, data); err != nil {
		slog.Error("set range failed", "workbook_id", id, "sheet", sheet, "start_cell", startCell, "error", err)
		return mcpError(err.Error()), nil
	}

	return mcpResult("ok"), nil
}

func (s *Server) handleAppendRows(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("workbook_id")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	sheet, err := req.RequireString("sheet")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	dataStr, err := req.RequireString("data")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	var data [][]any
	if err := json.Unmarshal([]byte(dataStr), &data); err != nil {
		return mcpError(fmt.Sprintf("invalid data format: %v", err)), nil
	}

	slog.Debug("appending rows", "workbook_id", id, "sheet", sheet, "rows", len(data))
	wb, err := s.manager.Get(id)
	if err != nil {
		return mcpError(err.Error()), nil
	}

	if err := rangeop.AppendRows(wb.File, sheet, data); err != nil {
		slog.Error("append rows failed", "workbook_id", id, "sheet", sheet, "error", err)
		return mcpError(err.Error()), nil
	}

	slog.Info("rows appended", "workbook_id", id, "sheet", sheet, "rows", len(data))
	return mcpResult("ok"), nil
}

func (s *Server) handleGetWorkbookInfo(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("workbook_id")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	slog.Debug("getting workbook info", "id", id)
	info, err := s.manager.GetInfo(id)
	if err != nil {
		slog.Error("get workbook info failed", "id", id, "error", err)
		return mcpError(err.Error()), nil
	}

	data, _ := json.Marshal(info)
	return mcpResult(string(data)), nil
}

func (s *Server) handleCopySheet(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("workbook_id")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	name, err := req.RequireString("name")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	slog.Info("copying sheet", "workbook_id", id, "name", name)
	wb, err := s.manager.Get(id)
	if err != nil {
		return mcpError(err.Error()), nil
	}

	newName, err := worksheet.CopySheet(wb.File, name)
	if err != nil {
		slog.Error("copy sheet failed", "workbook_id", id, "name", name, "error", err)
		return mcpError(err.Error()), nil
	}

	slog.Info("sheet copied", "workbook_id", id, "from", name, "to", newName)
	return mcpResult(newName), nil
}

func (s *Server) handleSetActiveSheet(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("workbook_id")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	name, err := req.RequireString("name")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	slog.Debug("setting active sheet", "workbook_id", id, "name", name)
	wb, err := s.manager.Get(id)
	if err != nil {
		return mcpError(err.Error()), nil
	}

	if err := worksheet.SetActiveSheet(wb.File, name); err != nil {
		slog.Error("set active sheet failed", "workbook_id", id, "name", name, "error", err)
		return mcpError(err.Error()), nil
	}

	return mcpResult("ok"), nil
}

func (s *Server) handleClearCell(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	slog.Debug("clearing cell", "workbook_id", id, "sheet", sheet, "cell", cell)
	wb, err := s.manager.Get(id)
	if err != nil {
		return mcpError(err.Error()), nil
	}

	if err := rangeop.ClearCell(wb.File, sheet, cell); err != nil {
		slog.Error("clear cell failed", "workbook_id", id, "sheet", sheet, "cell", cell, "error", err)
		return mcpError(err.Error()), nil
	}

	return mcpResult("ok"), nil
}

func (s *Server) handleInsertRows(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	count, err := req.RequireString("count")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	slog.Debug("inserting rows", "workbook_id", id, "sheet", sheet, "row", row, "count", count)
	wb, err := s.manager.Get(id)
	if err != nil {
		return mcpError(err.Error()), nil
	}

	rowNum, countNum := parseInt(row), parseInt(count)
	if err := rangeop.InsertRows(wb.File, sheet, rowNum, countNum); err != nil {
		slog.Error("insert rows failed", "workbook_id", id, "sheet", sheet, "error", err)
		return mcpError(err.Error()), nil
	}

	return mcpResult("ok"), nil
}

func (s *Server) handleDeleteRows(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	count, err := req.RequireString("count")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	slog.Debug("deleting rows", "workbook_id", id, "sheet", sheet, "row", row, "count", count)
	wb, err := s.manager.Get(id)
	if err != nil {
		return mcpError(err.Error()), nil
	}

	rowNum, countNum := parseInt(row), parseInt(count)
	if err := rangeop.DeleteRows(wb.File, sheet, rowNum, countNum); err != nil {
		slog.Error("delete rows failed", "workbook_id", id, "sheet", sheet, "error", err)
		return mcpError(err.Error()), nil
	}

	return mcpResult("ok"), nil
}

func (s *Server) handleInsertCols(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("workbook_id")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	sheet, err := req.RequireString("sheet")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	col, err := req.RequireString("col")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	count, err := req.RequireString("count")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	slog.Debug("inserting columns", "workbook_id", id, "sheet", sheet, "col", col, "count", count)
	wb, err := s.manager.Get(id)
	if err != nil {
		return mcpError(err.Error()), nil
	}

	countNum := parseInt(count)
	if err := rangeop.InsertCols(wb.File, sheet, col, countNum); err != nil {
		slog.Error("insert columns failed", "workbook_id", id, "sheet", sheet, "error", err)
		return mcpError(err.Error()), nil
	}

	return mcpResult("ok"), nil
}

func (s *Server) handleDeleteCols(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("workbook_id")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	sheet, err := req.RequireString("sheet")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	col, err := req.RequireString("col")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	count, err := req.RequireString("count")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	slog.Debug("deleting columns", "workbook_id", id, "sheet", sheet, "col", col, "count", count)
	wb, err := s.manager.Get(id)
	if err != nil {
		return mcpError(err.Error()), nil
	}

	countNum := parseInt(count)
	if err := rangeop.DeleteCols(wb.File, sheet, col, countNum); err != nil {
		slog.Error("delete columns failed", "workbook_id", id, "sheet", sheet, "error", err)
		return mcpError(err.Error()), nil
	}

	return mcpResult("ok"), nil
}

// Phase 2 handlers

func (s *Server) handleTableInfo(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("workbook_id")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	sheet, err := req.RequireString("sheet")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	slog.Debug("getting table info", "workbook_id", id, "sheet", sheet)
	wb, err := s.manager.Get(id)
	if err != nil {
		return mcpError(err.Error()), nil
	}

	info, err := analysis.GetTableInfo(wb.File, sheet)
	if err != nil {
		slog.Error("get table info failed", "workbook_id", id, "sheet", sheet, "error", err)
		return mcpError(err.Error()), nil
	}

	data, _ := json.Marshal(info)
	return mcpResult(string(data)), nil
}

func (s *Server) handleColumnTypes(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("workbook_id")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	sheet, err := req.RequireString("sheet")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	slog.Debug("getting column types", "workbook_id", id, "sheet", sheet)
	wb, err := s.manager.Get(id)
	if err != nil {
		return mcpError(err.Error()), nil
	}

	types, err := analysis.GetColumnTypes(wb.File, sheet, 100)
	if err != nil {
		slog.Error("get column types failed", "workbook_id", id, "sheet", sheet, "error", err)
		return mcpError(err.Error()), nil
	}

	data, _ := json.Marshal(types)
	return mcpResult(string(data)), nil
}

func (s *Server) handleSheetOverview(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("workbook_id")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	sheet, err := req.RequireString("sheet")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	slog.Debug("getting sheet overview", "workbook_id", id, "sheet", sheet)
	wb, err := s.manager.Get(id)
	if err != nil {
		return mcpError(err.Error()), nil
	}

	overview, err := analysis.GetSheetOverview(wb.File, sheet)
	if err != nil {
		slog.Error("get sheet overview failed", "workbook_id", id, "sheet", sheet, "error", err)
		return mcpError(err.Error()), nil
	}

	data, _ := json.Marshal(overview)
	return mcpResult(string(data)), nil
}

func (s *Server) handleAddComputedColumn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("workbook_id")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	sheet, err := req.RequireString("sheet")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	colName, err := req.RequireString("col_name")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	formula, err := req.RequireString("formula")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	slog.Info("adding computed column", "workbook_id", id, "sheet", sheet, "col_name", colName)
	wb, err := s.manager.Get(id)
	if err != nil {
		return mcpError(err.Error()), nil
	}

	if err := dataops.AddComputedColumn(wb.File, sheet, colName, formula); err != nil {
		slog.Error("add computed column failed", "workbook_id", id, "sheet", sheet, "error", err)
		return mcpError(err.Error()), nil
	}

	return mcpResult("ok"), nil
}

func (s *Server) handleSortTable(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("workbook_id")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	sheet, err := req.RequireString("sheet")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	columns, err := req.RequireString("columns")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	order := req.GetString("order", "asc")
	ascending := order != "desc"

	slog.Info("sorting table", "workbook_id", id, "sheet", sheet, "columns", columns, "order", order)
	wb, err := s.manager.Get(id)
	if err != nil {
		return mcpError(err.Error()), nil
	}

	cols := splitColumns(columns)
	if err := dataops.SortTable(wb.File, sheet, cols, ascending); err != nil {
		slog.Error("sort table failed", "workbook_id", id, "sheet", sheet, "error", err)
		return mcpError(err.Error()), nil
	}

	return mcpResult("ok"), nil
}

func (s *Server) handleFilterRows(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("workbook_id")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	sheet, err := req.RequireString("sheet")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	column, err := req.RequireString("column")
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

	slog.Debug("filtering rows", "workbook_id", id, "sheet", sheet, "column", column, "operator", operator, "value", value)
	wb, err := s.manager.Get(id)
	if err != nil {
		return mcpError(err.Error()), nil
	}

	result, err := dataops.FilterRows(wb.File, sheet, column, operator, value)
	if err != nil {
		slog.Error("filter rows failed", "workbook_id", id, "sheet", sheet, "error", err)
		return mcpError(err.Error()), nil
	}

	data, _ := json.Marshal(result)
	return mcpResult(string(data)), nil
}

func (s *Server) handleGroupBy(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("workbook_id")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	sheet, err := req.RequireString("sheet")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	groupCol, err := req.RequireString("group_column")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	aggCol, err := req.RequireString("agg_column")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	aggFunc, err := req.RequireString("agg_func")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	slog.Debug("group by", "workbook_id", id, "sheet", sheet, "group_col", groupCol, "agg_col", aggCol, "agg_func", aggFunc)
	wb, err := s.manager.Get(id)
	if err != nil {
		return mcpError(err.Error()), nil
	}

	resultSheet, err := dataops.GroupBy(wb.File, sheet, groupCol, aggCol, aggFunc)
	if err != nil {
		slog.Error("group by failed", "workbook_id", id, "sheet", sheet, "error", err)
		return mcpError(err.Error()), nil
	}

	return mcpResult(fmt.Sprintf("Group by completed. Result written to sheet: %s", resultSheet)), nil
}

func (s *Server) handleFillMissingValues(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("workbook_id")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	sheet, err := req.RequireString("sheet")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	column, err := req.RequireString("column")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	defaultValue, err := req.RequireString("default_value")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	slog.Info("filling missing values", "workbook_id", id, "sheet", sheet, "column", column)
	wb, err := s.manager.Get(id)
	if err != nil {
		return mcpError(err.Error()), nil
	}

	if err := dataops.FillMissingValues(wb.File, sheet, column, defaultValue); err != nil {
		slog.Error("fill missing values failed", "workbook_id", id, "sheet", sheet, "error", err)
		return mcpError(err.Error()), nil
	}

	return mcpResult("ok"), nil
}

func (s *Server) handleReplaceValues(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("workbook_id")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	sheet, err := req.RequireString("sheet")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	column, err := req.RequireString("column")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	oldValue, err := req.RequireString("old_value")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	newValue, err := req.RequireString("new_value")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	slog.Info("replacing values", "workbook_id", id, "sheet", sheet, "column", column, "old", oldValue, "new", newValue)
	wb, err := s.manager.Get(id)
	if err != nil {
		return mcpError(err.Error()), nil
	}

	if err := dataops.ReplaceValues(wb.File, sheet, column, oldValue, newValue); err != nil {
		slog.Error("replace values failed", "workbook_id", id, "sheet", sheet, "error", err)
		return mcpError(err.Error()), nil
	}

	return mcpResult("ok"), nil
}

func (s *Server) handleCleanupSheet(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("workbook_id")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	sheet, err := req.RequireString("sheet")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	slog.Info("cleaning up sheet", "workbook_id", id, "sheet", sheet)
	wb, err := s.manager.Get(id)
	if err != nil {
		return mcpError(err.Error()), nil
	}

	if err := dataops.CleanupSheet(wb.File, sheet); err != nil {
		slog.Error("cleanup sheet failed", "workbook_id", id, "sheet", sheet, "error", err)
		return mcpError(err.Error()), nil
	}

	return mcpResult("ok"), nil
}

func (s *Server) handleDeduplicate(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("workbook_id")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	sheet, err := req.RequireString("sheet")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	columns, err := req.RequireString("columns")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	slog.Info("deduplicating", "workbook_id", id, "sheet", sheet, "columns", columns)
	wb, err := s.manager.Get(id)
	if err != nil {
		return mcpError(err.Error()), nil
	}

	cols := splitColumns(columns)
	if err := dataops.Deduplicate(wb.File, sheet, cols); err != nil {
		slog.Error("deduplicate failed", "workbook_id", id, "sheet", sheet, "error", err)
		return mcpError(err.Error()), nil
	}

	return mcpResult("ok"), nil
}

func (s *Server) handleGetFormula(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	slog.Debug("getting formula", "workbook_id", id, "sheet", sheet, "cell", cell)
	wb, err := s.manager.Get(id)
	if err != nil {
		return mcpError(err.Error()), nil
	}

	formula, err := formulaPkg.GetFormula(wb.File, sheet, cell)
	if err != nil {
		slog.Error("get formula failed", "workbook_id", id, "sheet", sheet, "cell", cell, "error", err)
		return mcpError(err.Error()), nil
	}

	return mcpResult(formula), nil
}

func (s *Server) handleSetFormula(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	formulaStr, err := req.RequireString("formula")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	slog.Debug("setting formula", "workbook_id", id, "sheet", sheet, "cell", cell)
	wb, err := s.manager.Get(id)
	if err != nil {
		return mcpError(err.Error()), nil
	}

	if err := formulaPkg.SetFormula(wb.File, sheet, cell, formulaStr); err != nil {
		slog.Error("set formula failed", "workbook_id", id, "sheet", sheet, "cell", cell, "error", err)
		return mcpError(err.Error()), nil
	}

	return mcpResult("ok"), nil
}

func (s *Server) handleFillFormulaColumn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("workbook_id")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	sheet, err := req.RequireString("sheet")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	column, err := req.RequireString("column")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	startRowStr, err := req.RequireString("start_row")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	endRowStr, err := req.RequireString("end_row")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	formulaTmpl, err := req.RequireString("formula_template")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	startRow := parseInt(startRowStr)
	endRow := parseInt(endRowStr)

	slog.Debug("filling formula column", "workbook_id", id, "sheet", sheet, "column", column, "start_row", startRow, "end_row", endRow)
	wb, err := s.manager.Get(id)
	if err != nil {
		return mcpError(err.Error()), nil
	}

	if err := formulaPkg.FillFormulaColumn(wb.File, sheet, column, startRow, endRow, formulaTmpl); err != nil {
		slog.Error("fill formula column failed", "workbook_id", id, "sheet", sheet, "error", err)
		return mcpError(err.Error()), nil
	}

	return mcpResult("ok"), nil
}

func (s *Server) handleAutoFitColumns(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("workbook_id")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	sheet, err := req.RequireString("sheet")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	slog.Debug("auto fitting columns", "workbook_id", id, "sheet", sheet)
	wb, err := s.manager.Get(id)
	if err != nil {
		return mcpError(err.Error()), nil
	}

	if err := stylePkg.AutoFitColumns(wb.File, sheet); err != nil {
		slog.Error("auto fit columns failed", "workbook_id", id, "sheet", sheet, "error", err)
		return mcpError(err.Error()), nil
	}

	return mcpResult("ok"), nil
}

func (s *Server) handleAutoFitRows(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("workbook_id")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	sheet, err := req.RequireString("sheet")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	slog.Debug("auto fitting rows", "workbook_id", id, "sheet", sheet)
	wb, err := s.manager.Get(id)
	if err != nil {
		return mcpError(err.Error()), nil
	}

	if err := stylePkg.AutoFitRows(wb.File, sheet); err != nil {
		slog.Error("auto fit rows failed", "workbook_id", id, "sheet", sheet, "error", err)
		return mcpError(err.Error()), nil
	}

	return mcpResult("ok"), nil
}

func (s *Server) handleSetNumberFormat(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	format, err := req.RequireString("format")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	slog.Debug("setting number format", "workbook_id", id, "sheet", sheet, "cell", cell, "format", format)
	wb, err := s.manager.Get(id)
	if err != nil {
		return mcpError(err.Error()), nil
	}

	if err := stylePkg.SetNumberFormat(wb.File, sheet, cell, format); err != nil {
		slog.Error("set number format failed", "workbook_id", id, "sheet", sheet, "cell", cell, "error", err)
		return mcpError(err.Error()), nil
	}

	return mcpResult("ok"), nil
}

func splitColumns(columns string) []string {
	result := []string{}
	for _, col := range splitString(columns, ",") {
		col = trimSpace(col)
		if col != "" {
			result = append(result, col)
		}
	}
	return result
}

func splitString(s, sep string) []string {
	if s == "" {
		return []string{}
	}
	return strings.Split(s, sep)
}

func trimSpace(s string) string {
	return strings.TrimSpace(s)
}

func parseInt(s string) int {
	n := 0
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		}
	}
	return n
}

func mcpResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: text,
			},
		},
	}
}

func mcpError(msg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: msg,
			},
		},
		IsError: true,
	}
}
