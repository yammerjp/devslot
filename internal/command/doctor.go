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

	ctx.Println("Running devslot doctor...")
	ctx.Printf("Project root: %s\n\n", projectRoot)
	ctx.LogInfo("running doctor check", "projectRoot", projectRoot)

	hasIssues := false

	// Check configuration
	ctx.Println("Checking configuration...")
	cfg, err := config.Load(projectRoot)
	if err != nil {
		ctx.Printf("  ❌ Failed to load devslot.yaml: %v\n", err)
		ctx.LogError("failed to load configuration", "error", err)
		hasIssues = true
	} else {
		ctx.Println("  ✅ devslot.yaml is valid")
		ctx.Printf("  📦 Found %d repositories\n", len(cfg.Repositories))
		ctx.LogInfo("configuration loaded", "repositoryCount", len(cfg.Repositories))
	}

	// Check directories
	ctx.Println("\nChecking directories...")
	dirs := []string{"hooks", "repos", "slots"}
	for _, dir := range dirs {
		dirPath := filepath.Join(projectRoot, dir)
		if info, err := os.Stat(dirPath); err != nil {
			ctx.Printf("  ❌ Directory %s does not exist\n", dir)
			ctx.LogWarn("directory not found", "directory", dir)
			hasIssues = true
		} else if !info.IsDir() {
			ctx.Printf("  ❌ %s is not a directory\n", dir)
			ctx.LogWarn("path is not a directory", "path", dir)
			hasIssues = true
		} else {
			ctx.Printf("  ✅ Directory %s exists\n", dir)
		}
	}

	// Check repositories
	if cfg != nil {
		ctx.Println("\nChecking repositories...")
		for _, repo := range cfg.Repositories {
			bareRepoPath := filepath.Join(projectRoot, "repos", repo.BareRepoName())
			if git.IsValidRepository(bareRepoPath) {
				ctx.Printf("  ✅ Repository %s is cloned\n", repo.Name)
			} else {
				ctx.Printf("  ❌ Repository %s is not cloned (run 'devslot init')\n", repo.Name)
				ctx.LogWarn("repository not cloned", "repository", repo.Name)
				hasIssues = true
			}
		}
	}

	// Check hooks
	ctx.Println("\nChecking hooks...")
	hooks := []string{"post-init", "post-create", "pre-destroy", "post-destroy", "post-reload"}
	for _, hookName := range hooks {
		hookPath := filepath.Join(projectRoot, "hooks", hookName)
		if info, err := os.Stat(hookPath); err == nil {
			if info.Mode().Perm()&0111 != 0 {
				ctx.Printf("  ✅ Hook %s exists and is executable\n", hookName)
			} else {
				ctx.Printf("  ⚠️  Hook %s exists but is not executable\n", hookName)
				ctx.LogWarn("hook not executable", "hook", hookName)
			}
		} else {
			ctx.Printf("  ℹ️  Hook %s not found (optional)\n", hookName)
		}
	}

	// Summary
	ctx.Println("\n" + strings.Repeat("-", 40))
	if hasIssues {
		ctx.Println("❌ Some issues were found. Please fix them before continuing.")
		ctx.LogError("doctor check failed")
		return fmt.Errorf("doctor check failed")
	} else {
		ctx.Println("✅ Everything looks good!")
		ctx.LogInfo("doctor check passed")
	}

	return nil
}
