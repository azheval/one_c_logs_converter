package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// LogDir represents a directory to be monitored.
type LogDir struct {
	Path    string `json:"path"`
	Enabled bool   `json:"enabled"`
}

// Project represents a project with its log directories.
type Project struct {
	Name            string   `json:"name"`
	LogDirs         []LogDir `json:"log_dirs"`
	DeleteProcessed bool     `json:"delete_processed"`
}

// Config represents the application's configuration.
type Config struct {
	LogLevel  string    `json:"log_level"`
	AppLogDir string    `json:"app_log_dir"`
	OutputDir string    `json:"output_dir"`
	Projects  []Project `json:"projects"`
}

// Load reads the configuration file from the given path, resolves relative paths,
// and returns a Config object.
func Load(path string) (*Config, error) {
	configFile, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer configFile.Close()

	configBytes, err := io.ReadAll(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(configBytes, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Resolve all paths relative to the config file's directory.
	configDir := filepath.Dir(path)

	if !filepath.IsAbs(cfg.AppLogDir) {
		cfg.AppLogDir = filepath.Join(configDir, cfg.AppLogDir)
	}

	if !filepath.IsAbs(cfg.OutputDir) {
		cfg.OutputDir = filepath.Join(configDir, cfg.OutputDir)
	}

	for i := range cfg.Projects {
		for j := range cfg.Projects[i].LogDirs {
			dir := &cfg.Projects[i].LogDirs[j]
			if !filepath.IsAbs(dir.Path) {
				dir.Path = filepath.Join(configDir, dir.Path)
			}
		}
	}

	return &cfg, nil
}
