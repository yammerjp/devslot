package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestE2E runs end-to-end tests with actual git operations
func TestE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E tests in short mode")
	}

	// Check if git is available
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not found in PATH, skipping E2E tests")
	}

	t.Run("basic init flow", func(t *testing.T) {
		// Setup test environment
		testEnv := setupE2ETest(t)
		defer testEnv.cleanup()

		// Create test repositories
		repo1 := testEnv.createTestRepo("test-repo1")
		repo2 := testEnv.createTestRepo("test-repo2")

		// Create devslot.yaml
		yamlContent := fmt.Sprintf(`version: 1
repositories:
  - %s
  - %s
`, repo1, repo2)
		err := os.WriteFile(filepath.Join(testEnv.projectRoot, "devslot.yaml"), []byte(yamlContent), 0644)
		if err != nil {
			t.Fatalf("failed to create devslot.yaml: %v", err)
		}

		// Run init command
		output, err := testEnv.runDevslot("init")
		if err != nil {
			t.Fatalf("init command failed: %v\nOutput: %s", err, output)
		}

		// Verify repositories were cloned
		testEnv.assertBareRepoExists(t, "test-repo1.git")
		testEnv.assertBareRepoExists(t, "test-repo2.git")

		// Verify output contains expected messages
		if !strings.Contains(output, "Cloning") {
			t.Errorf("Expected output to contain 'Cloning', got: %s", output)
		}
		if !strings.Contains(output, "Init completed successfully") {
			t.Errorf("Expected output to contain 'Init completed successfully', got: %s", output)
		}
	})

	t.Run("skip existing repositories", func(t *testing.T) {
		testEnv := setupE2ETest(t)
		defer testEnv.cleanup()

		repo1 := testEnv.createTestRepo("test-repo1")

		// Create devslot.yaml
		yamlContent := fmt.Sprintf(`version: 1
repositories:
  - %s
`, repo1)
		err := os.WriteFile(filepath.Join(testEnv.projectRoot, "devslot.yaml"), []byte(yamlContent), 0644)
		if err != nil {
			t.Fatalf("failed to create devslot.yaml: %v", err)
		}

		// First init
		_, err = testEnv.runDevslot("init")
		if err != nil {
			t.Fatalf("first init failed: %v", err)
		}

		// Add a marker file to the cloned repo
		markerPath := filepath.Join(testEnv.projectRoot, "repos", "test-repo1.git", "MARKER")
		err = os.WriteFile(markerPath, []byte("test"), 0644)
		if err != nil {
			t.Fatalf("failed to create marker: %v", err)
		}

		// Second init
		output, err := testEnv.runDevslot("init")
		if err != nil {
			t.Fatalf("second init failed: %v", err)
		}

		// Verify repository was skipped
		if !strings.Contains(output, "already exists, skipping") {
			t.Errorf("Expected skip message, got: %s", output)
		}

		// Verify marker still exists
		if _, err := os.Stat(markerPath); os.IsNotExist(err) {
			t.Error("Repository was replaced when it should have been skipped")
		}
	})

	t.Run("--allow-delete removes unlisted repositories", func(t *testing.T) {
		testEnv := setupE2ETest(t)
		defer testEnv.cleanup()

		repo1 := testEnv.createTestRepo("test-repo1")
		repo2 := testEnv.createTestRepo("test-repo2")

		// Create devslot.yaml with both repos
		yamlContent := fmt.Sprintf(`version: 1
repositories:
  - %s
  - %s
`, repo1, repo2)
		err := os.WriteFile(filepath.Join(testEnv.projectRoot, "devslot.yaml"), []byte(yamlContent), 0644)
		if err != nil {
			t.Fatalf("failed to create devslot.yaml: %v", err)
		}

		// Initial init
		_, err = testEnv.runDevslot("init")
		if err != nil {
			t.Fatalf("initial init failed: %v", err)
		}

		// Verify both repos exist
		testEnv.assertBareRepoExists(t, "test-repo1.git")
		testEnv.assertBareRepoExists(t, "test-repo2.git")

		// Update devslot.yaml to only include repo1
		yamlContent = fmt.Sprintf(`version: 1
repositories:
  - %s
`, repo1)
		err = os.WriteFile(filepath.Join(testEnv.projectRoot, "devslot.yaml"), []byte(yamlContent), 0644)
		if err != nil {
			t.Fatalf("failed to update devslot.yaml: %v", err)
		}

		// Run init with --allow-delete
		output, err := testEnv.runDevslot("init", "--allow-delete")
		if err != nil {
			t.Fatalf("init --allow-delete failed: %v\nOutput: %s", err, output)
		}

		// Verify repo1 still exists and repo2 was removed
		testEnv.assertBareRepoExists(t, "test-repo1.git")
		testEnv.assertBareRepoNotExists(t, "test-repo2.git")

		// Verify output mentions removal
		if !strings.Contains(output, "Removing unlisted repository") {
			t.Errorf("Expected removal message, got: %s", output)
		}
	})

	t.Run("concurrent execution is prevented", func(t *testing.T) {
		testEnv := setupE2ETest(t)
		defer testEnv.cleanup()

		// Create a minimal devslot.yaml (no actual repos to speed up test)
		yamlContent := `version: 1
repositories: []
`
		err := os.WriteFile(filepath.Join(testEnv.projectRoot, "devslot.yaml"), []byte(yamlContent), 0644)
		if err != nil {
			t.Fatalf("failed to create devslot.yaml: %v", err)
		}

		// Change to project directory
		oldWd, err := os.Getwd()
		if err != nil {
			t.Fatalf("failed to get current directory: %v", err)
		}
		defer func() { _ = os.Chdir(oldWd) }()

		err = os.Chdir(testEnv.projectRoot)
		if err != nil {
			t.Fatalf("failed to change directory: %v", err)
		}

		// Acquire lock manually to simulate another process
		lock := NewFileLock(testEnv.projectRoot)
		if err := lock.Lock(); err != nil {
			t.Fatalf("failed to acquire lock: %v", err)
		}
		defer func() { _ = lock.Unlock() }()

		// Try to run init while lock is held using App directly
		var buf strings.Builder
		app := NewApp(&buf)
		err = app.Run([]string{"init"})

		// The command should fail due to lock
		if err == nil {
			t.Errorf("Init should have failed due to lock")
		}

		// Check that the error is about the lock
		if !strings.Contains(err.Error(), "another devslot process is already running") {
			t.Errorf("Expected lock error, got: %v", err)
		}
	})
}

