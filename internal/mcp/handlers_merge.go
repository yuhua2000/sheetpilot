package mcp

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/yuhua2000/sheetpilot/internal/rangeop"
)

func (s *Server) registerMergeCellTools() {
	s.mcpSrv.AddTool(
		mcp.NewTool("merge_cells",
			mcp.WithDescription("Merge a range of cells"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
			mcp.WithString("sheet", mcp.Required(), mcp.Description("Sheet name")),
			mcp.WithString("top_left", mcp.Required(), mcp.Description("Top-left cell (e.g., A1)")),
			mcp.WithString("bottom_right", mcp.Required(), mcp.Description("Bottom-right cell (e.g., C3)")),
		),
		s.handleMergeCells,
	)

	s.mcpSrv.AddTool(
		mcp.NewTool("unmerge_cells",
			mcp.WithDescription("Unmerge a range of cells"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
			mcp.WithString("sheet", mcp.Required(), mcp.Description("Sheet name")),
			mcp.WithString("top_left", mcp.Required(), mcp.Description("Top-left cell (e.g., A1)")),
			mcp.WithString("bottom_right", mcp.Required(), mcp.Description("Bottom-right cell (e.g., C3)")),
		),
		s.handleUnmergeCells,
	)

	s.mcpSrv.AddTool(
		mcp.NewTool("get_merged_cells",
			mcp.WithDescription("Get all merged cell ranges in a sheet"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
			mcp.WithString("sheet", mcp.Required(), mcp.Description("Sheet name")),
		),
		s.handleGetMergedCells,
	)
}

func (s *Server) handleMergeCells(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("workbook_id")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	sheet, err := req.RequireString("sheet")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	topLeft, err := req.RequireString("top_left")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	bottomRight, err := req.RequireString("bottom_right")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	slog.Debug("merging cells", "workbook_id", id, "sheet", sheet, "top_left", topLeft, "bottom_right", bottomRight)
	wb, err := s.manager.Get(id)
	if err != nil {
		return mcpError(err.Error()), nil
	}

	if err := rangeop.MergeCells(wb.File, sheet, topLeft, bottomRight); err != nil {
		slog.Error("merge cells failed", "workbook_id", id, "sheet", sheet, "error", err)
		return mcpError(err.Error()), nil
	}

	return mcpResult("ok"), nil
}

func (s *Server) handleUnmergeCells(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("workbook_id")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	sheet, err := req.RequireString("sheet")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	topLeft, err := req.RequireString("top_left")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	bottomRight, err := req.RequireString("bottom_right")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	slog.Debug("unmerging cells", "workbook_id", id, "sheet", sheet, "top_left", topLeft, "bottom_right", bottomRight)
	wb, err := s.manager.Get(id)
	if err != nil {
		return mcpError(err.Error()), nil
	}

	if err := rangeop.UnmergeCells(wb.File, sheet, topLeft, bottomRight); err != nil {
		slog.Error("unmerge cells failed", "workbook_id", id, "sheet", sheet, "error", err)
		return mcpError(err.Error()), nil
	}

	return mcpResult("ok"), nil
}

func (s *Server) handleGetMergedCells(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("workbook_id")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	sheet, err := req.RequireString("sheet")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	slog.Debug("getting merged cells", "workbook_id", id, "sheet", sheet)
	wb, err := s.manager.Get(id)
	if err != nil {
		return mcpError(err.Error()), nil
	}

	merged, err := rangeop.GetMergedCells(wb.File, sheet)
	if err != nil {
		slog.Error("get merged cells failed", "workbook_id", id, "sheet", sheet, "error", err)
		return mcpError(err.Error()), nil
	}

	data, _ := json.Marshal(merged)
	return mcpResult(string(data)), nil
}
