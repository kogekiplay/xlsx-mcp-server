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

func TestResolveAllowsDotDotPrefixFilename(t *testing.T) {
	root := t.TempDir()
	ws := New(root, "output", "")

	path, err := ws.Resolve("..backup.xlsx")
	if err != nil {
		t.Fatalf("Resolve() rejected safe filename: %v", err)
	}
	if filepath.Base(path) != "..backup.xlsx" {
		t.Fatalf("basename = %q", filepath.Base(path))
	}
}

func TestResolveRejectsSymlinkEscape(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	if err := os.WriteFile(filepath.Join(outside, "secret.xlsx"), []byte("secret"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "input")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	ws := New(root, "output", "")

	if _, err := ws.Resolve("input/secret.xlsx"); err == nil {
		t.Fatal("Resolve() accepted symlink escape")
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

func TestOutputPathRejectsSymlinkedOutputFile(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "output"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(outside, "target.xlsx"), []byte("target"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(outside, "target.xlsx"), filepath.Join(root, "output", "report.xlsx")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	ws := New(root, "output", "")

	if _, _, err := ws.OutputPath("report.xlsx"); err == nil {
		t.Fatal("OutputPath() accepted symlinked output file")
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

func TestOutputFileReturnsWorkspaceRelativePath(t *testing.T) {
	root := t.TempDir()
	ws := New(root, "output", "https://example.com/xlsx-download")

	_, rel, url, err := ws.OutputFile("report.xlsx")
	if err != nil {
		t.Fatalf("OutputFile() error = %v", err)
	}
	if rel != "output/report.xlsx" {
		t.Fatalf("rel = %q", rel)
	}
	if url != "https://example.com/xlsx-download/report.xlsx" {
		t.Fatalf("url = %q", url)
	}
}

func TestOutputFileAvoidsExistingFileCollision(t *testing.T) {
	root := t.TempDir()
	outputDir := filepath.Join(root, "output")
	if err := os.Mkdir(outputDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(outputDir, "workbook.xlsx"), []byte("existing"), 0o600); err != nil {
		t.Fatal(err)
	}
	ws := New(root, "output", "")

	path, rel, _, err := ws.OutputFile("销售.xlsx")
	if err != nil {
		t.Fatalf("OutputFile() error = %v", err)
	}
	if filepath.Base(path) != "workbook-2.xlsx" || rel != "output/workbook-2.xlsx" {
		t.Fatalf("path=%q rel=%q", path, rel)
	}
}

func TestOutputFileRejectsWorkspaceRootOutputDirectory(t *testing.T) {
	root := t.TempDir()
	ws := New(root, ".", "")

	if _, _, _, err := ws.OutputFile("report.xlsx"); err == nil {
		t.Fatal("OutputFile() accepted workspace root as output directory")
	}
}
