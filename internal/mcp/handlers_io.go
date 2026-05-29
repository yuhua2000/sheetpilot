package mcp

import (
	"context"
	"log/slog"

	"github.com/mark3labs/mcp-go/mcp"
	csvio "github.com/yuhua2000/sheetpilot/internal/io"
)

func (s *Server) registerIOTools() {
	s.mcpSrv.AddTool(
		mcp.NewTool("export_csv",
			mcp.WithDescription("Export sheet to CSV file"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
			mcp.WithString("sheet", mcp.Required(), mcp.Description("Sheet name")),
			mcp.WithString("output_path", mcp.Required(), mcp.Description("Output CSV file path")),
		),
		s.handleExportCSV,
	)

	s.mcpSrv.AddTool(
		mcp.NewTool("import_csv",
			mcp.WithDescription("Import CSV file into a sheet"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
			mcp.WithString("csv_path", mcp.Required(), mcp.Description("CSV file path")),
			mcp.WithString("sheet", mcp.Required(), mcp.Description("Target sheet name")),
		),
		s.handleImportCSV,
	)

	s.mcpSrv.AddTool(
		mcp.NewTool("export_json",
			mcp.WithDescription("Export sheet to JSON format"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
			mcp.WithString("sheet", mcp.Required(), mcp.Description("Sheet name")),
		),
		s.handleExportJSON,
	)
}

func (s *Server) handleExportCSV(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("workbook_id")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	sheet, err := req.RequireString("sheet")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	outputPath, err := req.RequireString("output_path")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	slog.Info("exporting CSV", "workbook_id", id, "sheet", sheet, "output", outputPath)
	wb, err := s.manager.Get(id)
	if err != nil {
		return mcpError(err.Error()), nil
	}

	if err := csvio.ExportCSV(wb.File, sheet, outputPath); err != nil {
		slog.Error("export CSV failed", "workbook_id", id, "sheet", sheet, "error", err)
		return mcpError(err.Error()), nil
	}

	return mcpResult("ok"), nil
}

func (s *Server) handleImportCSV(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("workbook_id")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	csvPath, err := req.RequireString("csv_path")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	sheet, err := req.RequireString("sheet")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	slog.Info("importing CSV", "workbook_id", id, "csv", csvPath, "sheet", sheet)
	wb, err := s.manager.Get(id)
	if err != nil {
		return mcpError(err.Error()), nil
	}

	if err := csvio.ImportCSV(wb.File, csvPath, sheet); err != nil {
		slog.Error("import CSV failed", "workbook_id", id, "csv", csvPath, "error", err)
		return mcpError(err.Error()), nil
	}

	return mcpResult("ok"), nil
}

func (s *Server) handleExportJSON(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("workbook_id")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	sheet, err := req.RequireString("sheet")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	slog.Debug("exporting JSON", "workbook_id", id, "sheet", sheet)
	wb, err := s.manager.Get(id)
	if err != nil {
		return mcpError(err.Error()), nil
	}

	result, err := csvio.ExportJSON(wb.File, sheet)
	if err != nil {
		slog.Error("export JSON failed", "workbook_id", id, "sheet", sheet, "error", err)
		return mcpError(err.Error()), nil
	}

	return mcpResult(result), nil
}
