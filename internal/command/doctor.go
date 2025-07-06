package command

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yammerjp/devslot/internal/config"
	"github.com/yammerjp/devslot/internal/git"
)

type DoctorCmd struct{}

func (c *DoctorCmd) Run(ctx *Context) error {
	// Find project root
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	projectRoot, err := config.FindProjectRoot(currentDir)
	if err != nil {
		return fmt.Errorf("not in a devslot project: %w", err)
	}

	fmt.Fprintln(ctx.Writer, "Running devslot doctor...")
	fmt.Fprintf(ctx.Writer, "Project root: %s\n\n", projectRoot)

	hasIssues := false

	// Check configuration
	fmt.Fprintln(ctx.Writer, "Checking configuration...")
	cfg, err := config.Load(projectRoot)
	if err != nil {
		fmt.Fprintf(ctx.Writer, "  ❌ Failed to load devslot.yaml: %v\n", err)
		hasIssues = true
	} else {
		fmt.Fprintln(ctx.Writer, "  ✅ devslot.yaml is valid")
		fmt.Fprintf(ctx.Writer, "  📦 Found %d repositories\n", len(cfg.Repositories))
	}

	// Check directories
	fmt.Fprintln(ctx.Writer, "\nChecking directories...")
	dirs := []string{"hooks", "repos", "slots"}
	for _, dir := range dirs {
		dirPath := filepath.Join(projectRoot, dir)
		if info, err := os.Stat(dirPath); err != nil {
			fmt.Fprintf(ctx.Writer, "  ❌ Directory %s does not exist\n", dir)
			hasIssues = true
		} else if !info.IsDir() {
			fmt.Fprintf(ctx.Writer, "  ❌ %s is not a directory\n", dir)
			hasIssues = true
		} else {
			fmt.Fprintf(ctx.Writer, "  ✅ Directory %s exists\n", dir)
		}
	}

	// Check repositories
	if cfg != nil {
		fmt.Fprintln(ctx.Writer, "\nChecking repositories...")
		for _, repo := range cfg.Repositories {
			bareRepoPath := filepath.Join(projectRoot, "repos", repo.Name)
			if git.IsValidRepository(bareRepoPath) {
				fmt.Fprintf(ctx.Writer, "  ✅ Repository %s is cloned\n", repo.Name)
			} else {
				fmt.Fprintf(ctx.Writer, "  ❌ Repository %s is not cloned (run 'devslot init')\n", repo.Name)
				hasIssues = true
			}
		}
	}

	// Check hooks
	fmt.Fprintln(ctx.Writer, "\nChecking hooks...")
	hooks := []string{"post-create", "pre-destroy", "post-reload"}
	for _, hookName := range hooks {
		hookPath := filepath.Join(projectRoot, "hooks", hookName)
		if info, err := os.Stat(hookPath); err == nil {
			if info.Mode().Perm()&0111 != 0 {
				fmt.Fprintf(ctx.Writer, "  ✅ Hook %s exists and is executable\n", hookName)
			} else {
				fmt.Fprintf(ctx.Writer, "  ⚠️  Hook %s exists but is not executable\n", hookName)
			}
		} else {
			fmt.Fprintf(ctx.Writer, "  ℹ️  Hook %s not found (optional)\n", hookName)
		}
	}

	// Summary
	fmt.Fprintln(ctx.Writer, "\n"+strings.Repeat("-", 40))
	if hasIssues {
		fmt.Fprintln(ctx.Writer, "❌ Some issues were found. Please fix them before continuing.")
		return fmt.Errorf("doctor check failed")
	} else {
		fmt.Fprintln(ctx.Writer, "✅ Everything looks good!")
	}

	return nil
}
