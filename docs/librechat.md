# LibreChat setup

This example adds `xlsx-mcp-server` to LibreChat as a stdio MCP server. It assumes the LibreChat API container can execute the server binary and can access a mounted workbook workspace.

## Suggested container paths

- Server binary: `/app/mcp/xlsx-mcp-server`
- Workspace: `/app/xlsx-workspace`
- Output directory: `/app/xlsx-workspace/output`

## `librechat.yaml`

```yaml
mcpServers:
  xlsx-editor:
    command: /app/mcp/xlsx-mcp-server
    env:
      XLSX_WORKSPACE_ROOT: /app/xlsx-workspace
      XLSX_OUTPUT_DIR: output
      XLSX_PUBLIC_BASE_URL: https://example.com/xlsx-download
```

## Download route

Expose only the workspace output directory through your reverse proxy or static file server. Do not expose the workspace root because uploaded input files may contain private data.

## User workflow

1. Put an XLSX file into the configured workspace.
2. Ask the AI client to inspect or edit the workbook with `xlsx-editor`.
3. The server writes generated workbooks to the output directory.
4. The AI response includes `download_url` when `XLSX_PUBLIC_BASE_URL` is configured.
