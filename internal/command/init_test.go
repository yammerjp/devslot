package command

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"github.com/yammerjp/devslot/internal/testutil"
)

// execCommand is a wrapper for exec.Command to make testing easier
func execCommand(name string, arg ...string) *exec.Cmd {
	return exec.Command(name, arg...)
}

func TestInitCmd_Run(t *testing.T) {
	// Skip network-dependent tests in CI
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping network-dependent test in CI")
	}

	tests := []struct {
		name         string
		allowDelete  bool
		setupFunc    func(t *testing.T, projectRoot string)
		wantErr      bool
		errContains  string
		validateFunc func(t *testing.T, projectRoot string)
	}{
		{
			name: "no devslot.yaml",
			setupFunc: func(t *testing.T, projectRoot string) {
				// Don't create devslot.yaml
			},
			wantErr:     true,
			errContains: "devslot.yaml not found",
		},
		{
			name: "empty repository list",
			setupFunc: func(t *testing.T, projectRoot string) {
				yamlContent := `version: 1
repositories: []
`
				testutil.CreateFile(t, filepath.Join(projectRoot, "devslot.yaml"), yamlContent)
			},
			wantErr: false,
			validateFunc: func(t *testing.T, projectRoot string) {
				// Should complete successfully with no repositories
			},
		},
		{
			name: "local repository",
			setupFunc: func(t *testing.T, projectRoot string) {
				// Create a local git repository
				localRepoPath := filepath.Join(projectRoot, "local-repo")
				if err := os.MkdirAll(localRepoPath, 0755); err != nil {
					t.Fatalf("failed to create local repo directory: %v", err)
				}

				// Initialize as a git repository
				cmd := execCommand("git", "init")
				cmd.Dir = localRepoPath
				if output, err := cmd.CombinedOutput(); err != nil {
					t.Fatalf("failed to init local repo: %v\nOutput: %s", err, output)
				}

				// Create a commit
				testutil.CreateFile(t, filepath.Join(localRepoPath, "README.md"), "# Test Repo")
				cmd = execCommand("git", "add", ".")
				cmd.Dir = localRepoPath
				if output, err := cmd.CombinedOutput(); err != nil {
					t.Fatalf("failed to add files: %v\nOutput: %s", err, output)
				}

				cmd = execCommand("git", "commit", "-m", "Initial commit")
				cmd.Dir = localRepoPath
				if output, err := cmd.CombinedOutput(); err != nil {
					t.Fatalf("failed to commit: %v\nOutput: %s", err, output)
				}

				yamlContent := `version: 1
repositories:
  - name: local-repo
    url: ` + localRepoPath + `
`
				testutil.CreateFile(t, filepath.Join(projectRoot, "devslot.yaml"), yamlContent)
			},
			wantErr: false,
			validateFunc: func(t *testing.T, projectRoot string) {
				// Check that bare repository was created
				bareRepoPath := filepath.Join(projectRoot, "repos", "local-repo.git")
				if _, err := os.Stat(bareRepoPath); os.IsNotExist(err) {
					t.Error("expected local-repo.git to exist")
				}

				// Check it's a bare repository
				configPath := filepath.Join(bareRepoPath, "config")
				if _, err := os.Stat(configPath); os.IsNotExist(err) {
					t.Error("expected bare repository config to exist")
				}
			},
		},
		{
			name: "post-init hook is executed",
			setupFunc: func(t *testing.T, projectRoot string) {
				yamlContent := `version: 1
repositories: []
`
				testutil.CreateFile(t, filepath.Join(projectRoot, "devslot.yaml"), yamlContent)

				// Create post-init hook
				hooksDir := filepath.Join(projectRoot, "hooks")
				if err := os.MkdirAll(hooksDir, 0755); err != nil {
					t.Fatalf("failed to create hooks directory: %v", err)
				}

				hookScript := `#!/bin/bash
echo "POST-INIT-HOOK-EXECUTED" > "$DEVSLOT_ROOT/post-init-marker"
echo "$DEVSLOT_REPOSITORIES" > "$DEVSLOT_ROOT/post-init-repos"
`
				hookPath := filepath.Join(hooksDir, "post-init")
				if err := os.WriteFile(hookPath, []byte(hookScript), 0755); err != nil {
					t.Fatalf("failed to create post-init hook: %v", err)
				}
			},
			wantErr: false,
			validateFunc: func(t *testing.T, projectRoot string) {
				// Check that post-init hook was executed
				markerPath := filepath.Join(projectRoot, "post-init-marker")
				data, err := os.ReadFile(markerPath)
				if err != nil {
					t.Error("post-init hook was not executed (marker file not found)")
					return
				}
				if strings.TrimSpace(string(data)) != "POST-INIT-HOOK-EXECUTED" {
					t.Errorf("post-init hook marker has wrong content: %q", string(data))
				}

				// Check that repository names were passed
				reposPath := filepath.Join(projectRoot, "post-init-repos")
				reposData, err := os.ReadFile(reposPath)
				if err != nil {
					t.Error("post-init hook did not write repos file")
					return
				}
				// Empty repository list, so should be empty
				if strings.TrimSpace(string(reposData)) != "" {
					t.Errorf("expected empty repos list, got: %q", string(reposData))
				}
			},
		},
		{
			name:        "existing repository gets skipped",
			allowDelete: false,
			setupFunc: func(t *testing.T, projectRoot string) {
				yamlContent := `version: 1
repositories:
  - name: existing-repo
    url: https://example.com/repo.git
`
				testutil.CreateFile(t, filepath.Join(projectRoot, "devslot.yaml"), yamlContent)

				// Create existing bare repository
				existingRepoPath := filepath.Join(projectRoot, "repos", "existing-repo.git")
				cmd := execCommand("git", "init", "--bare", existingRepoPath)
				if output, err := cmd.CombinedOutput(); err != nil {
					t.Fatalf("failed to create bare repo: %v\nOutput: %s", err, output)
				}
			},
			wantErr: false,
			validateFunc: func(t *testing.T, projectRoot string) {
				// Repository should still exist and init should have skipped it
				existingRepoPath := filepath.Join(projectRoot, "repos", "existing-repo.git")
				if _, err := os.Stat(existingRepoPath); os.IsNotExist(err) {
					t.Error("expected existing-repo.git to still exist")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projectRoot := testutil.TempDir(t)
			reposDir := filepath.Join(projectRoot, "repos")
			if err := os.MkdirAll(reposDir, 0755); err != nil {
				t.Fatalf("failed to create repos directory: %v", err)
			}

			// Change to project directory
			defer testutil.Chdir(t, projectRoot)()

			if tt.setupFunc != nil {
				tt.setupFunc(t, projectRoot)
			}

			var buf bytes.Buffer
			cmd := &InitCmd{
				AllowDelete: tt.allowDelete,
			}
			ctx := &Context{Writer: &buf, Logger: nil}

			err := cmd.Run(ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("InitCmd.Run() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil && tt.errContains != "" {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("InitCmd.Run() error = %v, want error containing %q", err, tt.errContains)
				}
			}

			if tt.validateFunc != nil && !tt.wantErr {
				tt.validateFunc(t, projectRoot)
			}
		})
	}
}

