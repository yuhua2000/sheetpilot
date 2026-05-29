package mcp

import (
	"log/slog"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/yuhua2000/sheetpilot/internal/workbook"
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

	slog.Info("MCP server created", "tools", 59)
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

	// Phase 1: Workbook tools
	s.registerWorkbookTools()

	// Phase 1: Sheet tools
	s.registerSheetTools()

	// Phase 1: Cell & Range tools
	s.registerRangeTools()

	// Phase 2: Analysis tools
	s.registerAnalysisTools()

	// Phase 2: Data operation tools
	s.registerDataOpsTools()

	// Phase 2: Formula tools
	s.registerFormulaTools()

	// Phase 2: Style tools
	s.registerStyleTools()

	// Common tools
	s.registerCommonTools()

	// Phase 3: Chart tools
	s.registerChartTools()

	// Phase 3: Advanced style tools
	s.registerAdvancedStyleTools()

	// Phase 3: Merge cell tools
	s.registerMergeCellTools()

	// Phase 3: Batch tools
	s.registerBatchTools()

	slog.Debug("MCP tools registered")
}

func (s *Server) registerWorkbookTools() {
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

	s.mcpSrv.AddTool(
		mcp.NewTool("get_workbook_info",
			mcp.WithDescription("Get workbook metadata"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
		),
		s.handleGetWorkbookInfo,
	)
}

func (s *Server) registerSheetTools() {
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

	s.mcpSrv.AddTool(
		mcp.NewTool("get_sheet_info",
			mcp.WithDescription("Get sheet metadata"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
			mcp.WithString("sheet", mcp.Required(), mcp.Description("Sheet name")),
		),
		s.handleGetSheetInfo,
	)
}

func (s *Server) registerRangeTools() {
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

	s.mcpSrv.AddTool(
		mcp.NewTool("clear_cell",
			mcp.WithDescription("Clear a cell value"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
			mcp.WithString("sheet", mcp.Required(), mcp.Description("Sheet name")),
			mcp.WithString("cell", mcp.Required(), mcp.Description("Cell reference (e.g., A1)")),
		),
		s.handleClearCell,
	)

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
}

func (s *Server) registerAnalysisTools() {
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
}

func (s *Server) registerDataOpsTools() {
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

	s.mcpSrv.AddTool(
		mcp.NewTool("split_sheet",
			mcp.WithDescription("Split sheet into multiple sheets by column value"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
			mcp.WithString("sheet", mcp.Required(), mcp.Description("Sheet name")),
			mcp.WithString("column", mcp.Required(), mcp.Description("Column to split by")),
		),
		s.handleSplitSheet,
	)

	s.mcpSrv.AddTool(
		mcp.NewTool("merge_sheets",
			mcp.WithDescription("Merge multiple sheets into one"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
			mcp.WithString("sheets", mcp.Required(), mcp.Description("Comma-separated sheet names to merge")),
			mcp.WithString("dest_sheet", mcp.Required(), mcp.Description("Destination sheet name")),
		),
		s.handleMergeSheets,
	)
}

func (s *Server) registerFormulaTools() {
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
}

func (s *Server) registerStyleTools() {
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
}

func (s *Server) registerCommonTools() {
	s.mcpSrv.AddTool(
		mcp.NewTool("copy_range",
			mcp.WithDescription("Copy a range to another location"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
			mcp.WithString("sheet", mcp.Required(), mcp.Description("Sheet name")),
			mcp.WithString("source_range", mcp.Required(), mcp.Description("Source range (e.g., A1:C5)")),
			mcp.WithString("dest_cell", mcp.Required(), mcp.Description("Destination cell (e.g., D1)")),
		),
		s.handleCopyRange,
	)

	s.mcpSrv.AddTool(
		mcp.NewTool("move_range",
			mcp.WithDescription("Move a range to another location"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
			mcp.WithString("sheet", mcp.Required(), mcp.Description("Sheet name")),
			mcp.WithString("source_range", mcp.Required(), mcp.Description("Source range (e.g., A1:C5)")),
			mcp.WithString("dest_cell", mcp.Required(), mcp.Description("Destination cell (e.g., D1)")),
		),
		s.handleMoveRange,
	)

	s.mcpSrv.AddTool(
		mcp.NewTool("find_replace",
			mcp.WithDescription("Find and replace text in the sheet"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
			mcp.WithString("sheet", mcp.Required(), mcp.Description("Sheet name")),
			mcp.WithString("find", mcp.Required(), mcp.Description("Text to find")),
			mcp.WithString("replace", mcp.Required(), mcp.Description("Replacement text")),
		),
		s.handleFindReplace,
	)

	s.mcpSrv.AddTool(
		mcp.NewTool("add_comment",
			mcp.WithDescription("Add a comment to a cell"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
			mcp.WithString("sheet", mcp.Required(), mcp.Description("Sheet name")),
			mcp.WithString("cell", mcp.Required(), mcp.Description("Cell reference (e.g., A1)")),
			mcp.WithString("author", mcp.Required(), mcp.Description("Comment author")),
			mcp.WithString("text", mcp.Required(), mcp.Description("Comment text")),
		),
		s.handleAddComment,
	)

	s.mcpSrv.AddTool(
		mcp.NewTool("add_hyperlink",
			mcp.WithDescription("Add a hyperlink to a cell"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
			mcp.WithString("sheet", mcp.Required(), mcp.Description("Sheet name")),
			mcp.WithString("cell", mcp.Required(), mcp.Description("Cell reference (e.g., A1)")),
			mcp.WithString("link", mcp.Required(), mcp.Description("URL or link target")),
			mcp.WithString("display", mcp.Required(), mcp.Description("Display text")),
		),
		s.handleAddHyperlink,
	)

	s.mcpSrv.AddTool(
		mcp.NewTool("set_data_validation",
			mcp.WithDescription("Set data validation (dropdown list) for a cell range"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
			mcp.WithString("sheet", mcp.Required(), mcp.Description("Sheet name")),
			mcp.WithString("range", mcp.Required(), mcp.Description("Cell range (e.g., A1:A10)")),
			mcp.WithString("options", mcp.Required(), mcp.Description("Comma-separated list of options")),
		),
		s.handleSetDataValidation,
	)
}
