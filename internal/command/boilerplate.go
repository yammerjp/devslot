package command

import (
	"fmt"
	"os"
	"path/filepath"
)

type BoilerplateCmd struct{}

func (c *BoilerplateCmd) Run(ctx *Context) error {
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Create directories
	directories := []string{
		"hooks",
		"repos",
		"slots",
	}

	for _, dir := range directories {
		dirPath := filepath.Join(currentDir, dir)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
		fmt.Fprintf(ctx.Writer, "Created directory: %s\n", dir)
	}

	// Create devslot.yaml
	devslotYamlPath := filepath.Join(currentDir, "devslot.yaml")
	devslotYamlContent := `# devslot configuration file
version: 1
repositories:
  # Add your repositories here
  # Example:
  # - name: my-app.git
  #   url: https://github.com/myorg/my-app.git
  # - name: my-lib.git
  #   url: https://github.com/myorg/my-lib.git
`
	if err := createFileIfNotExists(devslotYamlPath, devslotYamlContent); err != nil {
		return fmt.Errorf("failed to create devslot.yaml: %w", err)
	}
	fmt.Fprintf(ctx.Writer, "Created file: devslot.yaml\n")

	// Create .gitignore
	gitignorePath := filepath.Join(currentDir, ".gitignore")
	gitignoreContent := `# devslot directories
/repos/
/slots/

# OS files
.DS_Store
Thumbs.db

# Editor files
.vscode/
.idea/
*.swp
*.swo
*~
`
	if err := createOrAppendToFile(gitignorePath, gitignoreContent); err != nil {
		return fmt.Errorf("failed to update .gitignore: %w", err)
	}
	fmt.Fprintf(ctx.Writer, "Updated file: .gitignore\n")

	// Create hook examples
	hookExamples := map[string]string{
		"post-create": `#!/bin/bash
# This hook is called after a new slot is created
# Environment variables:
#   DEVSLOT_SLOT: The name of the slot
#   DEVSLOT_PROJECT_ROOT: The root directory of the project

echo "Slot $DEVSLOT_SLOT has been created!"

# Example: Install dependencies for each repository
# for repo in "$DEVSLOT_PROJECT_ROOT/slots/$DEVSLOT_SLOT"/*; do
#     if [ -f "$repo/package.json" ]; then
#         echo "Installing npm dependencies in $(basename "$repo")..."
#         (cd "$repo" && npm install)
#     fi
# done
`,
		"pre-destroy": `#!/bin/bash
# This hook is called before a slot is destroyed
# Environment variables:
#   DEVSLOT_SLOT: The name of the slot
#   DEVSLOT_PROJECT_ROOT: The root directory of the project

echo "Slot $DEVSLOT_SLOT will be destroyed!"

# Example: Backup important files
# backup_dir="$DEVSLOT_PROJECT_ROOT/backups/$DEVSLOT_SLOT-$(date +%Y%m%d-%H%M%S)"
# mkdir -p "$backup_dir"
# echo "Backing up slot to $backup_dir..."
`,
		"post-reload": `#!/bin/bash
# This hook is called after a slot is reloaded
# Environment variables:
#   DEVSLOT_SLOT: The name of the slot
#   DEVSLOT_PROJECT_ROOT: The root directory of the project

echo "Slot $DEVSLOT_SLOT has been reloaded!"

# Example: Sync dependencies or update configurations
# echo "Updating dependencies..."
`,
	}

	for hookName, content := range hookExamples {
		hookPath := filepath.Join(currentDir, "hooks", hookName+".example")
		if err := createFileIfNotExists(hookPath, content); err != nil {
			return fmt.Errorf("failed to create hook example %s: %w", hookName, err)
		}
		fmt.Fprintf(ctx.Writer, "Created hook example: hooks/%s.example\n", hookName)
	}

	fmt.Fprintln(ctx.Writer, "\nBoilerplate project structure created successfully!")
	fmt.Fprintln(ctx.Writer, "Next steps:")
	fmt.Fprintln(ctx.Writer, "1. Edit devslot.yaml to add your repositories")
	fmt.Fprintln(ctx.Writer, "2. Run 'devslot init' to clone the repositories")
	fmt.Fprintln(ctx.Writer, "3. Create your first slot with 'devslot create <slot-name>'")

	return nil
}

func createFileIfNotExists(path, content string) error {
	if _, err := os.Stat(path); err == nil {
		return nil // File already exists
	}

	return os.WriteFile(path, []byte(content), 0644)
}

func createOrAppendToFile(path, content string) error {
	// Check if file exists
	existingContent := ""
	if data, err := os.ReadFile(path); err == nil {
		existingContent = string(data)
	}

	// Check if devslot entries already exist
	if contains(existingContent, "/repos/") && contains(existingContent, "/slots/") {
		return nil // Already configured
	}

	// Append to existing content or create new
	var finalContent string
	if existingContent != "" {
		finalContent = existingContent
		if len(existingContent) > 0 && existingContent[len(existingContent)-1] != '\n' {
			finalContent += "\n"
		}
		finalContent += "\n" + content
	} else {
		finalContent = content
	}

	return os.WriteFile(path, []byte(finalContent), 0644)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && containsHelper(s, substr)
}

func containsHelper(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
