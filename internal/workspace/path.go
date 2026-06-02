package workspace

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

type Workspace struct {
	root          string
	outputDir     string
	publicBaseURL string
}

func New(root, outputDir, publicBaseURL string) Workspace {
	return Workspace{
		root:          filepath.Clean(root),
		outputDir:     strings.Trim(outputDir, "/"),
		publicBaseURL: strings.TrimRight(publicBaseURL, "/"),
	}
}

func (w Workspace) Resolve(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", errors.New("file name is required")
	}
	if filepath.IsAbs(name) {
		return "", errors.New("file path must stay inside workspace")
	}
	cleaned := filepath.Clean(name)
	if cleaned == "." || strings.HasPrefix(cleaned, "..") {
		return "", errors.New("file path must stay inside workspace")
	}
	path := filepath.Join(w.root, cleaned)
	if !isWithin(w.root, path) {
		return "", errors.New("file path must stay inside workspace")
	}
	return path, nil
}

func (w Workspace) OutputPath(name string) (string, string, error) {
	base := sanitizeBaseName(name)
	if base == "" {
		base = "workbook"
	}
	if strings.ToLower(filepath.Ext(base)) != ".xlsx" {
		base += ".xlsx"
	}
	rel := filepath.Join(w.outputDir, base)
	path, err := w.Resolve(rel)
	if err != nil {
		return "", "", err
	}
	if err := ensureRealDirectory(filepath.Dir(path)); err != nil {
		return "", "", err
	}
	if w.publicBaseURL == "" {
		return path, "", nil
	}
	return path, w.publicBaseURL + "/" + base, nil
}

func ensureRealDirectory(path string) error {
	info, err := os.Lstat(path)
	if err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			return errors.New("output directory must not be a symlink")
		}
		if !info.IsDir() {
			return errors.New("output path parent is not a directory")
		}
		return nil
	}
	if !os.IsNotExist(err) {
		return err
	}
	parent := filepath.Dir(path)
	if parent != path {
		if err := ensureRealDirectory(parent); err != nil {
			return err
		}
	}
	return os.Mkdir(path, 0o755)
}

func isWithin(root, path string) bool {
	rel, err := filepath.Rel(root, path)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func sanitizeBaseName(name string) string {
	base := filepath.Base(strings.TrimSpace(name))
	base = strings.ReplaceAll(base, " ", "-")
	var out strings.Builder
	for _, r := range base {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_' || r == '.' {
			out.WriteRune(r)
		}
	}
	return strings.Trim(out.String(), ".-")
}
