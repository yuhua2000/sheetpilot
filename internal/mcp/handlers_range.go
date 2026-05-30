package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/yuhua2000/sheetpilot/internal/rangeop"
)

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

func (s *Server) handleCopyRange(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("workbook_id")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	sheet, err := req.RequireString("sheet")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	srcRange, err := req.RequireString("source_range")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	dstCell, err := req.RequireString("dest_cell")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	slog.Debug("copying range", "workbook_id", id, "sheet", sheet, "src", srcRange, "dst", dstCell)
	wb, err := s.manager.Get(id)
	if err != nil {
		return mcpError(err.Error()), nil
	}

	if err := rangeop.CopyRange(wb.File, sheet, srcRange, dstCell); err != nil {
		slog.Error("copy range failed", "workbook_id", id, "sheet", sheet, "error", err)
		return mcpError(err.Error()), nil
	}

	return mcpResult("ok"), nil
}

func (s *Server) handleMoveRange(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("workbook_id")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	sheet, err := req.RequireString("sheet")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	srcRange, err := req.RequireString("source_range")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	dstCell, err := req.RequireString("dest_cell")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	slog.Debug("moving range", "workbook_id", id, "sheet", sheet, "src", srcRange, "dst", dstCell)
	wb, err := s.manager.Get(id)
	if err != nil {
		return mcpError(err.Error()), nil
	}

	if err := rangeop.MoveRange(wb.File, sheet, srcRange, dstCell); err != nil {
		slog.Error("move range failed", "workbook_id", id, "sheet", sheet, "error", err)
		return mcpError(err.Error()), nil
	}

	return mcpResult("ok"), nil
}

func (s *Server) handleFindReplace(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("workbook_id")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	sheet, err := req.RequireString("sheet")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	find, err := req.RequireString("find")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	replace, err := req.RequireString("replace")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	slog.Debug("find replace", "workbook_id", id, "sheet", sheet, "find", find, "replace", replace)
	wb, err := s.manager.Get(id)
	if err != nil {
		return mcpError(err.Error()), nil
	}

	count, err := rangeop.FindReplace(wb.File, sheet, find, replace, "")
	if err != nil {
		slog.Error("find replace failed", "workbook_id", id, "sheet", sheet, "error", err)
		return mcpError(err.Error()), nil
	}

	return mcpResult(fmt.Sprintf("Replaced %d occurrences", count)), nil
}

func (s *Server) handleAddComment(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	author, err := req.RequireString("author")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	text, err := req.RequireString("text")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	slog.Debug("adding comment", "workbook_id", id, "sheet", sheet, "cell", cell, "author", author)
	wb, err := s.manager.Get(id)
	if err != nil {
		return mcpError(err.Error()), nil
	}

	if err := rangeop.AddComment(wb.File, sheet, cell, author, text); err != nil {
		slog.Error("add comment failed", "workbook_id", id, "sheet", sheet, "cell", cell, "error", err)
		return mcpError(err.Error()), nil
	}

	return mcpResult("ok"), nil
}

func (s *Server) handleAddHyperlink(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	link, err := req.RequireString("link")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	display, err := req.RequireString("display")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	slog.Debug("adding hyperlink", "workbook_id", id, "sheet", sheet, "cell", cell, "link", link)
	wb, err := s.manager.Get(id)
	if err != nil {
		return mcpError(err.Error()), nil
	}

	if err := rangeop.AddHyperlink(wb.File, sheet, cell, link, display); err != nil {
		slog.Error("add hyperlink failed", "workbook_id", id, "sheet", sheet, "cell", cell, "error", err)
		return mcpError(err.Error()), nil
	}

	return mcpResult("ok"), nil
}

func (s *Server) handleSetDataValidation(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	optionsStr, err := req.RequireString("options")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	options := splitColumns(optionsStr)

	slog.Debug("setting data validation", "workbook_id", id, "sheet", sheet, "range", rangeRef, "options", options)
	wb, err := s.manager.Get(id)
	if err != nil {
		return mcpError(err.Error()), nil
	}

	if err := rangeop.SetDataValidation(wb.File, sheet, rangeRef, options); err != nil {
		slog.Error("set data validation failed", "workbook_id", id, "sheet", sheet, "error", err)
		return mcpError(err.Error()), nil
	}

	return mcpResult("ok"), nil
}
