package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/mark3labs/mcp-go/mcp"
)

func (s *Server) registerBatchTools() {
	s.mcpSrv.AddTool(
		mcp.NewTool("batch_update",
			mcp.WithDescription("Execute multiple operations in a single call"),
			mcp.WithString("workbook_id", mcp.Required(), mcp.Description("Workbook ID")),
			mcp.WithString("operations", mcp.Required(), mcp.Description("JSON array of operations")),
		),
		s.handleBatchUpdate,
	)
}

// Operation represents a single operation in a batch.
type Operation struct {
	Tool   string         `json:"tool"`
	Params map[string]any `json:"params"`
}

func (s *Server) handleBatchUpdate(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("workbook_id")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	opsStr, err := req.RequireString("operations")
	if err != nil {
		return mcpError(err.Error()), nil
	}

	var ops []Operation
	if err := json.Unmarshal([]byte(opsStr), &ops); err != nil {
		return mcpError(fmt.Sprintf("invalid operations format: %v", err)), nil
	}

	slog.Info("batch update", "workbook_id", id, "operations", len(ops))

	// Add workbook_id to each operation
	for i := range ops {
		if ops[i].Params == nil {
			ops[i].Params = make(map[string]any)
		}
		ops[i].Params["workbook_id"] = id
	}

	// Execute operations
	results := make([]any, len(ops))
	for i, op := range ops {
		result, err := s.executeOperation(ctx, op)
		if err != nil {
			results[i] = map[string]any{
				"error": err.Error(),
			}
			slog.Error("batch operation failed", "tool", op.Tool, "error", err)
		} else {
			results[i] = result
		}
	}

	data, _ := json.Marshal(map[string]any{
		"completed": len(ops),
		"results":   results,
	})

	return mcpResult(string(data)), nil
}

// executeOperation executes a single operation.
func (s *Server) executeOperation(ctx context.Context, op Operation) (any, error) {
	// Create a mock request for the tool
	params := map[string]any{
		"name":      op.Tool,
		"arguments": op.Params,
	}

	paramsJSON, _ := json.Marshal(params)

	// Find and execute the tool handler
	handler := s.getToolHandler(op.Tool)
	if handler == nil {
		return nil, fmt.Errorf("unknown tool: %s", op.Tool)
	}

	// Create CallToolRequest
	var toolReq mcp.CallToolRequest
	json.Unmarshal(paramsJSON, &toolReq)

	// Execute
	result, err := handler(ctx, toolReq)
	if err != nil {
		return nil, err
	}

	// Extract text content
	if result != nil && len(result.Content) > 0 {
		if tc, ok := result.Content[0].(mcp.TextContent); ok {
			return tc.Text, nil
		}
	}

	return "ok", nil
}

// getToolHandler returns the handler for a tool.
func (s *Server) getToolHandler(name string) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	handlers := map[string]func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error){
		// Workbook
		"open_workbook":     s.handleOpenWorkbook,
		"save_workbook":     s.handleSaveWorkbook,
		"close_workbook":    s.handleCloseWorkbook,
		"get_workbook_info": s.handleGetWorkbookInfo,

		// Sheet
		"list_sheets":       s.handleListSheets,
		"create_sheet":      s.handleCreateSheet,
		"delete_sheet":      s.handleDeleteSheet,
		"rename_sheet":      s.handleRenameSheet,
		"copy_sheet":        s.handleCopySheet,
		"set_active_sheet":  s.handleSetActiveSheet,
		"get_sheet_info":    s.handleGetSheetInfo,

		// Cell
		"get_cell":   s.handleGetCell,
		"set_cell":   s.handleSetCell,
		"clear_cell": s.handleClearCell,

		// Range
		"get_range":        s.handleGetRange,
		"set_range":        s.handleSetRange,
		"append_rows":      s.handleAppendRows,
		"insert_rows":      s.handleInsertRows,
		"delete_rows":      s.handleDeleteRows,
		"insert_columns":   s.handleInsertCols,
		"delete_columns":   s.handleDeleteCols,
		"copy_range":       s.handleCopyRange,
		"move_range":       s.handleMoveRange,

		// Data ops
		"add_computed_column": s.handleAddComputedColumn,
		"sort_table":          s.handleSortTable,
		"filter_rows":         s.handleFilterRows,
		"group_by":            s.handleGroupBy,
		"fill_missing_values": s.handleFillMissingValues,
		"replace_values":      s.handleReplaceValues,
		"cleanup_sheet":       s.handleCleanupSheet,
		"deduplicate":         s.handleDeduplicate,
		"split_sheet":         s.handleSplitSheet,
		"merge_sheets":        s.handleMergeSheets,

		// Formula
		"get_formula":         s.handleGetFormula,
		"set_formula":         s.handleSetFormula,
		"fill_formula_column": s.handleFillFormulaColumn,

		// Style
		"auto_fit_columns":       s.handleAutoFitColumns,
		"auto_fit_rows":          s.handleAutoFitRows,
		"set_number_format":      s.handleSetNumberFormat,
		"set_style":              s.handleSetStyle,
		"conditional_formatting": s.handleConditionalFormatting,
		"format_as_table":        s.handleFormatAsTable,
		"freeze_panes":           s.handleFreezePanes,
		"add_filter":             s.handleAddFilter,

		// Common
		"find_replace":        s.handleFindReplace,
		"add_comment":         s.handleAddComment,
		"add_hyperlink":       s.handleAddHyperlink,
		"set_data_validation": s.handleSetDataValidation,

		// Merge cells
		"merge_cells":      s.handleMergeCells,
		"unmerge_cells":    s.handleUnmergeCells,
		"get_merged_cells": s.handleGetMergedCells,

		// Chart
		"create_chart": s.handleCreateChart,
		"auto_chart":   s.handleAutoChart,
	}

	return handlers[name]
}
