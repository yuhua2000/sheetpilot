package mcp

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/mark3labs/mcp-go/mcp"
)

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
