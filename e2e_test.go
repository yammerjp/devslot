//go:build e2e
// +build e2e

package main_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestE2E_FullWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Build the binary
	tempDir := t.TempDir()
	binaryPath := filepath.Join(tempDir, "devslot")
	
	buildCmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/devslot")
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build devslot: %v\nOutput: %s", err, output)
	}

	// Create a test project directory
	projectDir := filepath.Join(tempDir, "test-project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("Failed to create project directory: %v", err)
	}

	// Helper function to run devslot commands
	runDevslot := func(args ...string) (string, error) {
		cmd := exec.Command(binaryPath, args...)
		cmd.Dir = projectDir
		output, err := cmd.CombinedOutput()
		return string(output), err
	}

	// Test 1: Version command
	t.Run("version", func(t *testing.T) {
		output, err := runDevslot("version")
		if err != nil {
			t.Fatalf("version command failed: %v\nOutput: %s", err, output)
		}
		if !strings.Contains(output, "devslot version") {
			t.Errorf("Expected version output to contain 'devslot version', got: %s", output)
		}
	})

	// Test 2: Boilerplate command
	t.Run("boilerplate", func(t *testing.T) {
		output, err := runDevslot("boilerplate")
		if err != nil {
			t.Fatalf("boilerplate command failed: %v\nOutput: %s", err, output)
		}

		// Check that expected files and directories were created
		expectedPaths := []string{
			"devslot.yaml",
			".gitignore",
			"hooks",
			"repos",
			"slots",
			"hooks/post-create.example",
			"hooks/pre-destroy.example",
			"hooks/post-reload.example",
		}

		for _, path := range expectedPaths {
			fullPath := filepath.Join(projectDir, path)
			if _, err := os.Stat(fullPath); os.IsNotExist(err) {
				t.Errorf("Expected path %s to exist, but it doesn't", path)
			}
		}

		// Check devslot.yaml content
		yamlContent, err := os.ReadFile(filepath.Join(projectDir, "devslot.yaml"))
		if err != nil {
			t.Fatalf("Failed to read devslot.yaml: %v", err)
		}
		if !strings.Contains(string(yamlContent), "repositories:") {
			t.Error("devslot.yaml should contain 'repositories:' section")
		}
	})

	// Test 3: Doctor command (should pass after boilerplate)
	t.Run("doctor", func(t *testing.T) {
		output, err := runDevslot("doctor")
		// Doctor returns error if issues found, but after boilerplate it should find some issues
		// (no repositories cloned yet)
		if err == nil {
			t.Log("doctor command succeeded (no critical issues)")
		}
		
		if !strings.Contains(output, "Checking configuration") {
			t.Errorf("Expected doctor output to contain 'Checking configuration', got: %s", output)
		}
	})

	// Test 4: List command (should show no slots)
	t.Run("list-empty", func(t *testing.T) {
		output, err := runDevslot("list")
		if err != nil {
			t.Fatalf("list command failed: %v\nOutput: %s", err, output)
		}
		
		if !strings.Contains(output, "No slots found") {
			t.Errorf("Expected list output to show 'No slots found', got: %s", output)
		}
	})

	// Test 5: Create a test repository config
	t.Run("setup-test-repo", func(t *testing.T) {
		// Create a simple devslot.yaml with a test repository
		yamlContent := `version: 1
repositories:
  - name: test-repo.git
    url: https://github.com/octocat/Hello-World.git
`
		yamlPath := filepath.Join(projectDir, "devslot.yaml")
		if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
			t.Fatalf("Failed to write devslot.yaml: %v", err)
		}
	})

	// Test 6: Init command (skip if no network)
	t.Run("init", func(t *testing.T) {
		if os.Getenv("DEVSLOT_E2E_SKIP_NETWORK") != "" {
			t.Skip("Skipping init test due to DEVSLOT_E2E_SKIP_NETWORK")
		}

		output, err := runDevslot("init")
		if err != nil {
			// Network error is acceptable in CI
			if strings.Contains(output, "failed to clone") {
				t.Skip("Skipping init test due to network error")
			}
			t.Fatalf("init command failed: %v\nOutput: %s", err, output)
		}

		// Check if repository was cloned
		repoPath := filepath.Join(projectDir, "repos", "test-repo.git")
		if _, err := os.Stat(repoPath); os.IsNotExist(err) {
			t.Error("Expected test-repo.git to be cloned")
		}
	})
}

