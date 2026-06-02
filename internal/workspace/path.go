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
	if strings.TrimSpace(name) == "" {
		return "", errors.New("file name is required")
	}
	cleaned := filepath.Clean(strings.TrimPrefix(name, "/"))
	if cleaned == "." || strings.HasPrefix(cleaned, "..") || filepath.IsAbs(cleaned) {
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
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", "", err
	}
	if w.publicBaseURL == "" {
		return path, "", nil
	}
	return path, w.publicBaseURL + "/" + base, nil
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
