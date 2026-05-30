package mcp

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/yuhua2000/sheetpilot/internal/chart"
)

func (s *Server) registerChartTools() {
	s.mcpSrv.AddTool(
		mcp.NewTool("create_chart",
			mcp.WithDescription("Create a chart (bar, line, pie, scatter)"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
			mcp.WithString("sheet", mcp.Required(), mcp.Description("Sheet name")),
			mcp.WithString("chart_type", mcp.Required(), mcp.Description("Chart type: bar, line, pie, scatter")),
			mcp.WithString("title", mcp.Required(), mcp.Description("Chart title")),
			mcp.WithString("x_axis", mcp.Required(), mcp.Description("X axis label")),
			mcp.WithString("y_axis", mcp.Required(), mcp.Description("Y axis label")),
			mcp.WithString("x_values", mcp.Required(), mcp.Description("X values range (e.g., A2:A10)")),
			mcp.WithString("y_values", mcp.Required(), mcp.Description("Y values range (e.g., B2:B10)")),
			mcp.WithString("series_name", mcp.Description("Series name")),
		),
		s.handleCreateChart,
	)

	s.mcpSrv.AddTool(
		mcp.NewTool("auto_chart",
			mcp.WithDescription("Get chart type recommendation based on data analysis"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
			mcp.WithString("sheet", mcp.Required(), mcp.Description("Sheet name")),
		),
		s.handleAutoChart,
	)
}

func (s *Server) handleCreateChart(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("workbook_id")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	sheet, err := req.RequireString("sheet")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	chartType, err := req.RequireString("chart_type")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	title, err := req.RequireString("title")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	xAxis, err := req.RequireString("x_axis")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	yAxis, err := req.RequireString("y_axis")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	xValues, err := req.RequireString("x_values")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	yValues, err := req.RequireString("y_values")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	seriesName := req.GetString("series_name", "Series 1")

	slog.Info("creating chart", "workbook_id", id, "sheet", sheet, "type", chartType)
	wb, err := s.manager.Get(id)
	if err != nil {
		return mcpError(err.Error()), nil
	}

	info := chart.ChartInfo{
		Type:  chart.ChartType(chartType),
		Title: title,
		XAxis: xAxis,
		YAxis: yAxis,
		Series: []chart.Series{
			{
				Name:    seriesName,
				XValues: xValues,
				YValues: yValues,
			},
		},
	}

	if err := chart.CreateChart(wb.File, sheet, info); err != nil {
		slog.Error("create chart failed", "workbook_id", id, "sheet", sheet, "error", err)
		return mcpError(err.Error()), nil
	}

	return mcpResult("ok"), nil
}

func (s *Server) handleAutoChart(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("workbook_id")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	sheet, err := req.RequireString("sheet")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	slog.Debug("analyzing chart recommendation", "workbook_id", id, "sheet", sheet)
	wb, err := s.manager.Get(id)
	if err != nil {
		return mcpError(err.Error()), nil
	}

	recommendation, err := chart.AutoChart(wb.File, sheet)
	if err != nil {
		slog.Error("auto chart failed", "workbook_id", id, "sheet", sheet, "error", err)
		return mcpError(err.Error()), nil
	}

	data, _ := json.Marshal(recommendation)
	return mcpResult(string(data)), nil
}
