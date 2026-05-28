package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/yuhua2000/sheetpilot/internal/rangeop"
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

	slog.Info("MCP server created", "tools", 23)
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
