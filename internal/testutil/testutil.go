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

	// Push to bare repo
	cmd = exec.Command("git", "push", dir, "master")
	cmd.Dir = tempDir
	if output, err := cmd.CombinedOutput(); err != nil {
		// Try with main branch if master fails
		cmd = exec.Command("git", "branch", "-M", "main")
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			// Log warning but don't fail, as this is a fallback operation
			t.Logf("Warning: failed to rename branch to main: %v", err)
		}

		cmd = exec.Command("git", "push", dir, "main")
		cmd.Dir = tempDir
		if output2, err2 := cmd.CombinedOutput(); err2 != nil {
			t.Fatalf("failed to push to bare repo: %v\nOutput: %s\n%s", err2, output, output2)
		}
	}
}
