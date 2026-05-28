package mcp

import (
	"context"
	"log/slog"

	"github.com/mark3labs/mcp-go/mcp"
	formulaPkg "github.com/yuhua2000/sheetpilot/internal/formula"
	stylePkg "github.com/yuhua2000/sheetpilot/internal/style"
)

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
