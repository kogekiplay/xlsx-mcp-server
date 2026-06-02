package mcpserver

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/xuri/excelize/v2"
)

func TestNewServerRegistersTools(t *testing.T) {
	server := New(Options{WorkspaceRoot: t.TempDir(), OutputDir: "output"})
	if server == nil {
		t.Fatal("New() returned nil")
	}
}

func TestToolResultText(t *testing.T) {
	result := textResult(map[string]string{"ok": "true"})
	if result == nil || len(result.Content) == 0 {
		t.Fatal("textResult returned empty content")
	}
}

func TestRequireStringRejectsMissing(t *testing.T) {
	request := mcp.CallToolRequest{}
	_, err := requireString(request, "file")
	if err == nil || !strings.Contains(err.Error(), "file") {
		t.Fatalf("err = %v", err)
	}
}

func TestHandlersRejectMissingFile(t *testing.T) {
	h := handlers{workspaceRoot: t.TempDir(), outputDir: "output"}
	result, err := h.inspectWorkbook(context.Background(), mcp.CallToolRequest{})
	if err != nil {
		t.Fatalf("inspectWorkbook err = %v", err)
	}
	if !result.IsError {
		t.Fatal("inspectWorkbook accepted missing file")
	}
}

func TestInspectWorkbookHandler(t *testing.T) {
	root := t.TempDir()
	makeTestWorkbook(t, filepath.Join(root, "book.xlsx"))
	h := handlers{workspaceRoot: root, outputDir: "output"}

	result, err := h.inspectWorkbook(context.Background(), request(map[string]any{"file": "book.xlsx"}))
	if err != nil {
		t.Fatalf("inspectWorkbook err = %v", err)
	}
	if result.IsError {
		t.Fatalf("inspectWorkbook returned error result: %#v", result.Content)
	}
	text := resultText(t, result)
	if !strings.Contains(text, "Sheet1") {
		t.Fatalf("result text = %s", text)
	}
}

func TestReadRangeHandler(t *testing.T) {
	root := t.TempDir()
	makeTestWorkbook(t, filepath.Join(root, "book.xlsx"))
	h := handlers{workspaceRoot: root, outputDir: "output"}

	result, err := h.readRange(context.Background(), request(map[string]any{
		"file":  "book.xlsx",
		"sheet": "Sheet1",
		"range": "A1:B2",
	}))
	if err != nil {
		t.Fatalf("readRange err = %v", err)
	}
	if result.IsError {
		t.Fatalf("readRange returned error result: %#v", result.Content)
	}
	if !strings.Contains(resultText(t, result), "Alice") {
		t.Fatalf("result text = %s", resultText(t, result))
	}
}

func TestWriteCellHandler(t *testing.T) {
	root := t.TempDir()
	makeTestWorkbook(t, filepath.Join(root, "book.xlsx"))
	h := handlers{workspaceRoot: root, outputDir: "output", publicBaseURL: "https://example.com/xlsx-download"}

	result, err := h.writeCell(context.Background(), request(map[string]any{
		"file":   "book.xlsx",
		"sheet":  "Sheet1",
		"cell":   "C1",
		"value":  "Status",
		"output": "report.xlsx",
	}))
	if err != nil {
		t.Fatalf("writeCell err = %v", err)
	}
	if result.IsError {
		t.Fatalf("writeCell returned error result: %#v", result.Content)
	}
	output := filepath.Join(root, "output", "report.xlsx")
	f, err := excelize.OpenFile(output)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	value, err := f.GetCellValue("Sheet1", "C1")
	if err != nil {
		t.Fatal(err)
	}
	if value != "Status" {
		t.Fatalf("C1 = %q", value)
	}
	if !strings.Contains(resultText(t, result), "https://example.com/xlsx-download/report.xlsx") {
		t.Fatalf("result text = %s", resultText(t, result))
	}
}

func TestAddSheetHandler(t *testing.T) {
	root := t.TempDir()
	makeTestWorkbook(t, filepath.Join(root, "book.xlsx"))
	h := handlers{workspaceRoot: root, outputDir: "output"}

	result, err := h.addSheet(context.Background(), request(map[string]any{
		"file":   "book.xlsx",
		"sheet":  "Summary",
		"output": "summary.xlsx",
	}))
	if err != nil {
		t.Fatalf("addSheet err = %v", err)
	}
	if result.IsError {
		t.Fatalf("addSheet returned error result: %#v", result.Content)
	}
	f, err := excelize.OpenFile(filepath.Join(root, "output", "summary.xlsx"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if index, err := f.GetSheetIndex("Summary"); err != nil || index < 0 {
		t.Fatalf("Summary sheet missing, index=%d err=%v", index, err)
	}
}

func request(args map[string]any) mcp.CallToolRequest {
	return mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: args}}
}

func resultText(t *testing.T, result *mcp.CallToolResult) string {
	t.Helper()
	data, err := json.Marshal(result.Content)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func makeTestWorkbook(t *testing.T, path string) {
	t.Helper()
	f := excelize.NewFile()
	defer f.Close()
	if err := f.SetCellValue("Sheet1", "A1", "Name"); err != nil {
		t.Fatal(err)
	}
	if err := f.SetCellValue("Sheet1", "B1", "Amount"); err != nil {
		t.Fatal(err)
	}
	if err := f.SetCellValue("Sheet1", "A2", "Alice"); err != nil {
		t.Fatal(err)
	}
	if err := f.SetCellValue("Sheet1", "B2", 42); err != nil {
		t.Fatal(err)
	}
	if err := f.SaveAs(path); err != nil {
		t.Fatal(err)
	}
}
