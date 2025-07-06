package command

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/yammerjp/devslot/internal/testutil"
)

func TestBoilerplateCmd_Run(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := testutil.TempDir(t)
	
	// Change to temp directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}
	defer os.Chdir(originalDir)

	// Run boilerplate command
	var buf bytes.Buffer
	cmd := &BoilerplateCmd{}
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
		"Created hook example: hooks/post-create.example",
		"Created hook example: hooks/pre-destroy.example",
		"Created hook example: hooks/post-reload.example",
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
		"hooks/post-create.example",
		"hooks/pre-destroy.example",
		"hooks/post-reload.example",
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
}

func TestBoilerplateCmd_RunTwice(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := testutil.TempDir(t)
	
	// Change to temp directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}
	defer os.Chdir(originalDir)

	// Run boilerplate command twice
	cmd := &BoilerplateCmd{}
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
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}
	defer os.Chdir(originalDir)

	// Create existing .gitignore
	existingContent := "node_modules/\n*.log\n"
	testutil.CreateFile(t, filepath.Join(tempDir, ".gitignore"), existingContent)

	// Run boilerplate command
	cmd := &BoilerplateCmd{}
	ctx := &Context{Writer: &bytes.Buffer{}}

	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("BoilerplateCmd.Run() error = %v", err)
	}

	// Check that .gitignore was updated
	gitignore := testutil.ReadFile(t, filepath.Join(tempDir, ".gitignore"))
	if !contains(gitignore, "node_modules/") {
		t.Error(".gitignore missing existing content")
	}
	if !contains(gitignore, "/repos/") || !contains(gitignore, "/slots/") {
		t.Error(".gitignore missing devslot directories")
	}
}

