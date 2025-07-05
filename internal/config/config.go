package config

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the devslot.yaml configuration
type Config struct {
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
		return nil, err
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