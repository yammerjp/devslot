package command

import (
	"fmt"
	"os"
	"path/filepath"
)

type BoilerplateCmd struct {
	Dir string `arg:"" required:"" help:"Directory to create project structure in (use . for current directory)"`
}

func (c *BoilerplateCmd) Help() string {
	return `Creates initial project structure for devslot.

Creates the following:
  - devslot.yaml    (project configuration template)
  - .gitignore      (ignores repos/ and slots/)
  - hooks/          (optional lifecycle scripts)
    - post-init     (runs after 'devslot init')
    - post-create   (runs after 'devslot create')
    - pre-destroy   (runs before 'devslot destroy')
    - post-destroy  (runs after 'devslot destroy')
    - post-reload   (runs after 'devslot reload')
  - repos/          (for bare repositories)
  - slots/          (for worktrees)

Creates the target directory if it doesn't exist.
All hooks are optional and include helpful examples.`
}

func (c *BoilerplateCmd) Run(ctx *Context) error {
	// Resolve target directory
	targetDir := c.Dir
	if !filepath.IsAbs(targetDir) {
		currentDir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
		targetDir = filepath.Join(currentDir, targetDir)
	}

	// Create target directory if it doesn't exist
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Create directories
	directories := []string{
		"hooks",
		"repos",
		"slots",
	}

	for _, dir := range directories {
		dirPath := filepath.Join(targetDir, dir)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
		ctx.Printf("Created directory: %s\n", dir)
		ctx.LogInfo("directory created", "directory", dir)
	}

	// Create devslot.yaml
	devslotYamlPath := filepath.Join(targetDir, "devslot.yaml")
	devslotYamlContent := `# devslot configuration file
version: 1
repositories:
  # Add your repositories here (without .git suffix)
  # Example:
  # - name: my-app
  #   url: https://github.com/myorg/my-app.git
  # - name: my-lib
  #   url: https://github.com/myorg/my-lib.git
`
	if err := createFileIfNotExists(devslotYamlPath, devslotYamlContent); err != nil {
		return fmt.Errorf("failed to create devslot.yaml: %w", err)
	}
	ctx.Printf("Created file: devslot.yaml\n")
	ctx.LogInfo("devslot.yaml created")

	// Create .gitignore
	gitignorePath := filepath.Join(targetDir, ".gitignore")
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
		"post-init": `#!/bin/bash
# This hook is called after 'devslot init' clones/updates repositories
# Environment variables:
#   DEVSLOT_ROOT: The root directory of the project
#   DEVSLOT_REPOS_DIR: The full path to the repos directory
#   DEVSLOT_REPOSITORIES: Space-separated list of repository names

# echo "Repositories initialized: $DEVSLOT_REPOSITORIES"

# Example: Set up git config for all repositories
# for repo in "$DEVSLOT_REPOS_DIR"/*.git; do
#     if [ -d "$repo" ]; then
#         echo "Configuring $(basename "$repo")..."
#         git -C "$repo" config core.hooksPath "$DEVSLOT_ROOT/hooks/git"
#     fi
# done

# Example: Fetch all remote branches
# for repo in "$DEVSLOT_REPOS_DIR"/*.git; do
#     if [ -d "$repo" ]; then
#         echo "Fetching all branches for $(basename "$repo")..."
#         git -C "$repo" fetch --all
#     fi
# done
`,
		"post-create": `#!/bin/bash
# This hook is called after a new slot is created
# Environment variables:
#   DEVSLOT_ROOT: The root directory of the project
#   DEVSLOT_SLOT_NAME: The name of the slot
#   DEVSLOT_SLOT_DIR: The full path to the slot directory
#   DEVSLOT_REPOS_DIR: The full path to the repos directory
#   DEVSLOT_REPOSITORIES: Space-separated list of repository names

# echo "Slot $DEVSLOT_SLOT_NAME has been created with repos: $DEVSLOT_REPOSITORIES"

# Example: Install dependencies for each repository
# for repo in "$DEVSLOT_SLOT_DIR"/*; do
#     if [ -f "$repo/package.json" ]; then
#         echo "Installing npm dependencies in $(basename "$repo")..."
#         (cd "$repo" && npm install)
#     fi
# done
`,
		"pre-destroy": `#!/bin/bash
# This hook is called before a slot is destroyed
# Environment variables:
#   DEVSLOT_ROOT: The root directory of the project
#   DEVSLOT_SLOT_NAME: The name of the slot
#   DEVSLOT_SLOT_DIR: The full path to the slot directory
#   DEVSLOT_REPOS_DIR: The full path to the repos directory
#   DEVSLOT_REPOSITORIES: Space-separated list of repository names

# echo "Slot $DEVSLOT_SLOT_NAME will be destroyed (repos: $DEVSLOT_REPOSITORIES)"

# Example: Backup important files
# backup_dir="$DEVSLOT_ROOT/backups/$DEVSLOT_SLOT_NAME-$(date +%Y%m%d-%H%M%S)"
# mkdir -p "$backup_dir"
# echo "Backing up slot to $backup_dir..."
`,
		"post-destroy": `#!/bin/bash
# This hook is called after a slot is destroyed
# Environment variables:
#   DEVSLOT_ROOT: The root directory of the project
#   DEVSLOT_SLOT_NAME: The name of the slot that was destroyed
#   DEVSLOT_REPOS_DIR: The full path to the repos directory
#   DEVSLOT_REPOSITORIES: Space-separated list of repository names

# echo "Slot $DEVSLOT_SLOT_NAME has been destroyed"

# Example: Clean up related resources
# rm -f "$DEVSLOT_ROOT/.cache/$DEVSLOT_SLOT_NAME"*

# Example: Log the destruction
# echo "$(date): Destroyed slot $DEVSLOT_SLOT_NAME" >> "$DEVSLOT_ROOT/destruction.log"

# Example: Send notification
# notify-send "DevSlot" "Slot $DEVSLOT_SLOT_NAME was destroyed" || true
`,
		"post-reload": `#!/bin/bash
# This hook is called after a slot is reloaded
# Environment variables:
#   DEVSLOT_ROOT: The root directory of the project
#   DEVSLOT_SLOT_NAME: The name of the slot
#   DEVSLOT_SLOT_DIR: The full path to the slot directory
#   DEVSLOT_REPOS_DIR: The full path to the repos directory
#   DEVSLOT_REPOSITORIES: Space-separated list of repository names

# echo "Slot $DEVSLOT_SLOT_NAME has been reloaded (repos: $DEVSLOT_REPOSITORIES)"

# Example: Sync dependencies or update configurations
# echo "Updating dependencies..."
`,
	}

	for hookName, content := range hookScripts {
		hookPath := filepath.Join(targetDir, "hooks", hookName)
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
	ctx.LogInfo("boilerplate created", "directory", targetDir)

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
