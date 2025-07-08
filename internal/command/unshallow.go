package command

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/yammerjp/devslot/internal/config"
	"github.com/yammerjp/devslot/internal/git"
	"github.com/yammerjp/devslot/internal/lock"
	"golang.org/x/sync/errgroup"
)

type UnshallowCmd struct {
	Repository string `arg:"" optional:"" help:"Specific repository to unshallow (unshallows all if not specified)"`
	Parallel   int    `help:"Number of parallel unshallow operations" default:"4"`
}

func (c *UnshallowCmd) Help() string {
	return `Convert shallow repositories to complete ones by fetching full history.

This command:
  - Fetches the complete history for shallow repositories
  - Can target a specific repository or all repositories
  - Runs in parallel for better performance
  - Only affects shallow repositories (skips already complete ones)

Examples:
  devslot unshallow              # Unshallow all repositories
  devslot unshallow frontend     # Unshallow only the frontend repository`
}

func (c *UnshallowCmd) Run(ctx *Context) error {
	// Find project root
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	ctx.LogDebug("looking for project root", "currentDir", currentDir)
	projectRoot, err := config.FindProjectRoot(currentDir)
	if err != nil {
		return err
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

	reposDir := filepath.Join(projectRoot, "repos")

	// Determine which repositories to process
	var targetRepos []config.Repository
	if c.Repository != "" {
		// Find specific repository
		found := false
		for _, repo := range cfg.Repositories {
			if repo.Name == c.Repository {
				targetRepos = append(targetRepos, repo)
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("repository '%s' not found in devslot.yaml", c.Repository)
		}
	} else {
		// Process all repositories
		targetRepos = cfg.Repositories
	}

	// Process repositories in parallel
	var g errgroup.Group
	g.SetLimit(c.Parallel)
	var mu sync.Mutex

	for _, repo := range targetRepos {
		repo := repo // Capture for goroutine
		bareRepoPath := filepath.Join(reposDir, repo.BareRepoName())

		// Check if repository exists
		if !git.IsValidRepository(bareRepoPath) {
			ctx.LogWarn("repository does not exist", "name", repo.Name)
			continue
		}

		g.Go(func() error {
			// Check if repository is shallow
			isShallow, err := git.IsShallow(bareRepoPath)
			if err != nil {
				return fmt.Errorf("failed to check if %s is shallow: %w", repo.Name, err)
			}

			if !isShallow {
				mu.Lock()
				ctx.Printf("Repository %s is already complete, skipping...\n", repo.Name)
				ctx.LogInfo("skipping complete repository", "name", repo.Name)
				mu.Unlock()
				return nil
			}

			mu.Lock()
			ctx.Printf("Unshallowing %s...\n", repo.Name)
			ctx.LogInfo("unshallowing repository", "name", repo.Name)
			mu.Unlock()

			if err := git.Unshallow(bareRepoPath); err != nil {
				return fmt.Errorf("failed to unshallow %s: %w", repo.Name, err)
			}

			mu.Lock()
			ctx.Printf("Successfully unshallowed %s\n", repo.Name)
			mu.Unlock()
			return nil
		})
	}

	// Wait for all operations to complete
	if err := g.Wait(); err != nil {
		return err
	}

	ctx.Println("\nUnshallow complete!")
	ctx.LogInfo("unshallow completed")

	return nil
}