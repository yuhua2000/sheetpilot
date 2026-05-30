package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/yuhua2000/sheetpilot/internal/dataops"
)

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

func (s *Server) handleSplitSheet(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	slog.Info("splitting sheet", "workbook_id", id, "sheet", sheet, "column", column)
	wb, err := s.manager.Get(id)
	if err != nil {
		return mcpError(err.Error()), nil
	}

	newSheets, err := dataops.SplitSheet(wb.File, sheet, column)
	if err != nil {
		slog.Error("split sheet failed", "workbook_id", id, "sheet", sheet, "error", err)
		return mcpError(err.Error()), nil
	}

	data, _ := json.Marshal(newSheets)
	return mcpResult(string(data)), nil
}

func (s *Server) handleMergeSheets(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("workbook_id")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	sheetsStr, err := req.RequireString("sheets")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	destSheet, err := req.RequireString("dest_sheet")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	sheets := splitColumns(sheetsStr)

	slog.Info("merging sheets", "workbook_id", id, "sheets", sheets, "dest", destSheet)
	wb, err := s.manager.Get(id)
	if err != nil {
		return mcpError(err.Error()), nil
	}

	if err := dataops.MergeSheets(wb.File, sheets, destSheet); err != nil {
		slog.Error("merge sheets failed", "workbook_id", id, "error", err)
		return mcpError(err.Error()), nil
	}

	return mcpResult("ok"), nil
}
