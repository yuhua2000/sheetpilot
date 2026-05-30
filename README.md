# SheetPilot

![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?style=flat&logo=go&logoColor=white)
![MCP](https://img.shields.io/badge/MCP-Server-8B5CF6?style=flat&logo=modelcontextprotocol&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-green)

[English](README_EN.md) | 中文

基于 Go 的 MCP (Model Context Protocol) 服务器，让 AI 像熟练用户一样操作 Excel 文件。

## 功能特性

- **76 个 MCP 工具** — 覆盖工作簿管理、Sheet 操作、单元格/区域操作、数据分析、公式、样式、图表等
- **AI First** — 工具负责提取数据返回给 AI，由 AI 理解和决策
- **批量优先** — 支持 `batch_update` 一次执行多个操作，减少 MCP 往返
- **单二进制** — 无需依赖，跨平台运行（Linux / macOS / Windows）

## 快速开始

### 安装

```bash
git clone https://github.com/yuhua2000/excelMcp.git
cd excelMcp
go build -o sheetpilot .
```

### 运行

```bash
# stdio 模式（默认，适用于 Claude Code 等 MCP 客户端）
./sheetpilot serve

# SSE 模式
./sheetpilot serve --transport sse --addr :8080
```

### 配置 Claude Code

在项目根目录创建 `.mcp.json`：

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

## 工具列表

| 分类 | 工具 |
|------|------|
| 工作簿 | `open_workbook` `save_workbook` `save_workbook_as` `close_workbook` `list_workbooks` `get_workbook_info` |
| Sheet | `list_sheets` `create_sheet` `delete_sheet` `rename_sheet` `copy_sheet` `set_active_sheet` `get_sheet_info` |
| 单元格 | `get_cell` `set_cell` `clear_cell` `get_range` `set_range` `append_rows` |
| 行列 | `insert_rows` `delete_rows` `insert_columns` `delete_columns` `copy_range` `move_range` |
| 数据分析 | `table_info` `column_types` `sheet_overview` |
| 数据处理 | `sort_table` `filter_rows` `group_by` `deduplicate` `split_sheet` `merge_sheets` `cleanup_sheet` |
| 数据清洗 | `add_computed_column` `fill_missing_values` `replace_values` `find_replace` |
| 公式 | `get_formula` `set_formula` `fill_formula_column` |
| 样式 | `set_style` `set_number_format` `auto_fit_columns` `auto_fit_rows` `format_as_table` |
| 条件格式 | `conditional_formatting` `freeze_panes` `add_filter` |
| 图表 | `create_chart` `auto_chart` |
| 合并单元格 | `merge_cells` `unmerge_cells` `get_merged_cells` |
| 视图 | `hide_sheet` `show_sheet` `hide_rows` `show_rows` `hide_columns` `show_columns` |
| 尺寸 | `set_row_height` `set_col_width` |
| 保护 | `protect_sheet` `unprotect_sheet` |
| 打印 | `set_print_area` `set_header_footer` |
| 命名 | `set_defined_name` |
| 批注/链接 | `add_comment` `add_hyperlink` `set_data_validation` |
| 导入导出 | `export_csv` `import_csv` `export_json` |
| 批量 | `batch_update` |

## 项目结构

```
main.go → cmd/ (cobra CLI)
              └── mcp_serve.go → internal/mcp/server.go
                                      ├── handlers_*.go
                                      └── helpers.go
internal/
├── workbook/    # 工作簿生命周期管理（线程安全）
├── worksheet/   # Sheet CRUD
├── rangeop/     # 单元格 & 区域操作
├── dataops/     # 数据处理（排序、筛选、分组、去重等）
├── analysis/    # 表格识别、列类型推断
├── formula/     # 公式读写
├── style/       # 样式、条件格式
├── chart/       # 图表创建与推荐
├── view/        # 视图控制、保护、打印
└── io/          # CSV/JSON 导入导出
```

## 开发

```bash
# 构建
go build -o sheetpilot .

# 运行全部测试
go test ./...

# 运行单个包的测试
go test ./internal/dataops/

# 运行单个测试
go test ./internal/workbook/ -run TestManager_Open

# 打印版本
./sheetpilot version
```

## 依赖

- [excelize](https://github.com/xuri/excelize) — Excel 读写引擎
- [mcp-go](https://github.com/mark3labs/mcp-go) — MCP 协议 SDK
- [cobra](https://github.com/spf13/cobra) — CLI 框架
- [testify](https://github.com/stretchr/testify) — 测试断言

## 许可证

[MIT License](LICENSE)
