package mcpserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/kogekiplay/xlsx-mcp-server/internal/workspace"
	"github.com/kogekiplay/xlsx-mcp-server/internal/xlsx"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type Options struct {
	WorkspaceRoot string
	OutputDir     string
	PublicBaseURL string
}

func New(opts Options) *server.MCPServer {
	s := server.NewMCPServer(
		"xlsx-mcp-server",
		"0.1.0",
		server.WithToolCapabilities(false),
	)
	h := handlers{workspaceRoot: opts.WorkspaceRoot, outputDir: opts.OutputDir, publicBaseURL: opts.PublicBaseURL}
	s.AddTool(mcp.NewTool(
		"inspect_workbook",
		mcp.WithDescription("List workbook sheets"),
		mcp.WithString("file", mcp.Required(), mcp.Description("Workspace-relative XLSX file path")),
	), h.inspectWorkbook)
	s.AddTool(mcp.NewTool(
		"read_range",
		mcp.WithDescription("Read a sheet range"),
		mcp.WithString("file", mcp.Required(), mcp.Description("Workspace-relative XLSX file path")),
		mcp.WithString("sheet", mcp.Required(), mcp.Description("Worksheet name")),
		mcp.WithString("range", mcp.Description("Excel range such as A1:D20 or C3; omit for bounded all-row read")),
	), h.readRange)
	s.AddTool(mcp.NewTool(
		"write_cell",
		mcp.WithDescription("Write one cell and save as a new XLSX"),
		mcp.WithString("file", mcp.Required(), mcp.Description("Workspace-relative XLSX file path")),
		mcp.WithString("sheet", mcp.Required(), mcp.Description("Worksheet name")),
		mcp.WithString("cell", mcp.Required(), mcp.Description("Cell address such as C1")),
		mcp.WithAny("value", mcp.Required(), mcp.Description("String, number, or boolean value to write")),
		mcp.WithString("output", mcp.Required(), mcp.Description("Output file name")),
	), h.writeCell)
	s.AddTool(mcp.NewTool(
		"add_sheet",
		mcp.WithDescription("Add a sheet and save as a new XLSX"),
		mcp.WithString("file", mcp.Required(), mcp.Description("Workspace-relative XLSX file path")),
		mcp.WithString("sheet", mcp.Required(), mcp.Description("Worksheet name to create")),
		mcp.WithString("output", mcp.Required(), mcp.Description("Output file name")),
	), h.addSheet)
	return s
}

type handlers struct {
	workspaceRoot string
	outputDir     string
	publicBaseURL string
}

func (h handlers) workspace() workspace.Workspace {
	return workspace.New(h.workspaceRoot, h.outputDir, h.publicBaseURL)
}

func (h handlers) inspectWorkbook(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if result := contextResult(ctx); result != nil {
		return result, nil
	}
	file, err := requireString(request, "file")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	path, err := h.workspace().Resolve(file)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if result := contextResult(ctx); result != nil {
		return result, nil
	}
	info, err := (xlsx.Service{}).Inspect(path)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return textResult(info), nil
}

func (h handlers) readRange(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if result := contextResult(ctx); result != nil {
		return result, nil
	}
	file, err := requireString(request, "file")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	sheet, err := requireString(request, "sheet")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	cellRange, err := optionalString(request, "range")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	path, err := h.workspace().Resolve(file)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if result := contextResult(ctx); result != nil {
		return result, nil
	}
	values, err := (xlsx.Service{}).ReadRange(path, sheet, cellRange)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return textResult(map[string]any{"values": values}), nil
}

func (h handlers) writeCell(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if result := contextResult(ctx); result != nil {
		return result, nil
	}
	file, err := requireString(request, "file")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	sheet, err := requireString(request, "sheet")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	cell, err := requireString(request, "cell")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	value, err := requirePrimitive(request, "value")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	output, err := requireString(request, "output")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	ws := h.workspace()
	inputPath, err := ws.Resolve(file)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	outputPath, outputFile, url, err := ws.OutputFile(output)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if result := contextResult(ctx); result != nil {
		_ = os.Remove(outputPath)
		return result, nil
	}
	if err := (xlsx.Service{}).WriteCell(inputPath, outputPath, sheet, cell, value); err != nil {
		_ = os.Remove(outputPath)
		return mcp.NewToolResultError(err.Error()), nil
	}
	if result := contextResult(ctx); result != nil {
		_ = os.Remove(outputPath)
		return result, nil
	}
	return textResult(map[string]string{"file": outputFile, "download_url": url}), nil
}

func (h handlers) addSheet(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if result := contextResult(ctx); result != nil {
		return result, nil
	}
	file, err := requireString(request, "file")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	sheet, err := requireString(request, "sheet")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	output, err := requireString(request, "output")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	ws := h.workspace()
	inputPath, err := ws.Resolve(file)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	outputPath, outputFile, url, err := ws.OutputFile(output)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if result := contextResult(ctx); result != nil {
		_ = os.Remove(outputPath)
		return result, nil
	}
	if err := (xlsx.Service{}).AddSheet(inputPath, outputPath, sheet); err != nil {
		_ = os.Remove(outputPath)
		return mcp.NewToolResultError(err.Error()), nil
	}
	if result := contextResult(ctx); result != nil {
		_ = os.Remove(outputPath)
		return result, nil
	}
	return textResult(map[string]string{"file": outputFile, "download_url": url}), nil
}

func requireString(request mcp.CallToolRequest, name string) (string, error) {
	value, err := request.RequireString(name)
	if err != nil {
		return "", fmt.Errorf("%s is required", name)
	}
	return value, nil
}

func requirePrimitive(request mcp.CallToolRequest, name string) (any, error) {
	value, ok := request.GetArguments()[name]
	if !ok || value == nil {
		return nil, fmt.Errorf("%s is required", name)
	}
	switch v := value.(type) {
	case string, bool, float32, float64, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return v, nil
	case json.Number:
		if i, err := v.Int64(); err == nil {
			return i, nil
		}
		f, err := v.Float64()
		if err != nil {
			return nil, errors.New(name + " must be a string, number, or boolean")
		}
		return f, nil
	default:
		return nil, errors.New(name + " must be a string, number, or boolean")
	}
}

func optionalString(request mcp.CallToolRequest, name string) (string, error) {
	value, ok := request.GetArguments()[name]
	if !ok || value == nil {
		return "", nil
	}
	str, ok := value.(string)
	if !ok {
		return "", errors.New(name + " must be a string")
	}
	return str, nil
}

func contextResult(ctx context.Context) *mcp.CallToolResult {
	if err := ctx.Err(); err != nil {
		return mcp.NewToolResultError(err.Error())
	}
	return nil
}

func textResult(value any) *mcp.CallToolResult {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(err.Error())
	}
	return mcp.NewToolResultText(string(data))
}
