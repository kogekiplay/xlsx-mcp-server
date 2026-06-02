package main

import (
	"fmt"
	"os"

	"github.com/kogekiplay/xlsx-mcp-server/internal/config"
	"github.com/kogekiplay/xlsx-mcp-server/internal/mcpserver"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}
	s := mcpserver.New(mcpserver.Options{
		WorkspaceRoot: cfg.WorkspaceRoot,
		OutputDir:     cfg.OutputDir,
		PublicBaseURL: cfg.PublicBaseURL,
	})
	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}
