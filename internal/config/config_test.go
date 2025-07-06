package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yammerjp/devslot/internal/testutil"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name        string
		yamlContent string
		wantErr     bool
		wantRepos   int
	}{
		{
			name: "valid config",
			yamlContent: `version: 1
repositories:
  - name: repo1.git
    url: https://github.com/example/repo1.git
  - name: repo2.git
    url: https://github.com/example/repo2.git
`,
			wantErr:   false,
			wantRepos: 2,
		},
		{
			name: "empty config",
			yamlContent: `version: 1
repositories: []
`,
			wantErr:   false,
			wantRepos: 0,
		},
		{
			name:        "invalid yaml",
			yamlContent: `invalid: [yaml content`,
			wantErr:     true,
			wantRepos:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := testutil.TempDir(t)
			testutil.CreateFile(t, filepath.Join(tempDir, "devslot.yaml"), tt.yamlContent)

			cfg, err := Load(tempDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(cfg.Repositories) != tt.wantRepos {
					t.Errorf("Load() got %d repositories, want %d", len(cfg.Repositories), tt.wantRepos)
				}
			}
		})
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	tempDir := testutil.TempDir(t)
	_, err := Load(tempDir)
	if err == nil {
		t.Error("Load() expected error for missing file, got nil")
	}
}

func TestFindProjectRoot(t *testing.T) {
	// Create a nested directory structure
	tempDir := testutil.TempDir(t)
	projectRoot := filepath.Join(tempDir, "project")
	nestedDir := filepath.Join(projectRoot, "deeply", "nested", "directory")

	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatalf("Failed to create nested directory: %v", err)
	}

	// Create devslot.yaml in project root
	testutil.CreateFile(t, filepath.Join(projectRoot, "devslot.yaml"), "version: 1\nrepositories: []")

	tests := []struct {
		name      string
		startPath string
		wantPath  string
		wantErr   bool
	}{
		{
			name:      "from project root",
			startPath: projectRoot,
			wantPath:  projectRoot,
			wantErr:   false,
		},
		{
			name:      "from nested directory",
			startPath: nestedDir,
			wantPath:  projectRoot,
			wantErr:   false,
		},
		{
			name:      "not in project",
			startPath: tempDir,
			wantPath:  "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPath, err := FindProjectRoot(tt.startPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindProjectRoot() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if gotPath != tt.wantPath {
				t.Errorf("FindProjectRoot() = %v, want %v", gotPath, tt.wantPath)
			}
		})
	}
}

func TestRepository(t *testing.T) {
	// Test Repository struct
	repo := Repository{
		Name: "test-repo",
		URL:  "https://github.com/test/repo.git",
	}

	if repo.Name != "test-repo" {
		t.Errorf("Repository.Name = %v, want test-repo", repo.Name)
	}

	if repo.URL != "https://github.com/test/repo.git" {
		t.Errorf("Repository.URL = %v, want https://github.com/test/repo.git", repo.URL)
	}
}