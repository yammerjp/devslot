package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
)

// Config represents the devslot.yaml configuration
type Config struct {
	Version      int          `yaml:"version"`
	Repositories []Repository `yaml:"repositories"`
}

// Repository represents a single repository in the configuration
type Repository struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
}

// Load reads and parses the devslot.yaml configuration file
func Load(rootPath string) (*Config, error) {
	configPath := filepath.Join(rootPath, "devslot.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Default version to 1 if not specified
	if config.Version == 0 {
		config.Version = 1
	}

	if config.Version != 1 {
		return nil, fmt.Errorf("unsupported config version: %d", config.Version)
	}

	return &config, nil
}

// FindProjectRoot searches for the project root containing devslot.yaml
func FindProjectRoot(startPath string) (string, error) {
	currentPath := startPath
	for {
		configPath := filepath.Join(currentPath, "devslot.yaml")
		if _, err := os.Stat(configPath); err == nil {
			return currentPath, nil
		}

		parent := filepath.Dir(currentPath)
		if parent == currentPath {
			return "", errors.New("devslot.yaml not found in any parent directory")
		}
		currentPath = parent
	}
}