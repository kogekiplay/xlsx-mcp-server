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

func TestRequirePrimitiveAcceptsTypedValues(t *testing.T) {
	for name, value := range map[string]any{"string": "ok", "number": 12.5, "bool": true} {
		t.Run(name, func(t *testing.T) {
			got, err := requirePrimitive(request(map[string]any{"value": value}), "value")
			if err != nil {
				t.Fatalf("requirePrimitive() error = %v", err)
			}
			if got != value {
				t.Fatalf("got = %#v", got)
			}
		})
	}
}

func TestRequirePrimitiveRejectsStructuredValue(t *testing.T) {
	_, err := requirePrimitive(request(map[string]any{"value": map[string]any{"nested": true}}), "value")
	if err == nil || !strings.Contains(err.Error(), "string, number, or boolean") {
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

func TestHandlersHonorCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	h := handlers{workspaceRoot: t.TempDir(), outputDir: "output"}
	result, err := h.inspectWorkbook(ctx, request(map[string]any{"file": "book.xlsx"}))
	if err != nil {
		t.Fatalf("inspectWorkbook err = %v", err)
	}
	if !result.IsError || !strings.Contains(resultText(t, result), "context canceled") {
		t.Fatalf("result = %#v", result)
	}
}

func TestOptionalRowsAcceptsPrimitiveCells(t *testing.T) {
	rows, err := optionalRows(request(map[string]any{"rows": []any{
		[]any{"日期", "销售额", "已确认"},
		[]any{"2025-01-05", 314955, true},
	}}), "rows")
	if err != nil {
		t.Fatalf("optionalRows() error = %v", err)
	}
	if len(rows) != 2 || rows[1][1] != 314955 || rows[1][2] != true {
		t.Fatalf("rows = %#v", rows)
	}
}

func TestOptionalRowsRejectsStructuredCells(t *testing.T) {
	_, err := optionalRows(request(map[string]any{"rows": []any{[]any{map[string]any{"bad": true}}}}), "rows")
	if err == nil || !strings.Contains(err.Error(), "rows must contain") {
		t.Fatalf("err = %v", err)
	}
}

func TestCreateWorkbookHandler(t *testing.T) {
	root := t.TempDir()
	h := handlers{workspaceRoot: root, outputDir: "output", publicBaseURL: "https://example.com/xlsx-download"}

	result, err := h.createWorkbook(context.Background(), request(map[string]any{
		"sheet":  "销售数据分析",
		"output": "sales.xlsx",
		"rows": []any{
			[]any{"日期", "产品名称", "销售数量", "销售额(元)"},
			[]any{"2025-01-05", "iPhone 16", 45, 314955},
		},
	}))
	if err != nil {
		t.Fatalf("createWorkbook err = %v", err)
	}
	if result.IsError {
		t.Fatalf("createWorkbook returned error result: %#v", result.Content)
	}
	f, err := excelize.OpenFile(filepath.Join(root, "output", "sales.xlsx"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	value, err := f.GetCellValue("销售数据分析", "B2")
	if err != nil {
		t.Fatal(err)
	}
	if value != "iPhone 16" {
		t.Fatalf("B2 = %q", value)
	}
	text := resultText(t, result)
	if !strings.Contains(text, "output/sales.xlsx") || strings.Contains(text, root) {
		t.Fatalf("result text = %s", text)
	}
	if !strings.Contains(text, "https://example.com/xlsx-download/sales.xlsx") {
		t.Fatalf("result text = %s", text)
	}
}

func TestCreateWorkbookHandlerWithoutRows(t *testing.T) {
	root := t.TempDir()
	h := handlers{workspaceRoot: root, outputDir: "output"}

	result, err := h.createWorkbook(context.Background(), request(map[string]any{"output": "blank.xlsx"}))
	if err != nil {
		t.Fatalf("createWorkbook err = %v", err)
	}
	if result.IsError {
		t.Fatalf("createWorkbook returned error result: %#v", result.Content)
	}
	f, err := excelize.OpenFile(filepath.Join(root, "output", "blank.xlsx"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if f.GetSheetName(0) != "Sheet1" {
		t.Fatalf("sheet = %q", f.GetSheetName(0))
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
	text := resultText(t, result)
	if !strings.Contains(text, "output/report.xlsx") || strings.Contains(text, root) {
		t.Fatalf("result text = %s", text)
	}
	if !strings.Contains(text, "https://example.com/xlsx-download/report.xlsx") {
		t.Fatalf("result text = %s", text)
	}
}

func TestWriteCellHandlerAcceptsNumericValue(t *testing.T) {
	root := t.TempDir()
	makeTestWorkbook(t, filepath.Join(root, "book.xlsx"))
	h := handlers{workspaceRoot: root, outputDir: "output"}

	result, err := h.writeCell(context.Background(), request(map[string]any{
		"file":   "book.xlsx",
		"sheet":  "Sheet1",
		"cell":   "C1",
		"value":  12.5,
		"output": "report.xlsx",
	}))
	if err != nil {
		t.Fatalf("writeCell err = %v", err)
	}
	if result.IsError {
		t.Fatalf("writeCell returned error result: %#v", result.Content)
	}
	f, err := excelize.OpenFile(filepath.Join(root, "output", "report.xlsx"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	value, err := f.GetCellValue("Sheet1", "C1")
	if err != nil {
		t.Fatal(err)
	}
	if value != "12.5" {
		t.Fatalf("C1 = %q", value)
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
	if !strings.Contains(resultText(t, result), "output/summary.xlsx") {
		t.Fatalf("result text = %s", resultText(t, result))
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
