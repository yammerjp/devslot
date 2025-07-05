package testutil

import (
	"os"
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
	configContent := `repositories:
  - name: example-repo
    url: https://github.com/example/repo.git
`
	CreateFile(t, filepath.Join(root, "devslot.yaml"), configContent)
	
	// Create .gitignore
	gitignoreContent := `repos/
slots/
`
	CreateFile(t, filepath.Join(root, ".gitignore"), gitignoreContent)
}