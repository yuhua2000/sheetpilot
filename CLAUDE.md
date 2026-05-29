# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

SheetPilot is a Go-based MCP (Model Context Protocol) server that enables AI to operate Excel files. It exposes ~76 MCP tools covering workbook management, sheet operations, cell/range manipulation, data analysis, formulas, styles, charts, and more.

Binary name: `sheetpilot`

## Build & Run

```bash
# Build
go build -o sheetpilot .

# Run tests (all)
go test ./...

# Run tests for a single package
go test ./internal/workbook/
go test ./internal/dataops/
go test ./internal/mcp/

# Run a single test
go test ./internal/workbook/ -run TestManager_Open

# Start MCP server (stdio mode, default)
./sheetpilot serve

# Start MCP server (SSE mode)
./sheetpilot serve --transport sse --addr :8080

# Print version
./sheetpilot version
```

## Architecture

```
main.go → cmd/ (cobra CLI)
              └── mcp_serve.go → internal/mcp/server.go (MCP server, tool registration)
                                      ├── handlers_*.go (MCP tool handlers, split by domain)
                                      └── helpers.go (shared MCP helpers)

internal/
├── workbook/    # Workbook lifecycle: open, save, close, metadata. Thread-safe Manager with RWMutex.
├── worksheet/   # Sheet CRUD: list, create, delete, rename, copy, info
├── rangeop/     # Cell & range ops: get/set cell, get/set range, insert/delete rows/cols, copy/move, merge cells
├── dataops/     # Data processing: sort, filter, group_by, deduplicate, split/merge sheets, cleanup
├── analysis/    # Table detection, column type inference, sheet overview
├── formula/     # Formula read/write/fill
├── style/       # Cell styling, conditional formatting, freeze panes, auto-filter, table formatting
├── chart/       # Chart creation and auto-recommendation
├── view/        # Show/hide sheets/rows/cols, row height/col width, protection, print area, named ranges
└── io/          # CSV/JSON import/export
```

### Key Design Decisions

- **Workbook Manager** (`internal/workbook/manager.go`): Central registry of open workbooks. Each workbook gets an auto-incremented ID (`wb_1`, `wb_2`, ...). All MCP handlers look up the workbook by ID via `s.manager.Get(id)`.
- **Column reference resolution**: `internal/dataops` supports both column letters (A, B) and column names (header text) via `resolveColumnRef()`. Functions like `SortTable`, `FilterRows`, `GroupBy`, `Deduplicate` accept either form.
- **MCP handler pattern**: Each handler extracts params via `req.RequireString()`, gets the workbook from manager, delegates to the internal package, returns `mcpResult()` or `mcpError()`.
- **Batch operations**: `batch_update` tool accepts a JSON array of operations and executes them sequentially, re-dispatching to individual tool handlers via `getToolHandler()`.

### Key Dependencies

- `github.com/xuri/excelize/v2` — Excel read/write engine
- `github.com/mark3labs/mcp-go` — MCP protocol SDK
- `github.com/spf13/cobra` — CLI framework
- `github.com/stretchr/testify` — Test assertions

### Tool Registration

Tools are registered in `internal/mcp/server.go` via `registerTools()`, which delegates to phase-specific registration methods. Handler implementations are split across multiple files:
- `handlers_workbook.go` — open/save/close/list/info
- `handlers_sheet.go` — sheet CRUD
- `handlers_range.go` — cell/range/row/col operations
- `handlers_dataops.go` — data processing
- `handlers_analysis.go` — table info, column types, overview
- `handlers_formula_style.go` — formula and basic style tools
- `handlers_advanced_style.go` — set_style, conditional_formatting, format_as_table, freeze_panes, add_filter
- `handlers_chart.go` — create_chart, auto_chart
- `handlers_merge.go` — merge/unmerge cells
- `handlers_batch.go` — batch_update
- `handlers_view.go` — show/hide, dimensions, protection, print, named ranges
- `handlers_io.go` — CSV/JSON import/export

## Testing Pattern

Tests use `testify/require` and create temporary Excel files via `t.TempDir()`. Example:

```go
func TestSomething(t *testing.T) {
    m := workbook.NewManager()
    tmpDir := t.TempDir()
    path := filepath.Join(tmpDir, "test.xlsx")
    wb, err := m.Open(path)
    require.NoError(t, err)
    // ... test operations on wb.File ...
}
```
