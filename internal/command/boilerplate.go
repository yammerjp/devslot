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
		ctx.Printf("Created directory: %s\n", dir)
		ctx.LogInfo("directory created", "directory", dir)
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
	ctx.Printf("Created file: devslot.yaml\n")
	ctx.LogInfo("devslot.yaml created")

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
	ctx.Printf("Updated file: .gitignore\n")
	ctx.LogInfo(".gitignore updated")

	// Create hook scripts with executable permissions
	hookScripts := map[string]string{
		"post-create": `#!/bin/bash
# This hook is called after a new slot is created
# Environment variables:
#   DEVSLOT_SLOT: The name of the slot
#   DEVSLOT_PROJECT_ROOT: The root directory of the project

# Example output (comment out if not needed)
# echo "🎉 Post-create hook executed!"
# echo "Slot: $DEVSLOT_SLOT"
# echo "Project root: $DEVSLOT_PROJECT_ROOT"
# echo "Working dir: $(pwd)"
# echo "Hook executed at: $(date)"

# Example: Install dependencies for each repository
# for repo in "$DEVSLOT_PROJECT_ROOT/slots/$DEVSLOT_SLOT"/*; do
#     if [ -f "$repo/package.json" ]; then
#         echo "Installing npm dependencies in $(basename "$repo")..."
#         (cd "$repo" && npm install)
#     fi
# done

# Example: Set up development environment
# echo "Setting up development environment for $DEVSLOT_SLOT..."
# Copy environment files, install tools, etc.

# Example: Send notification
# echo "Slot $DEVSLOT_SLOT is ready!" | notify-send "DevSlot" || true
`,
		"pre-destroy": `#!/bin/bash
# This hook is called before a slot is destroyed
# Environment variables:
#   DEVSLOT_SLOT: The name of the slot
#   DEVSLOT_PROJECT_ROOT: The root directory of the project

# Example output (comment out if not needed)
# echo "🗑️ Pre-destroy hook executed!"
# echo "About to destroy slot: $DEVSLOT_SLOT"
# echo "Project root: $DEVSLOT_PROJECT_ROOT"
# echo "Working dir: $(pwd)"

# Example: Check for uncommitted changes
# for repo in "$DEVSLOT_PROJECT_ROOT/slots/$DEVSLOT_SLOT"/*; do
#     if [ -d "$repo/.git" ]; then
#         if [ -n "$(cd "$repo" && git status --porcelain)" ]; then
#             echo "WARNING: Uncommitted changes in $(basename "$repo")"
#             # Optionally exit with error to prevent destruction
#             # exit 1
#         fi
#     fi
# done

# Example: Backup important files
# backup_dir="$DEVSLOT_PROJECT_ROOT/backups/$DEVSLOT_SLOT-$(date +%Y%m%d-%H%M%S)"
# mkdir -p "$backup_dir"
# echo "Backing up slot to $backup_dir..."
# cp -r "$DEVSLOT_PROJECT_ROOT/slots/$DEVSLOT_SLOT" "$backup_dir/"
`,
		"post-reload": `#!/bin/bash
# This hook is called after a slot is reloaded
# Environment variables:
#   DEVSLOT_SLOT: The name of the slot
#   DEVSLOT_PROJECT_ROOT: The root directory of the project

# Example output (comment out if not needed)
# echo "🔄 Post-reload hook executed!"
# echo "Reloaded slot: $DEVSLOT_SLOT"
# echo "Project root: $DEVSLOT_PROJECT_ROOT"
# echo "Working dir: $(pwd)"

# Example: Sync dependencies or update configurations
# echo "Updating dependencies for $DEVSLOT_SLOT..."
# for repo in "$DEVSLOT_PROJECT_ROOT/slots/$DEVSLOT_SLOT"/*; do
#     if [ -f "$repo/package.json" ]; then
#         echo "Updating npm dependencies in $(basename "$repo")..."
#         (cd "$repo" && npm install)
#     fi
# done

# Example: Run migrations or update database schema
# echo "Running database migrations..."
# (cd "$DEVSLOT_PROJECT_ROOT/slots/$DEVSLOT_SLOT/main-app" && npm run migrate)
`,
	}

	for hookName, content := range hookScripts {
		hookPath := filepath.Join(currentDir, "hooks", hookName)
		if err := createExecutableFile(hookPath, content); err != nil {
			return fmt.Errorf("failed to create hook script %s: %w", hookName, err)
		}
		ctx.Printf("Created hook script: hooks/%s\n", hookName)
		ctx.LogInfo("hook script created", "hook", hookName)
	}

	ctx.Println("\nBoilerplate project structure created successfully!")
	ctx.Println("Next steps:")
	ctx.Println("1. Edit devslot.yaml to add your repositories")
	ctx.Println("2. Run 'devslot init' to clone the repositories")
	ctx.Println("3. Create your first slot with 'devslot create <slot-name>'")
	ctx.LogInfo("boilerplate created")

	return nil
}

func createFileIfNotExists(path, content string) error {
	if _, err := os.Stat(path); err == nil {
		return nil // File already exists
	}

	return os.WriteFile(path, []byte(content), 0644)
}

func createExecutableFile(path, content string) error {
	if _, err := os.Stat(path); err == nil {
		return nil // File already exists
	}

	return os.WriteFile(path, []byte(content), 0755)
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
