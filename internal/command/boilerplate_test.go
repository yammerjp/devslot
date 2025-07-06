package command

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yammerjp/devslot/internal/testutil"
)

func TestBoilerplateCmd_Run(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := testutil.TempDir(t)

	// Change to temp directory
	defer testutil.Chdir(t, tempDir)()

	// Run boilerplate command with current directory
	var buf bytes.Buffer
	cmd := &BoilerplateCmd{Dir: "."}
	ctx := &Context{Writer: &buf}

	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("BoilerplateCmd.Run() error = %v", err)
	}

	// Check output
	output := buf.String()
	expectedOutputs := []string{
		"Created directory: hooks",
		"Created directory: repos",
		"Created directory: slots",
		"Created file: devslot.yaml",
		"Updated file: .gitignore",
		"Created hook script: hooks/post-init",
		"Created hook script: hooks/post-create",
		"Created hook script: hooks/pre-destroy",
		"Created hook script: hooks/post-reload",
		"Boilerplate project structure created successfully!",
	}

	for _, expected := range expectedOutputs {
		if !contains(output, expected) {
			t.Errorf("Output missing expected text: %s", expected)
		}
	}

	// Check that directories were created
	dirs := []string{"hooks", "repos", "slots"}
	for _, dir := range dirs {
		dirPath := filepath.Join(tempDir, dir)
		if !testutil.FileExists(t, dirPath) {
			t.Errorf("Directory %s was not created", dir)
		}
	}

	// Check that files were created
	files := []string{
		"devslot.yaml",
		".gitignore",
		"hooks/post-init",
		"hooks/post-create",
		"hooks/pre-destroy",
		"hooks/post-reload",
	}
	for _, file := range files {
		filePath := filepath.Join(tempDir, file)
		if !testutil.FileExists(t, filePath) {
			t.Errorf("File %s was not created", file)
		}
	}

	// Check devslot.yaml content
	devslotYaml := testutil.ReadFile(t, filepath.Join(tempDir, "devslot.yaml"))
	if !contains(devslotYaml, "version: 1") {
		t.Error("devslot.yaml missing 'version: 1'")
	}
	if !contains(devslotYaml, "repositories:") {
		t.Error("devslot.yaml missing 'repositories:' section")
	}

	// Check .gitignore content
	gitignore := testutil.ReadFile(t, filepath.Join(tempDir, ".gitignore"))
	if !contains(gitignore, "/repos/") || !contains(gitignore, "/slots/") {
		t.Error(".gitignore missing devslot directories")
	}

	// Check hook files are executable
	hookFiles := []string{
		"hooks/post-init",
		"hooks/post-create",
		"hooks/pre-destroy",
		"hooks/post-reload",
	}
	for _, hook := range hookFiles {
		hookPath := filepath.Join(tempDir, hook)
		info, err := os.Stat(hookPath)
		if err != nil {
			t.Errorf("Failed to get file info for %s: %v", hook, err)
			continue
		}
		// Check if executable (0755 = -rwxr-xr-x)
		if info.Mode().Perm() != 0755 {
			t.Errorf("Hook file %s has incorrect permissions: %v (expected 0755)", hook, info.Mode().Perm())
		}
	}
}

func TestBoilerplateCmd_RunTwice(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := testutil.TempDir(t)

	// Change to temp directory
	defer testutil.Chdir(t, tempDir)()

	// Run boilerplate command twice
	cmd := &BoilerplateCmd{Dir: "."}
	ctx := &Context{Writer: &bytes.Buffer{}}

	// First run
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("First BoilerplateCmd.Run() error = %v", err)
	}

	// Second run - should not error
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("Second BoilerplateCmd.Run() error = %v", err)
	}

	// Files should still exist
	if !testutil.FileExists(t, filepath.Join(tempDir, "devslot.yaml")) {
		t.Error("devslot.yaml was removed on second run")
	}
}

func TestBoilerplateCmd_ExistingGitignore(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := testutil.TempDir(t)

	// Change to temp directory
	defer testutil.Chdir(t, tempDir)()

	// Create existing .gitignore
	existingContent := "node_modules/\n*.log\n"
	testutil.CreateFile(t, filepath.Join(tempDir, ".gitignore"), existingContent)

	// Run boilerplate command
	cmd := &BoilerplateCmd{Dir: "."}
	ctx := &Context{Writer: &bytes.Buffer{}}

	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("BoilerplateCmd.Run() error = %v", err)
	}

	// Check that .gitignore was updated
	gitignore := testutil.ReadFile(t, filepath.Join(tempDir, ".gitignore"))
	if !strings.Contains(gitignore, "node_modules/") {
		t.Error(".gitignore missing existing content")
	}
	if !strings.Contains(gitignore, "/repos/") || !strings.Contains(gitignore, "/slots/") {
		t.Error(".gitignore missing devslot directories")
	}
}

func TestBoilerplateCmd_WithSubdirectory(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := testutil.TempDir(t)

	// Change to temp directory
	defer testutil.Chdir(t, tempDir)()

	// Run boilerplate command with subdirectory
	var buf bytes.Buffer
	cmd := &BoilerplateCmd{Dir: "my-project"}
	ctx := &Context{Writer: &buf}

	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("BoilerplateCmd.Run() error = %v", err)
	}

	// Check that files were created in subdirectory
	projectDir := filepath.Join(tempDir, "my-project")
	
	// Check directory was created
	if !testutil.DirExists(t, projectDir) {
		t.Error("Project directory was not created")
	}

	// Check all files exist in subdirectory
	expectedFiles := []string{
		"devslot.yaml",
		".gitignore",
		"hooks/post-create",
		"hooks/pre-destroy",
		"hooks/post-reload",
	}

	for _, file := range expectedFiles {
		filePath := filepath.Join(projectDir, file)
		if !testutil.FileExists(t, filePath) {
			t.Errorf("File %s was not created in subdirectory", file)
		}
	}

	// Check directories
	expectedDirs := []string{"hooks", "repos", "slots"}
	for _, dir := range expectedDirs {
		dirPath := filepath.Join(projectDir, dir)
		if !testutil.DirExists(t, dirPath) {
			t.Errorf("Directory %s was not created in subdirectory", dir)
		}
	}
}

func TestBoilerplateCmd_WithAbsolutePath(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := testutil.TempDir(t)
	projectDir := filepath.Join(tempDir, "absolute-project")

	// Run boilerplate command with absolute path
	var buf bytes.Buffer
	cmd := &BoilerplateCmd{Dir: projectDir}
	ctx := &Context{Writer: &buf}

	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("BoilerplateCmd.Run() error = %v", err)
	}

	// Check that files were created in absolute path
	if !testutil.DirExists(t, projectDir) {
		t.Error("Project directory was not created at absolute path")
	}

	// Check devslot.yaml exists
	if !testutil.FileExists(t, filepath.Join(projectDir, "devslot.yaml")) {
		t.Error("devslot.yaml was not created at absolute path")
	}
}