// e2eTestEnv holds the test environment
type e2eTestEnv struct {
	t             *testing.T
	rootDir       string
	projectRoot   string
	testReposRoot string
	devslotBinary string
}

func setupE2ETest(t *testing.T) *e2eTestEnv {
	t.Helper()

	// Create temporary directory structure
	rootDir := t.TempDir()
	projectRoot := filepath.Join(rootDir, "project")
	testReposRoot := filepath.Join(rootDir, "test-repos")

	if err := os.MkdirAll(projectRoot, 0755); err != nil {
		t.Fatalf("failed to create project root: %v", err)
	}
	if err := os.MkdirAll(testReposRoot, 0755); err != nil {
		t.Fatalf("failed to create test repos root: %v", err)
	}

	// Build devslot binary
	devslotBinary := filepath.Join(rootDir, "devslot")
	cmd := exec.Command("go", "build", "-o", devslotBinary, ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build devslot: %v\nOutput: %s", err, output)
	}

	return &e2eTestEnv{
		t:             t,
		rootDir:       rootDir,
		projectRoot:   projectRoot,
		testReposRoot: testReposRoot,
		devslotBinary: devslotBinary,
	}
}

func (env *e2eTestEnv) cleanup() {
	// Cleanup is handled by t.TempDir()
}

func (env *e2eTestEnv) createTestRepo(name string) string {
	env.t.Helper()

	repoPath := filepath.Join(env.testReposRoot, name)
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		env.t.Fatalf("failed to create repo directory: %v", err)
	}

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		env.t.Fatalf("failed to init git repo: %v\nOutput: %s", err, output)
	}

	// Configure git user for the test repo
	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		env.t.Fatalf("failed to set git email: %v", err)
	}

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		env.t.Fatalf("failed to set git name: %v", err)
	}

	// Create initial commit
	testFile := filepath.Join(repoPath, "README.md")
	if err := os.WriteFile(testFile, []byte("# Test Repo\n"), 0644); err != nil {
		env.t.Fatalf("failed to create test file: %v", err)
	}

	cmd = exec.Command("git", "add", ".")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		env.t.Fatalf("failed to add files: %v", err)
	}

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		env.t.Fatalf("failed to commit: %v\nOutput: %s", err, output)
	}

	return repoPath
}

func (env *e2eTestEnv) runDevslot(args ...string) (string, error) {
	env.t.Helper()

	cmd := exec.Command(env.devslotBinary, args...)
	cmd.Dir = env.projectRoot

	output, err := cmd.CombinedOutput()
	return string(output), err
}

func (env *e2eTestEnv) assertBareRepoExists(t *testing.T, repoName string) {
	t.Helper()

	repoPath := filepath.Join(env.projectRoot, "repos", repoName)
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		t.Errorf("Expected bare repository %s to exist", repoName)
		return
	}

	// Verify it's a bare repository
	cmd := exec.Command("git", "config", "--get", "core.bare")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil || strings.TrimSpace(string(output)) != "true" {
		t.Errorf("Repository %s is not a bare repository", repoName)
	}
}

func (env *e2eTestEnv) assertBareRepoNotExists(t *testing.T, repoName string) {
	t.Helper()

	repoPath := filepath.Join(env.projectRoot, "repos", repoName)
	if _, err := os.Stat(repoPath); !os.IsNotExist(err) {
		t.Errorf("Expected bare repository %s to not exist", repoName)
	}
}
