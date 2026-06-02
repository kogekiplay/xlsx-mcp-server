package config

import "testing"

func TestLoadDefaults(t *testing.T) {
	t.Setenv("XLSX_WORKSPACE_ROOT", "")
	t.Setenv("XLSX_OUTPUT_DIR", "")
	t.Setenv("XLSX_PUBLIC_BASE_URL", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.WorkspaceRoot != "workspace" {
		t.Fatalf("WorkspaceRoot = %q, want workspace", cfg.WorkspaceRoot)
	}
	if cfg.OutputDir != "output" {
		t.Fatalf("OutputDir = %q, want output", cfg.OutputDir)
	}
	if cfg.PublicBaseURL != "" {
		t.Fatalf("PublicBaseURL = %q, want empty", cfg.PublicBaseURL)
	}
}

func TestLoadEnvironmentOverrides(t *testing.T) {
	t.Setenv("XLSX_WORKSPACE_ROOT", "/srv/xlsx")
	t.Setenv("XLSX_OUTPUT_DIR", "generated")
	t.Setenv("XLSX_PUBLIC_BASE_URL", "https://example.com/xlsx-download")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.WorkspaceRoot != "/srv/xlsx" {
		t.Fatalf("WorkspaceRoot = %q", cfg.WorkspaceRoot)
	}
	if cfg.OutputDir != "generated" {
		t.Fatalf("OutputDir = %q", cfg.OutputDir)
	}
	if cfg.PublicBaseURL != "https://example.com/xlsx-download" {
		t.Fatalf("PublicBaseURL = %q", cfg.PublicBaseURL)
	}
}
