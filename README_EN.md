# SheetPilot

![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?style=flat&logo=go&logoColor=white)
![MCP](https://img.shields.io/badge/MCP-Server-8B5CF6?style=flat&logo=modelcontextprotocol&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-green)

English | [中文](README.md)

A Go-based MCP (Model Context Protocol) server that enables AI to operate Excel files like a human user.

## Features

- **76 MCP Tools** — Covers workbook management, sheet operations, cell/range manipulation, data analysis, formulas, styles, charts, and more
- **AI First** — Tools extract data and return it to AI; AI understands and makes decisions
- **Batch First** — `batch_update` executes multiple operations in a single call, reducing MCP round-trips
- **Single Binary** — No dependencies, cross-platform (Linux / macOS / Windows)

## Quick Start

### Install

**Option 1: go install (recommended)**

```bash
go install github.com/yuhua2000/sheetpilot@latest
```

**Option 2: Download pre-built binary**

Download from [Releases](https://github.com/yuhua2000/sheetpilot/releases):

| Platform | Filename |
|----------|----------|
| Linux x86_64 | `sheetpilot-linux-amd64` |
| Linux ARM64 | `sheetpilot-linux-arm64` |
| macOS Intel | `sheetpilot-darwin-amd64` |
| macOS Apple Silicon | `sheetpilot-darwin-arm64` |
| Windows x86_64 | `sheetpilot-windows-amd64.exe` |
| Windows ARM64 | `sheetpilot-windows-arm64.exe` |

Set executable permission (Linux/macOS):

```bash
chmod +x sheetpilot-*
mv sheetpilot-* /usr/local/bin/sheetpilot
```

**Option 3: Build from source**

```bash
git clone https://github.com/yuhua2000/sheetpilot.git
cd excelMcp
go build -o sheetpilot .
```

### Run

```bash
# Stdio mode (default, for Claude Code and other MCP clients)
./sheetpilot serve

# SSE mode
./sheetpilot serve --transport sse --addr :8080
```

### Configure Claude Code

Create `.mcp.json` in your project root:

```json
{
  "mcpServers": {
    "excel": {
      "command": "/path/to/sheetpilot",
      "args": ["serve"]
    }
  }
}
```

## Tool List

| Category | Tools |
|----------|-------|
| Workbook | `open_workbook` `save_workbook` `save_workbook_as` `close_workbook` `list_workbooks` `get_workbook_info` |
| Sheet | `list_sheets` `create_sheet` `delete_sheet` `rename_sheet` `copy_sheet` `set_active_sheet` `get_sheet_info` |
| Cell | `get_cell` `set_cell` `clear_cell` `get_range` `set_range` `append_rows` |
| Row/Col | `insert_rows` `delete_rows` `insert_columns` `delete_columns` `copy_range` `move_range` |
| Analysis | `table_info` `column_types` `sheet_overview` |
| Data Ops | `sort_table` `filter_rows` `group_by` `deduplicate` `split_sheet` `merge_sheets` `cleanup_sheet` |
| Cleanup | `add_computed_column` `fill_missing_values` `replace_values` `find_replace` |
| Formula | `get_formula` `set_formula` `fill_formula_column` |
| Style | `set_style` `set_number_format` `auto_fit_columns` `auto_fit_rows` `format_as_table` |
| Conditional | `conditional_formatting` `freeze_panes` `add_filter` |
| Chart | `create_chart` `auto_chart` |
| Merge | `merge_cells` `unmerge_cells` `get_merged_cells` |
| View | `hide_sheet` `show_sheet` `hide_rows` `show_rows` `hide_columns` `show_columns` |
| Size | `set_row_height` `set_col_width` |
| Protection | `protect_sheet` `unprotect_sheet` |
| Print | `set_print_area` `set_header_footer` |
| Named | `set_defined_name` |
| Comment/Link | `add_comment` `add_hyperlink` `set_data_validation` |
| Import/Export | `export_csv` `import_csv` `export_json` |
| Batch | `batch_update` |

## Architecture

```
main.go → cmd/ (cobra CLI)
              └── mcp_serve.go → internal/mcp/server.go
                                      ├── handlers_*.go
                                      └── helpers.go
internal/
├── workbook/    # Workbook lifecycle management (thread-safe)
├── worksheet/   # Sheet CRUD
├── rangeop/     # Cell & range operations
├── dataops/     # Data processing (sort, filter, group, dedupe, etc.)
├── analysis/    # Table detection, column type inference
├── formula/     # Formula read/write
├── style/       # Styling, conditional formatting
├── chart/       # Chart creation and recommendation
├── view/        # View control, protection, print
└── io/          # CSV/JSON import/export
```

## Development

```bash
# Build
go build -o sheetpilot .

# Run all tests
go test ./...

# Run tests for a single package
go test ./internal/dataops/

# Run a single test
go test ./internal/workbook/ -run TestManager_Open

# Print version
./sheetpilot version
```

## Dependencies

- [excelize](https://github.com/xuri/excelize) — Excel read/write engine
- [mcp-go](https://github.com/mark3labs/mcp-go) — MCP protocol SDK
- [cobra](https://github.com/spf13/cobra) — CLI framework
- [testify](https://github.com/stretchr/testify) — Test assertions

## License

[MIT License](LICENSE)
