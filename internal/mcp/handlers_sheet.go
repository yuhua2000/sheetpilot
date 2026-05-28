package mcp

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/yuhua2000/sheetpilot/internal/worksheet"
)

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
