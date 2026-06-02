package workspace

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveAllowsWorkspaceRelativePath(t *testing.T) {
	root := t.TempDir()
	ws := New(root, "output", "https://example.com/xlsx-download")

	path, err := ws.Resolve("input/book.xlsx")
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	want := filepath.Join(root, "input", "book.xlsx")
	if path != want {
		t.Fatalf("path = %q, want %q", path, want)
	}
}

func TestResolveRejectsTraversal(t *testing.T) {
	root := t.TempDir()
	ws := New(root, "output", "")

	if _, err := ws.Resolve("../secret.xlsx"); err == nil {
		t.Fatal("Resolve() accepted path traversal")
	}
}

func TestResolveRejectsAbsolutePath(t *testing.T) {
	root := t.TempDir()
	ws := New(root, "output", "")

	if _, err := ws.Resolve("/private/secret.xlsx"); err == nil {
		t.Fatal("Resolve() accepted absolute path")
	}
}

func TestOutputPathRejectsSymlinkedOutputDirectory(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(root, "output")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	ws := New(root, "output", "")

	if _, _, err := ws.OutputPath("report.xlsx"); err == nil {
		t.Fatal("OutputPath() accepted symlinked output directory")
	}
}

func TestOutputPathRejectsIntermediateSymlinkedDirectory(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(root, "output")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	ws := New(root, "output/reports", "")

	if _, _, err := ws.OutputPath("report.xlsx"); err == nil {
		t.Fatal("OutputPath() accepted intermediate symlinked directory")
	}
}

func TestOutputPathCreatesXLSXName(t *testing.T) {
	root := t.TempDir()
	ws := New(root, "output", "https://example.com/xlsx-download")

	path, url, err := ws.OutputPath("report")
	if err != nil {
		t.Fatalf("OutputPath() error = %v", err)
	}
	if filepath.Base(path) != "report.xlsx" {
		t.Fatalf("output basename = %q", filepath.Base(path))
	}
	if url != "https://example.com/xlsx-download/report.xlsx" {
		t.Fatalf("url = %q", url)
	}
	if _, err := os.Stat(filepath.Dir(path)); err != nil {
		t.Fatalf("output directory was not created: %v", err)
	}
}

func TestOutputPathSanitizesName(t *testing.T) {
	root := t.TempDir()
	ws := New(root, "output", "")

	path, _, err := ws.OutputPath("../bad name.xlsx")
	if err != nil {
		t.Fatalf("OutputPath() error = %v", err)
	}
	if filepath.Base(path) != "bad-name.xlsx" {
		t.Fatalf("output basename = %q", filepath.Base(path))
	}
}
