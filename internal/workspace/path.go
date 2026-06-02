package workspace

import (
	"errors"
	"fmt"
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
		outputDir:     strings.TrimSpace(outputDir),
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
	if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
		return "", errors.New("file path must stay inside workspace")
	}
	path := filepath.Join(w.root, cleaned)
	if !isWithin(w.root, path) {
		return "", errors.New("file path must stay inside workspace")
	}
	if err := ensureNoSymlinkEscape(w.root, path); err != nil {
		return "", err
	}
	return path, nil
}

func (w Workspace) OutputPath(name string) (string, string, error) {
	path, _, url, err := w.OutputFile(name)
	return path, url, err
}

func (w Workspace) OutputFile(name string) (string, string, string, error) {
	base := sanitizeBaseName(name)
	if base == "" {
		base = "workbook"
	}
	if strings.ToLower(filepath.Ext(base)) != ".xlsx" {
		base += ".xlsx"
	}
	dir, err := w.outputDirectory()
	if err != nil {
		return "", "", "", err
	}
	for attempt := 0; attempt < 1000; attempt++ {
		candidate := suffixName(base, attempt)
		rel := filepath.Join(dir, candidate)
		path, err := w.Resolve(rel)
		if err != nil {
			return "", "", "", err
		}
		if err := ensureRealDirectory(filepath.Dir(path)); err != nil {
			return "", "", "", err
		}
		reserved, err := reserveOutputPath(path)
		if err != nil {
			return "", "", "", err
		}
		if !reserved {
			continue
		}
		url := ""
		if w.publicBaseURL != "" {
			url = w.publicBaseURL + "/" + candidate
		}
		return path, filepath.ToSlash(rel), url, nil
	}
	return "", "", "", fmt.Errorf("no available output name for %q", base)
}

func (w Workspace) outputDirectory() (string, error) {
	dir := strings.TrimSpace(w.outputDir)
	if dir == "" {
		dir = "output"
	}
	if filepath.IsAbs(dir) {
		return "", errors.New("output directory must stay inside workspace")
	}
	cleaned := filepath.Clean(dir)
	if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
		return "", errors.New("output directory must be a workspace subdirectory")
	}
	return cleaned, nil
}

func ensureNoSymlinkEscape(root, path string) error {
	realRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return err
	}
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return err
	}
	current := realRoot
	for _, part := range strings.Split(rel, string(filepath.Separator)) {
		if part == "" || part == "." {
			continue
		}
		candidate := filepath.Join(current, part)
		info, err := os.Lstat(candidate)
		if os.IsNotExist(err) {
			return nil
		}
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			target, err := filepath.EvalSymlinks(candidate)
			if err != nil {
				return err
			}
			if !isWithin(realRoot, target) {
				return errors.New("file path must stay inside workspace")
			}
			current = target
			continue
		}
		current = candidate
	}
	return nil
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
	if err := os.Mkdir(path, 0o755); err != nil {
		if os.IsExist(err) {
			return ensureRealDirectory(path)
		}
		return err
	}
	return nil
}

func reserveOutputPath(path string) (bool, error) {
	info, err := os.Lstat(path)
	if err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			return false, errors.New("output file must not be a symlink")
		}
		if info.IsDir() {
			return false, errors.New("output file path is a directory")
		}
		return false, nil
	}
	if !os.IsNotExist(err) {
		return false, err
	}
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0o600)
	if os.IsExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, file.Close()
}

func isWithin(root, path string) bool {
	rel, err := filepath.Rel(root, path)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func sanitizeBaseName(name string) string {
	base := filepath.Base(strings.TrimSpace(name))
	ext := filepath.Ext(base)
	stem := strings.TrimSuffix(base, ext)
	stem = strings.ReplaceAll(stem, " ", "-")
	var out strings.Builder
	for _, r := range stem {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_' || r == '.' {
			out.WriteRune(r)
		}
	}
	cleaned := strings.Trim(out.String(), ".-")
	if cleaned == "" {
		return ""
	}
	if strings.EqualFold(ext, ".xlsx") {
		return cleaned + ".xlsx"
	}
	return cleaned
}

func suffixName(name string, attempt int) string {
	if attempt == 0 {
		return name
	}
	ext := filepath.Ext(name)
	stem := strings.TrimSuffix(name, ext)
	return fmt.Sprintf("%s-%d%s", stem, attempt+1, ext)
}
