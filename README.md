# xlsx-mcp-server

A generic MCP server for reading, editing, and saving XLSX workbooks with AI clients. It uses [Excelize](https://github.com/qax-os/excelize) through the Go module `github.com/xuri/excelize/v2` and serves tools over MCP stdio.

## Tools

- `inspect_workbook` — list workbook sheets.
- `read_range` — read all rows or a specific Excel range.
- `write_cell` — write one cell and save a new workbook.
- `add_sheet` — add a worksheet and save a new workbook.

## Safety model

The server only reads and writes files inside `XLSX_WORKSPACE_ROOT`. Output files are saved under `XLSX_OUTPUT_DIR` within that workspace. Set `XLSX_PUBLIC_BASE_URL` when your host exposes output files through a static download route.

## Environment

| Variable | Default | Description |
| --- | --- | --- |
| `XLSX_WORKSPACE_ROOT` | `workspace` | Root directory for input and output workbooks. |
| `XLSX_OUTPUT_DIR` | `output` | Output subdirectory inside the workspace. |
| `XLSX_PUBLIC_BASE_URL` | empty | Optional public base URL for generated output files. |

## Build

```bash
go build ./cmd/xlsx-mcp-server
```

## Test

```bash
go test ./...
```

## MCP client example

```json
{
  "mcpServers": {
    "xlsx-editor": {
      "command": "/path/to/xlsx-mcp-server",
      "env": {
        "XLSX_WORKSPACE_ROOT": "/srv/xlsx-workspace",
        "XLSX_OUTPUT_DIR": "output",
        "XLSX_PUBLIC_BASE_URL": "https://example.com/xlsx-download"
      }
    }
  }
}
```
