package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func (cmd *InitCmd) Run(ctx *Context) error {
	// Find project root
	projectRoot, err := FindProjectRoot()
	if err != nil {
		return err
	}

	// Acquire lock
	lock := NewFileLock(projectRoot)
	if err := lock.Lock(); err != nil {
		return err
	}
	defer lock.Unlock()

	// Load configuration
	configPath := filepath.Join(projectRoot, "devslot.yaml")
	config, err := LoadConfig(configPath)
	if err != nil {
		return err
	}

	// Ensure repos directory exists
	reposDir := filepath.Join(projectRoot, "repos")
	if err := os.MkdirAll(reposDir, 0755); err != nil {
		return fmt.Errorf("failed to create repos directory: %w", err)
	}

	// Clone missing repositories
	for _, repoURL := range config.Repositories {
		repoName := extractRepoName(repoURL)
		if repoName == "" {
			fmt.Fprintf(ctx.Writer, "Warning: skipping invalid repository URL: %s\n", repoURL)
			continue
		}

		repoPath := filepath.Join(reposDir, repoName)

		// Check if repository already exists
		if _, err := os.Stat(repoPath); err == nil {
			fmt.Fprintf(ctx.Writer, "Repository %s already exists, skipping\n", repoName)
			continue
		}

		// Clone as bare repository
		fmt.Fprintf(ctx.Writer, "Cloning %s into %s\n", repoURL, repoName)
		gitCmd := exec.Command("git", "clone", "--bare", repoURL, repoPath)
		gitCmd.Stdout = ctx.Writer
		gitCmd.Stderr = ctx.Writer
		if err := gitCmd.Run(); err != nil {
			return fmt.Errorf("failed to clone %s: %w", repoURL, err)
		}
	}

	// Handle --allow-delete flag
	if cmd.AllowDelete {
		// Build a set of expected repositories
		expectedRepos := make(map[string]bool)
		for _, repoURL := range config.Repositories {
			repoName := extractRepoName(repoURL)
			if repoName != "" {
				expectedRepos[repoName] = true
			}
		}

		// Find and remove unlisted repositories
		entries, err := os.ReadDir(reposDir)
		if err != nil {
			return fmt.Errorf("failed to read repos directory: %w", err)
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			if !expectedRepos[entry.Name()] {
				repoPath := filepath.Join(reposDir, entry.Name())
				fmt.Fprintf(ctx.Writer, "Removing unlisted repository: %s\n", entry.Name())
				if err := os.RemoveAll(repoPath); err != nil {
					return fmt.Errorf("failed to remove %s: %w", entry.Name(), err)
				}
			}
		}
	}

	fmt.Fprintln(ctx.Writer, "Init completed successfully")
	return nil
}

func extractRepoName(repoURL string) string {
	// Remove trailing .git if present
	repoURL = strings.TrimSuffix(repoURL, ".git")

	// Extract the last part of the URL path
	parts := strings.Split(repoURL, "/")
	if len(parts) < 2 {
		return ""
	}

	repoName := parts[len(parts)-1]
	if repoName == "" {
		return ""
	}

	// Add .git suffix for bare repository
	return repoName + ".git"
}