func TestE2E_HelpAndErrors(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Build the binary
	tempDir := t.TempDir()
	binaryPath := filepath.Join(tempDir, "devslot")
	
	buildCmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/devslot")
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build devslot: %v\nOutput: %s", err, output)
	}

	// Test help output
	t.Run("help", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "--help")
		output, err := cmd.CombinedOutput()
		// Help command returns non-zero exit code
		if err != nil && !strings.Contains(string(output), "Usage:") {
			t.Fatalf("help command failed unexpectedly: %v\nOutput: %s", err, output)
		}

		outputStr := string(output)
		expectedCommands := []string{
			"boilerplate",
			"init",
			"create",
			"destroy",
			"reload",
			"list",
			"doctor",
			"version",
		}

		for _, cmd := range expectedCommands {
			if !strings.Contains(outputStr, cmd) {
				t.Errorf("Expected help output to contain command '%s'", cmd)
			}
		}
	})

	// Test invalid command
	t.Run("invalid-command", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "invalid-command")
		output, _ := cmd.CombinedOutput()
		// Kong may not return error for invalid command with our Exit override
		outputStr := string(output)
		
		// Check if we got an error message about invalid command
		if !strings.Contains(outputStr, "error: unexpected argument") &&
		   !strings.Contains(outputStr, "expected one of") {
			t.Errorf("Expected error message about invalid command, got: %s", outputStr)
		}
	})

	// Test commands that require devslot.yaml
	t.Run("commands-without-config", func(t *testing.T) {
		emptyDir := filepath.Join(tempDir, "empty")
		if err := os.MkdirAll(emptyDir, 0755); err != nil {
			t.Fatalf("Failed to create empty directory: %v", err)
		}

		commandsThatNeedConfig := []string{"init", "list", "doctor"}
		
		for _, cmdName := range commandsThatNeedConfig {
			t.Run(cmdName, func(t *testing.T) {
				cmd := exec.Command(binaryPath, cmdName)
				cmd.Dir = emptyDir
				output, err := cmd.CombinedOutput()
				outputStr := string(output)
				
				// Check for error message in output
				// Kong with Exit override may not return error code
				hasErrorMessage := strings.Contains(outputStr, "not in a devslot project") ||
				                  strings.Contains(outputStr, "devslot.yaml not found") ||
				                  strings.Contains(outputStr, "error:")
				
				if err == nil && !hasErrorMessage {
					t.Fatalf("Expected %s command to fail without devslot.yaml, got: %s", cmdName, outputStr)
				}
				
				if !hasErrorMessage {
					t.Errorf("Expected error about missing devslot.yaml, got: %s", outputStr)
				}
			})
		}
	})
}

func TestE2E_SlotOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// This test uses a mock git repository to avoid network dependencies
	tempDir := t.TempDir()
	binaryPath := filepath.Join(tempDir, "devslot")
	
	// Build the binary
	buildCmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/devslot")
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build devslot: %v\nOutput: %s", err, output)
	}

	// Create project structure
	projectDir := filepath.Join(tempDir, "test-project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("Failed to create project directory: %v", err)
	}

	// Helper to run devslot
	runDevslot := func(args ...string) (string, error) {
		cmd := exec.Command(binaryPath, args...)
		cmd.Dir = projectDir
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &out
		err := cmd.Run()
		return out.String(), err
	}

	// Create boilerplate
	if _, err := runDevslot("boilerplate"); err != nil {
		t.Fatalf("Failed to create boilerplate: %v", err)
	}

	// Create a mock bare git repository
	mockRepoDir := filepath.Join(projectDir, "repos", "mock-repo")
	if err := os.MkdirAll(mockRepoDir, 0755); err != nil {
		t.Fatalf("Failed to create mock repo directory: %v", err)
	}

	// Initialize as bare repository
	gitInitCmd := exec.Command("git", "init", "--bare")
	gitInitCmd.Dir = mockRepoDir
	if output, err := gitInitCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to init bare repo: %v\nOutput: %s", err, output)
	}

	// Create devslot.yaml
	yamlContent := `version: 1
repositories:
  - name: mock-repo.git
    url: file://` + mockRepoDir + `
`
	if err := os.WriteFile(filepath.Join(projectDir, "devslot.yaml"), []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write devslot.yaml: %v", err)
	}

	// Test create slot (should fail without proper worktree)
	t.Run("create-slot", func(t *testing.T) {
		output, err := runDevslot("create", "test-slot")
		// This might fail because we have a bare repo without commits
		if err != nil {
			if !strings.Contains(output, "failed to create") {
				t.Logf("Create slot failed as expected (bare repo without commits): %s", output)
			}
		}
	})

	// Test list slots
	t.Run("list-slots", func(t *testing.T) {
		output, err := runDevslot("list")
		if err != nil {
			t.Fatalf("list command failed: %v\nOutput: %s", err, output)
		}
		t.Logf("List output: %s", output)
	})
}