func TestInitCmd_ConcurrentLock(t *testing.T) {
	// Skip in CI to avoid timing issues
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping concurrent lock test in CI")
	}

	projectRoot := testutil.TempDir(t)

	// Create a simple config with no repositories to avoid network issues
	yamlContent := `version: 1
repositories: []
`
	testutil.CreateFile(t, filepath.Join(projectRoot, "devslot.yaml"), yamlContent)

	// Change to project directory
	defer testutil.Chdir(t, projectRoot)()

	// Create lock manually to simulate concurrent access
	lockPath := filepath.Join(projectRoot, ".devslot.lock")
	lockFile, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		t.Fatalf("failed to create lock file: %v", err)
	}
	defer lockFile.Close()

	// Acquire lock on the file
	if err := syscall.Flock(int(lockFile.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		t.Fatalf("failed to acquire lock: %v", err)
	}
	defer func() {
		if err := syscall.Flock(int(lockFile.Fd()), syscall.LOCK_UN); err != nil {
			t.Logf("Warning: failed to unlock file: %v", err)
		}
	}()

	// Try to run init command while lock is held
	var buf bytes.Buffer
	cmd := &InitCmd{}
	ctx := &Context{Writer: &buf, Logger: nil}

	err = cmd.Run(ctx)
	if err == nil {
		t.Error("expected error due to lock contention, got nil")
	}
	if !strings.Contains(err.Error(), "another devslot process") && !strings.Contains(err.Error(), "lock is already held") {
		t.Errorf("expected lock error, got: %v", err)
	}
}
