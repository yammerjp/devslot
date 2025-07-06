package command

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yammerjp/devslot/internal/testutil"
)

func TestCreateCmd_Run(t *testing.T) {

	tests := []struct {
		name         string
		slotName     string
		branch       string
		setupFunc    func(t *testing.T, projectRoot string) error
		wantErr      bool
		errContains  string
		validateFunc func(t *testing.T, projectRoot string)
	}{
		{
			name:     "basic slot creation",
			slotName: "test-slot",
			setupFunc: func(t *testing.T, projectRoot string) error {
				// Create devslot.yaml
				yamlContent := `version: 1
repositories:
  - name: repo1
    url: https://github.com/example/repo1.git
`
				testutil.CreateFile(t, filepath.Join(projectRoot, "devslot.yaml"), yamlContent)

				// Create a bare repository
				repo1Path := filepath.Join(projectRoot, "repos", "repo1.git")
				if err := os.MkdirAll(filepath.Dir(repo1Path), 0755); err != nil {
					return err
				}
				testutil.InitBareRepo(t, repo1Path)

				// Set git config for test
				cmd := exec.Command("git", "config", "--global", "user.email", "test@example.com")
				_ = cmd.Run() // Ignore error in test setup

				return nil
			},
			wantErr: false,
			validateFunc: func(t *testing.T, projectRoot string) {
				slotPath := filepath.Join(projectRoot, "slots", "test-slot")
				if _, err := os.Stat(slotPath); os.IsNotExist(err) {
					t.Error("expected slot directory to exist")
				}

				// Check if worktree was created
				worktreePath := filepath.Join(slotPath, "repo1")
				if _, err := os.Stat(worktreePath); os.IsNotExist(err) {
					t.Error("expected worktree to exist")
				}

				// Check if we're on the expected branch
				// The branch should be devslot/test/test-slot
				cmd := exec.Command("git", "-C", worktreePath, "branch", "--show-current")
				output, err := cmd.Output()
				if err != nil {
					t.Errorf("failed to get current branch: %v", err)
				}
				branch := strings.TrimSpace(string(output))
				// Should be "devslot/test/test-slot" based on git email
				if !strings.Contains(branch, "test-slot") || !strings.Contains(branch, "devslot/test/") {
					t.Errorf("expected branch to be 'devslot/test/test-slot', got %q", branch)
				}
			},
		},
		{
			name:     "slot with specific branch",
			slotName: "feature-slot",
			branch:   "feature-branch",
			setupFunc: func(t *testing.T, projectRoot string) error {
				// Create devslot.yaml
				yamlContent := `version: 1
repositories:
  - name: repo1
    url: https://github.com/example/repo1.git
`
				testutil.CreateFile(t, filepath.Join(projectRoot, "devslot.yaml"), yamlContent)

				// Create a bare repository with a branch
				repo1Path := filepath.Join(projectRoot, "repos", "repo1.git")
				if err := os.MkdirAll(filepath.Dir(repo1Path), 0755); err != nil {
					return err
				}
				testutil.InitBareRepo(t, repo1Path)
				return nil
			},
			wantErr: false,
			validateFunc: func(t *testing.T, projectRoot string) {
				slotPath := filepath.Join(projectRoot, "slots", "feature-slot")
				if _, err := os.Stat(slotPath); os.IsNotExist(err) {
					t.Error("expected slot directory to exist")
				}
			},
		},
		{
			name:     "slot already exists",
			slotName: "existing-slot",
			setupFunc: func(t *testing.T, projectRoot string) error {
				// Create devslot.yaml
				yamlContent := `version: 1
repositories: []
`
				testutil.CreateFile(t, filepath.Join(projectRoot, "devslot.yaml"), yamlContent)

				// Create existing slot
				slotPath := filepath.Join(projectRoot, "slots", "existing-slot")
				if err := os.MkdirAll(slotPath, 0755); err != nil {
					return err
				}
				return nil
			},
			wantErr:     true,
			errContains: "already exists",
		},
		{
			name:     "missing bare repository",
			slotName: "missing-repo-slot",
			setupFunc: func(t *testing.T, projectRoot string) error {
				// Create devslot.yaml with non-existent repository
				yamlContent := `version: 1
repositories:
  - name: missing
    url: https://github.com/example/missing.git
`
				testutil.CreateFile(t, filepath.Join(projectRoot, "devslot.yaml"), yamlContent)
				return nil
			},
			wantErr:     true,
			errContains: "does not exist",
		},
		{
			name:     "invalid slot name",
			slotName: "invalid/name",
			setupFunc: func(t *testing.T, projectRoot string) error {
				// Create devslot.yaml
				yamlContent := `version: 1
repositories: []
`
				testutil.CreateFile(t, filepath.Join(projectRoot, "devslot.yaml"), yamlContent)
				return nil
			},
			wantErr:     true,
			errContains: "cannot contain path separators",
		},
		{
			name:        "not in devslot project",
			slotName:    "test-slot",
			setupFunc:   func(t *testing.T, projectRoot string) error { return nil },
			wantErr:     true,
			errContains: "devslot.yaml not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projectRoot := testutil.TempDir(t)

			// Change to project directory
			defer testutil.Chdir(t, projectRoot)()

			if tt.setupFunc != nil {
				if err := tt.setupFunc(t, projectRoot); err != nil {
					t.Fatalf("setup failed: %v", err)
				}
			}

			var buf bytes.Buffer
			cmd := &CreateCmd{
				SlotName: tt.slotName,
				Branch:   tt.branch,
			}
			ctx := &Context{Writer: &buf, Logger: nil}

			err := cmd.Run(ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateCmd.Run() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil && tt.errContains != "" {
				if !contains(err.Error(), tt.errContains) {
					t.Errorf("CreateCmd.Run() error = %v, want error containing %q", err, tt.errContains)
				}
			}

			if tt.validateFunc != nil && !tt.wantErr {
				tt.validateFunc(t, projectRoot)
			}
		})
	}
}
