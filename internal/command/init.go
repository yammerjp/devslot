package command

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/yammerjp/devslot/internal/config"
	"github.com/yammerjp/devslot/internal/git"
	"github.com/yammerjp/devslot/internal/lock"
)

type InitCmd struct {
	AllowDelete bool `help:"Delete repositories no longer listed in devslot.yaml"`
}

func (c *InitCmd) Run(ctx *Context) error {
	// Find project root
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	projectRoot, err := config.FindProjectRoot(currentDir)
	if err != nil {
		return fmt.Errorf("not in a devslot project: %w", err)
	}

	// Acquire lock
	l := lock.New(filepath.Join(projectRoot, ".devslot.lock"))
	if err := l.Acquire(); err != nil {
		return err
	}
	defer func() {
		if err := l.Release(); err != nil {
			fmt.Fprintf(ctx.Writer, "Warning: failed to release lock: %v\n", err)
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

	// Clone each repository as bare
	for _, repo := range cfg.Repositories {
		bareRepoPath := filepath.Join(reposDir, repo.Name)
		
		// Check if repository already exists
		if git.IsValidRepository(bareRepoPath) {
			fmt.Fprintf(ctx.Writer, "Repository %s already exists, skipping...\n", repo.Name)
			continue
		}

		fmt.Fprintf(ctx.Writer, "Cloning %s from %s...\n", repo.Name, repo.URL)
		if err := git.CloneBare(repo.URL, bareRepoPath); err != nil {
			return fmt.Errorf("failed to clone %s: %w", repo.Name, err)
		}
		fmt.Fprintf(ctx.Writer, "Successfully cloned %s\n", repo.Name)
	}

	fmt.Fprintln(ctx.Writer, "\nInitialization complete!")
	fmt.Fprintln(ctx.Writer, "You can now create a slot with 'devslot create <slot-name>'")

	return nil
}