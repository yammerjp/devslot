package main

import (
	"fmt"
	"os"
	"path/filepath"
)

const devslotYamlTemplate = `version: 1
repositories:
  # - https://github.com/example/app1
  # - https://github.com/example/app2
`

const gitignoreTemplate = `# devslot managed directories
repos/
slots/
.devslot.lock
`

const hookScriptTemplate = `#!/bin/bash
# This hook receives the following environment variables:
# DEVSLOT_ROOT       - Path to devslot project root
# DEVSLOT_SLOT_NAME  - Name of the target slot
# DEVSLOT_SLOT_DIR   - Path to the slot directory
# DEVSLOT_REPOS_DIR  - Path to the bare repositories

# Add your custom logic here
`

func (cmd *BoilerplateCmd) Run(ctx *Context) error {
	// Resolve target directory
	targetDir := cmd.Dir
	if targetDir == "" {
		return fmt.Errorf("directory argument is required")
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create subdirectories
	dirs := []string{"hooks", "repos", "slots"}
	for _, dir := range dirs {
		path := filepath.Join(targetDir, dir)
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("failed to create %s directory: %w", dir, err)
		}
	}

	// Create devslot.yaml
	yamlPath := filepath.Join(targetDir, "devslot.yaml")
	if err := os.WriteFile(yamlPath, []byte(devslotYamlTemplate), 0644); err != nil {
		return fmt.Errorf("failed to create devslot.yaml: %w", err)
	}

	// Create .gitignore
	gitignorePath := filepath.Join(targetDir, ".gitignore")
	if err := os.WriteFile(gitignorePath, []byte(gitignoreTemplate), 0644); err != nil {
		return fmt.Errorf("failed to create .gitignore: %w", err)
	}

	// Create hook scripts
	hooks := []string{"post-create", "pre-destroy", "post-reload"}
	for _, hook := range hooks {
		hookPath := filepath.Join(targetDir, "hooks", hook)
		if err := os.WriteFile(hookPath, []byte(hookScriptTemplate), 0755); err != nil {
			return fmt.Errorf("failed to create hook %s: %w", hook, err)
		}
	}

	return nil
}