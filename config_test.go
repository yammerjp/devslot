package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name        string
		yamlContent string
		want        *Config
		wantErr     bool
		errContains string
	}{
		{
			name: "valid config",
			yamlContent: `version: 1
repositories:
  - https://github.com/example/repo1
  - https://github.com/example/repo2
`,
			want: &Config{
				Version: 1,
				Repositories: []string{
					"https://github.com/example/repo1",
					"https://github.com/example/repo2",
				},
			},
			wantErr: false,
		},
		{
			name: "empty repositories",
			yamlContent: `version: 1
repositories: []
`,
			want: &Config{
				Version:      1,
				Repositories: []string{},
			},
			wantErr: false,
		},
		{
			name: "unsupported version",
			yamlContent: `version: 2
repositories: []
`,
			wantErr:     true,
			errContains: "unsupported config version: 2",
		},
		{
			name:        "invalid YAML",
			yamlContent: `version: 1 repositories: [`,
			wantErr:     true,
			errContains: "failed to parse YAML",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "devslot.yaml")
			
			err := os.WriteFile(configPath, []byte(tt.yamlContent), 0644)
			if err != nil {
				t.Fatalf("failed to write test config: %v", err)
			}

			got, err := LoadConfig(configPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errContains != "" {
				if err == nil || !contains(err.Error(), tt.errContains) {
					t.Errorf("LoadConfig() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			if !tt.wantErr {
				if got.Version != tt.want.Version {
					t.Errorf("LoadConfig() Version = %v, want %v", got.Version, tt.want.Version)
				}
				if len(got.Repositories) != len(tt.want.Repositories) {
					t.Errorf("LoadConfig() Repositories length = %v, want %v", len(got.Repositories), len(tt.want.Repositories))
				} else {
					for i, repo := range got.Repositories {
						if repo != tt.want.Repositories[i] {
							t.Errorf("LoadConfig() Repositories[%d] = %v, want %v", i, repo, tt.want.Repositories[i])
						}
					}
				}
			}
		})
	}
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	_, err := LoadConfig("/nonexistent/path/devslot.yaml")
	if err == nil {
		t.Error("LoadConfig() expected error for nonexistent file")
	}
	if !contains(err.Error(), "failed to read config file") {
		t.Errorf("LoadConfig() error = %v, want error containing 'failed to read config file'", err)
	}
}

func TestFindProjectRoot(t *testing.T) {
	t.Run("finds devslot.yaml in current directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "devslot.yaml")
		err := os.WriteFile(configPath, []byte("version: 1\n"), 0644)
		if err != nil {
			t.Fatalf("failed to create test config: %v", err)
		}

		oldWd, err := os.Getwd()
		if err != nil {
			t.Fatalf("failed to get current directory: %v", err)
		}
		defer os.Chdir(oldWd)

		err = os.Chdir(tmpDir)
		if err != nil {
			t.Fatalf("failed to change directory: %v", err)
		}

		got, err := FindProjectRoot()
		if err != nil {
			t.Errorf("FindProjectRoot() error = %v, want nil", err)
		}
		// Resolve symlinks for comparison (macOS temp dirs use symlinks)
		gotResolved, _ := filepath.EvalSymlinks(got)
		tmpDirResolved, _ := filepath.EvalSymlinks(tmpDir)
		if gotResolved != tmpDirResolved {
			t.Errorf("FindProjectRoot() = %v, want %v", got, tmpDir)
		}
	})

	t.Run("finds devslot.yaml in parent directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "devslot.yaml")
		err := os.WriteFile(configPath, []byte("version: 1\n"), 0644)
		if err != nil {
			t.Fatalf("failed to create test config: %v", err)
		}

		subDir := filepath.Join(tmpDir, "subdir")
		err = os.Mkdir(subDir, 0755)
		if err != nil {
			t.Fatalf("failed to create subdirectory: %v", err)
		}

		oldWd, err := os.Getwd()
		if err != nil {
			t.Fatalf("failed to get current directory: %v", err)
		}
		defer os.Chdir(oldWd)

		err = os.Chdir(subDir)
		if err != nil {
			t.Fatalf("failed to change directory: %v", err)
		}

		got, err := FindProjectRoot()
		if err != nil {
			t.Errorf("FindProjectRoot() error = %v, want nil", err)
		}
		// Resolve symlinks for comparison (macOS temp dirs use symlinks)
		gotResolved, _ := filepath.EvalSymlinks(got)
		tmpDirResolved, _ := filepath.EvalSymlinks(tmpDir)
		if gotResolved != tmpDirResolved {
			t.Errorf("FindProjectRoot() = %v, want %v", got, tmpDir)
		}
	})

	t.Run("returns error when devslot.yaml not found", func(t *testing.T) {
		tmpDir := t.TempDir()

		oldWd, err := os.Getwd()
		if err != nil {
			t.Fatalf("failed to get current directory: %v", err)
		}
		defer os.Chdir(oldWd)

		err = os.Chdir(tmpDir)
		if err != nil {
			t.Fatalf("failed to change directory: %v", err)
		}

		_, err = FindProjectRoot()
		if err == nil {
			t.Error("FindProjectRoot() expected error when devslot.yaml not found")
		}
		if !contains(err.Error(), "devslot.yaml not found") {
			t.Errorf("FindProjectRoot() error = %v, want error containing 'devslot.yaml not found'", err)
		}
	})
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr || len(s) > len(substr) && contains(s[1:], substr)
}