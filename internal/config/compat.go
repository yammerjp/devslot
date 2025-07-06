package config

import (
	"github.com/yammerjp/devslot/internal/git"
)

// LegacyConfig represents the legacy configuration format with string array repositories
type LegacyConfig struct {
	Version      int      `yaml:"version"`
	Repositories []string `yaml:"repositories"`
}

// ConvertLegacyConfig converts a legacy config to the new format
func ConvertLegacyConfig(legacy *LegacyConfig) *Config {
	config := &Config{
		Version:      legacy.Version,
		Repositories: make([]Repository, 0, len(legacy.Repositories)),
	}

	for _, repoURL := range legacy.Repositories {
		name, _ := git.ParseRepoURL(repoURL)
		if name != "" {
			config.Repositories = append(config.Repositories, Repository{
				Name: name,
				URL:  repoURL,
			})
		}
	}

	return config
}