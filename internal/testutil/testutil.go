package testutil

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TempDir creates a temporary directory for testing
func TempDir(t *testing.T) string {
	t.Helper()

	dir, err := os.MkdirTemp("", "devslot-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	t.Cleanup(func() {
		os.RemoveAll(dir)
	})

	return dir
}

// CreateFile creates a file with the given content
func CreateFile(t *testing.T, path, content string) {
	t.Helper()

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("Failed to create directory %s: %v", dir, err)
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create file %s: %v", path, err)
	}
}

// CreateExecutable creates an executable file
func CreateExecutable(t *testing.T, path, content string) {
	t.Helper()

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("Failed to create directory %s: %v", dir, err)
	}

	if err := os.WriteFile(path, []byte(content), 0755); err != nil {
		t.Fatalf("Failed to create executable %s: %v", path, err)
	}
}

// FileExists checks if a file exists
func FileExists(t *testing.T, path string) bool {
	t.Helper()

	_, err := os.Stat(path)
	return err == nil
}

// DirExists checks if a directory exists
func DirExists(t *testing.T, path string) bool {
	t.Helper()

	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// ReadFile reads and returns file content
func ReadFile(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read file %s: %v", path, err)
	}

	return string(content)
}

// AssertFileContent checks if a file has the expected content
func AssertFileContent(t *testing.T, path, expected string) {
	t.Helper()

	actual := ReadFile(t, path)
	if actual != expected {
		t.Errorf("File %s content mismatch:\nExpected:\n%s\nActual:\n%s", path, expected, actual)
	}
}

// CreateProjectStructure creates a basic devslot project structure
func CreateProjectStructure(t *testing.T, root string) {
	t.Helper()

	// Create directories
	dirs := []string{
		filepath.Join(root, "hooks"),
		filepath.Join(root, "repos"),
		filepath.Join(root, "slots"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// Create devslot.yaml
	configContent := `version: 1
repositories:
  - name: example-repo.git
    url: https://github.com/example/repo.git
`
	CreateFile(t, filepath.Join(root, "devslot.yaml"), configContent)

	// Create .gitignore
	gitignoreContent := `repos/
slots/
`
	CreateFile(t, filepath.Join(root, ".gitignore"), gitignoreContent)
}

// Chdir changes directory and returns a cleanup function that logs errors
func Chdir(t *testing.T, dir string) func() {
	t.Helper()
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}
	return func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("Warning: failed to restore directory: %v", err)
		}
	}
}

// InitGitRepo initializes a git repository in the given directory
func InitGitRepo(t *testing.T, dir string) {
	t.Helper()

	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to init git repo: %v\nOutput: %s", err, output)
	}

	// Configure git user for commits
	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to config user.email: %v\nOutput: %s", err, output)
	}

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to config user.name: %v\nOutput: %s", err, output)
	}
}

// InitBareRepo initializes a bare git repository in the given directory
func InitBareRepo(t *testing.T, dir string) {
	t.Helper()

	cmd := exec.Command("git", "init", "--bare", dir)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to init bare repo: %v\nOutput: %s", err, output)
	}

	// Create a temporary non-bare repo to push initial commit
	tempDir := TempDir(t)
	InitGitRepo(t, tempDir)

	// Add the bare repo as origin
	cmd = exec.Command("git", "remote", "add", "origin", dir)
	cmd.Dir = tempDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to add remote: %v\nOutput: %s", err, output)
	}

	// Create initial commit
	CreateFile(t, filepath.Join(tempDir, "README.md"), "# Test Repository")
	cmd = exec.Command("git", "add", ".")
	cmd.Dir = tempDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to add files: %v\nOutput: %s", err, output)
	}

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = tempDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to commit: %v\nOutput: %s", err, output)
	}

	// Try pushing with main branch first (modern default)
	cmd = exec.Command("git", "branch", "-M", "main")
	cmd.Dir = tempDir
	_ = cmd.Run() // Ignore error as it might already be main

	// Push to bare repo with -u to set upstream
	cmd = exec.Command("git", "push", "-u", "origin", "main")
	cmd.Dir = tempDir
	if output, err := cmd.CombinedOutput(); err != nil {
		// Try with master branch if main fails
		cmd = exec.Command("git", "branch", "-M", "master")
		cmd.Dir = tempDir
		_ = cmd.Run() // Ignore error intentionally

		cmd = exec.Command("git", "push", "-u", "origin", "master")
		cmd.Dir = tempDir
		if output2, err2 := cmd.CombinedOutput(); err2 != nil {
			t.Fatalf("failed to push to bare repo: %v\nOutput: %s\n%s", err2, output, output2)
		}
	}

	// Set HEAD in the bare repo to point to the default branch
	cmd = exec.Command("git", "symbolic-ref", "HEAD", "refs/heads/main")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		// Try master if main fails
		cmd = exec.Command("git", "symbolic-ref", "HEAD", "refs/heads/master")
		cmd.Dir = dir
		_ = cmd.Run() // Ignore error intentionally
	}

	// Don't configure remote origin for test repos
	// This prevents fetch attempts during tests

	// Set up refs/remotes/origin/HEAD to point to main/master
	originHeadPath := filepath.Join(dir, "refs", "remotes", "origin")
	if err := os.MkdirAll(originHeadPath, 0755); err != nil {
		t.Fatalf("Failed to create origin directory: %v", err)
	}

	// Check which branch exists and set origin/HEAD accordingly
	if _, err := os.Stat(filepath.Join(dir, "refs", "heads", "main")); err == nil {
		if err := os.WriteFile(filepath.Join(originHeadPath, "HEAD"), []byte("ref: refs/remotes/origin/main\n"), 0644); err != nil {
			t.Fatalf("Failed to write origin/HEAD: %v", err)
		}
		// Copy main to origin/main
		mainContent, _ := os.ReadFile(filepath.Join(dir, "refs", "heads", "main"))
		if err := os.WriteFile(filepath.Join(originHeadPath, "main"), mainContent, 0644); err != nil {
			t.Fatalf("Failed to write origin/main: %v", err)
		}
	} else if _, err := os.Stat(filepath.Join(dir, "refs", "heads", "master")); err == nil {
		if err := os.WriteFile(filepath.Join(originHeadPath, "HEAD"), []byte("ref: refs/remotes/origin/master\n"), 0644); err != nil {
			t.Fatalf("Failed to write origin/HEAD: %v", err)
		}
		// Copy master to origin/master
		masterContent, _ := os.ReadFile(filepath.Join(dir, "refs", "heads", "master"))
		if err := os.WriteFile(filepath.Join(originHeadPath, "master"), masterContent, 0644); err != nil {
			t.Fatalf("Failed to write origin/master: %v", err)
		}
	}
}
