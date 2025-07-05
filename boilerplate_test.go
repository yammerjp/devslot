package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBoilerplateCommand(t *testing.T) {
	tests := []struct {
		name          string
		dir           string
		shouldCreate  bool
		expectedFiles []string
		expectedDirs  []string
	}{
		{
			name:         "create in new directory",
			dir:          "testdata/new-project",
			shouldCreate: true,
			expectedFiles: []string{
				"devslot.yaml",
				".gitignore",
				"hooks/post-create",
				"hooks/pre-destroy",
				"hooks/post-reload",
			},
			expectedDirs: []string{
				"hooks",
				"repos",
				"slots",
			},
		},
		{
			name:         "create in current directory",
			dir:          ".",
			shouldCreate: false,
			expectedFiles: []string{
				"devslot.yaml",
				".gitignore",
				"hooks/post-create",
				"hooks/pre-destroy",
				"hooks/post-reload",
			},
			expectedDirs: []string{
				"hooks",
				"repos",
				"slots",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory for testing
			tempDir := t.TempDir()
			targetDir := filepath.Join(tempDir, tt.dir)

			// Run boilerplate command
			cmd := &BoilerplateCmd{Dir: targetDir}
			err := cmd.Run(&Context{})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Check if directory was created
			if tt.shouldCreate {
				if _, err := os.Stat(targetDir); os.IsNotExist(err) {
					t.Errorf("expected directory %s to be created", targetDir)
				}
			}

			// Check expected files
			for _, file := range tt.expectedFiles {
				path := filepath.Join(targetDir, file)
				info, err := os.Stat(path)
				if err != nil {
					t.Errorf("expected file %s to exist: %v", file, err)
					continue
				}
				
				// Check if hook files are executable
				if strings.HasPrefix(file, "hooks/") && !isExecutable(info) {
					t.Errorf("expected %s to be executable", file)
				}
			}

			// Check expected directories
			for _, dir := range tt.expectedDirs {
				path := filepath.Join(targetDir, dir)
				info, err := os.Stat(path)
				if err != nil {
					t.Errorf("expected directory %s to exist: %v", dir, err)
					continue
				}
				if !info.IsDir() {
					t.Errorf("expected %s to be a directory", dir)
				}
			}

			// Check devslot.yaml content
			yamlPath := filepath.Join(targetDir, "devslot.yaml")
			content, err := os.ReadFile(yamlPath)
			if err != nil {
				t.Fatalf("failed to read devslot.yaml: %v", err)
			}
			if !strings.Contains(string(content), "version: 1") {
				t.Errorf("expected devslot.yaml to contain 'version: 1'")
			}

			// Check .gitignore content
			gitignorePath := filepath.Join(targetDir, ".gitignore")
			content, err = os.ReadFile(gitignorePath)
			if err != nil {
				t.Fatalf("failed to read .gitignore: %v", err)
			}
			if !strings.Contains(string(content), "repos/") || !strings.Contains(string(content), "slots/") {
				t.Errorf("expected .gitignore to contain 'repos/' and 'slots/'")
			}
		})
	}
}

func isExecutable(info os.FileInfo) bool {
	return info.Mode()&0111 != 0
}

func TestBoilerplateCommand_Errors(t *testing.T) {
	tests := []struct {
		name        string
		dir         string
		expectError string
	}{
		{
			name:        "empty directory",
			dir:         "",
			expectError: "directory argument is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &BoilerplateCmd{Dir: tt.dir}
			err := cmd.Run(&Context{})
			if err == nil {
				t.Errorf("expected error, got nil")
			} else if !strings.Contains(err.Error(), tt.expectError) {
				t.Errorf("expected error containing %q, got %q", tt.expectError, err.Error())
			}
		})
	}
}