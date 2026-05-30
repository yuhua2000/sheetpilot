package mcp

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/yuhua2000/sheetpilot/internal/analysis"
)

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
