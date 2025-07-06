package command

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yammerjp/devslot/internal/config"
	"github.com/yammerjp/devslot/internal/errors"
	"github.com/yammerjp/devslot/internal/git"
	"github.com/yammerjp/devslot/internal/hook"
	"github.com/yammerjp/devslot/internal/lock"
)

type InitCmd struct {
	AllowDelete bool `help:"Delete repositories no longer listed in devslot.yaml"`
}

func (c *InitCmd) Help() string {
	return `Clones repositories defined in devslot.yaml as bare repositories into repos/.

This command:
  - Only clones missing repositories (skips existing ones)
  - Does not affect existing slots or worktrees
  - Preserves unlisted repositories unless --allow-delete is used
  - Runs post-init hook if it exists

Safe to run multiple times.`
}

func (c *InitCmd) Run(ctx *Context) error {
	// Find project root
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	ctx.LogDebug("looking for project root", "currentDir", currentDir)
	projectRoot, err := config.FindProjectRoot(currentDir)
	if err != nil {
		return err // config.FindProjectRoot already returns a user-friendly error
	}
	ctx.LogDebug("found project root", "projectRoot", projectRoot)

	// Acquire lock
	l := lock.New(filepath.Join(projectRoot, ".devslot.lock"))
	if err := l.Acquire(); err != nil {
		return err
	}
	defer func() {
		if err := l.Release(); err != nil {
			ctx.LogWarn("failed to release lock", "error", err)
		}
	}()

	// Load configuration
	cfg, err := config.Load(projectRoot)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create repos directory if it doesn't exist
	reposDir := filepath.Join(projectRoot, "repos")
	if err := os.MkdirAll(reposDir, 0755); err != nil {
		return fmt.Errorf("failed to create repos directory: %w", err)
	}

	// Test mode sleep for concurrent lock testing
	if testDelay := os.Getenv("DEVSLOT_TEST_INIT_DELAY"); testDelay != "" {
		if delay, err := time.ParseDuration(testDelay); err == nil {
			time.Sleep(delay)
		}
	}

	// Clone each repository as bare
	for _, repo := range cfg.Repositories {
		bareRepoPath := filepath.Join(reposDir, repo.BareRepoName())

		// Check if repository already exists
		if git.IsValidRepository(bareRepoPath) {
			ctx.Printf("Repository %s already exists, skipping...\n", repo.Name)
			ctx.LogInfo("skipping existing repository", "name", repo.Name)
			continue
		}

		ctx.Printf("Cloning %s from %s...\n", repo.Name, repo.URL)
		ctx.LogInfo("cloning repository", "name", repo.Name, "url", repo.URL)
		if err := git.CloneBare(repo.URL, bareRepoPath); err != nil {
			return errors.CloneFailed(repo.Name, err)
		}
		ctx.Printf("Successfully cloned %s\n", repo.Name)
	}

	// Handle --allow-delete flag
	if c.AllowDelete {
		// Get list of existing repositories
		entries, err := os.ReadDir(reposDir)
		if err != nil {
			return fmt.Errorf("failed to read repos directory: %w", err)
		}

		// Build a map of configured repositories
		configuredRepos := make(map[string]bool)
		for _, repo := range cfg.Repositories {
			configuredRepos[repo.BareRepoName()] = true
		}

		// Remove repositories not in configuration
		for _, entry := range entries {
			if entry.IsDir() && !configuredRepos[entry.Name()] {
				repoPath := filepath.Join(reposDir, entry.Name())
				ctx.Printf("Removing unlisted repository: %s\n", entry.Name())
				ctx.LogInfo("removing unlisted repository", "name", entry.Name())
				if err := os.RemoveAll(repoPath); err != nil {
					ctx.LogWarn("failed to remove repository", "name", entry.Name(), "error", err)
				}
			}
		}
	}

	// Run post-init hook
	hookRunner := hook.NewRunner(projectRoot)
	ctx.LogDebug("running post-init hook")

	// Build repository names list
	repoNames := make([]string, len(cfg.Repositories))
	for i, repo := range cfg.Repositories {
		repoNames[i] = repo.Name
	}

	hookEnv := map[string]string{
		"DEVSLOT_REPOSITORIES": strings.Join(repoNames, " "),
	}

	if err := hookRunner.Run(hook.PostInit, "", hookEnv); err != nil {
		ctx.LogWarn("post-init hook failed", "error", err)
		return fmt.Errorf("post-init hook failed: %w", err)
	}

	ctx.Println("\nInitialization complete!")
	ctx.Println("You can now create a slot with 'devslot create <slot-name>'")
	ctx.LogInfo("initialization completed")

	return nil
}
