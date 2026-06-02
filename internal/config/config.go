package config

import "os"

type Config struct {
	WorkspaceRoot string
	OutputDir     string
	PublicBaseURL string
}

func Load() (Config, error) {
	return Config{
		WorkspaceRoot: envOrDefault("XLSX_WORKSPACE_ROOT", "workspace"),
		OutputDir:     envOrDefault("XLSX_OUTPUT_DIR", "output"),
		PublicBaseURL: trimTrailingSlash(os.Getenv("XLSX_PUBLIC_BASE_URL")),
	}, nil
}

func envOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func trimTrailingSlash(value string) string {
	for len(value) > 1 && value[len(value)-1] == '/' {
		value = value[:len(value)-1]
	}
	return value
}
