package slot

import (
	stderrors "errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yammerjp/devslot/internal/config"
	"github.com/yammerjp/devslot/internal/errors"
	"github.com/yammerjp/devslot/internal/git"
	"github.com/yammerjp/devslot/internal/hook"
)

// Manager manages slots
type Manager struct {
	projectRoot string
	hookRunner  *hook.Runner
}

// CreateOptions contains options for creating a slot
type CreateOptions struct {
	Branch string // Branch to checkout (empty means default branch)
}

// NewManager creates a new slot manager
func NewManager(projectRoot string) *Manager {
	return &Manager{
		projectRoot: projectRoot,
		hookRunner:  hook.NewRunner(projectRoot),
	}
}

// Create creates a new slot
func (m *Manager) Create(name string, cfg *config.Config, opts *CreateOptions) error {
	if err := m.validateSlotName(name); err != nil {
		return err
	}

	slotPath := m.getSlotPath(name)
	if _, err := os.Stat(slotPath); err == nil {
		return errors.SlotAlreadyExists(name)
	}

	// Create slot directory
	if err := os.MkdirAll(slotPath, 0755); err != nil {
		return fmt.Errorf("failed to create slot directory: %w", err)
	}

	// Create worktrees for each repository
	for _, repo := range cfg.Repositories {
		bareRepoPath := filepath.Join(m.projectRoot, "repos", repo.BareRepoName())
		worktreePath := filepath.Join(slotPath, repo.Name)

		// Ensure bare repository exists
		if !git.IsValidRepository(bareRepoPath) {
			// Cleanup on failure
			os.RemoveAll(slotPath)
			return fmt.Errorf("bare repository %s does not exist (run 'devslot init' first)", repo.Name)
		}

		// Create worktree
		if opts.Branch != "" {
			// Use specified branch
			if err := git.CreateWorktree(bareRepoPath, worktreePath, opts.Branch); err != nil {
				// Cleanup on failure
				os.RemoveAll(slotPath)
				return errors.WorktreeFailed(repo.Name, err)
			}
		} else {
			// Create new branch with fetch
			if err := git.CreateWorktreeWithFetch(bareRepoPath, worktreePath, name); err != nil {
				// Cleanup on failure
				os.RemoveAll(slotPath)
				return errors.WorktreeFailed(repo.Name, err)
			}
		}
	}

	// Run post-create hook
	// Build repository names list
	repoNames := make([]string, len(cfg.Repositories))
	for i, repo := range cfg.Repositories {
		repoNames[i] = repo.Name
	}

	hookEnv := map[string]string{
		"DEVSLOT_REPOSITORIES": strings.Join(repoNames, " "),
	}

	if err := m.hookRunner.Run(hook.PostCreate, name, hookEnv); err != nil {
		// Cleanup on hook failure
		if destroyErr := m.Destroy(name, cfg); destroyErr != nil {
			return fmt.Errorf("post-create hook failed: %w (cleanup also failed: %v)", err, destroyErr)
		}
		return fmt.Errorf("post-create hook failed: %w", err)
	}

	return nil
}

// Destroy removes a slot
func (m *Manager) Destroy(name string, cfg *config.Config) error {
	slotPath := m.getSlotPath(name)
	if _, err := os.Stat(slotPath); os.IsNotExist(err) {
		return errors.SlotNotFound(name)
	}

	// Run pre-destroy hook
	// Build repository names list
	repoNames := make([]string, len(cfg.Repositories))
	for i, repo := range cfg.Repositories {
		repoNames[i] = repo.Name
	}

	hookEnv := map[string]string{
		"DEVSLOT_REPOSITORIES": strings.Join(repoNames, " "),
	}

	if err := m.hookRunner.Run(hook.PreDestroy, name, hookEnv); err != nil {
		return fmt.Errorf("pre-destroy hook failed: %w", err)
	}

	// Remove worktrees
	entries, err := os.ReadDir(slotPath)
	if err != nil {
		return fmt.Errorf("failed to read slot directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Try both with and without .git suffix for backward compatibility
		bareRepoPath := filepath.Join(m.projectRoot, "repos", entry.Name()+".git")
		if !git.IsValidRepository(bareRepoPath) {
			// Fallback to old naming convention
			bareRepoPath = filepath.Join(m.projectRoot, "repos", entry.Name())
		}

		worktreePath := filepath.Join(slotPath, entry.Name())

		if git.IsValidRepository(bareRepoPath) {
			if err := git.RemoveWorktree(bareRepoPath, worktreePath); err != nil {
				// Continue with other worktrees even if one fails
				fmt.Fprintf(os.Stderr, "Warning: failed to remove worktree %s: %v\n", entry.Name(), err)
			}
		}
	}

	// Remove slot directory
	if err := os.RemoveAll(slotPath); err != nil {
		return fmt.Errorf("failed to remove slot directory: %w", err)
	}

	// Run post-destroy hook
	if err := m.hookRunner.Run(hook.PostDestroy, name, hookEnv); err != nil {
		// Just log warning since slot is already destroyed
		fmt.Fprintf(os.Stderr, "Warning: post-destroy hook failed: %v\n", err)
	}

	return nil
}

// List returns all existing slots
func (m *Manager) List() ([]string, error) {
	slotsPath := filepath.Join(m.projectRoot, "slots")

	entries, err := os.ReadDir(slotsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read slots directory: %w", err)
	}

	slots := []string{}
	for _, entry := range entries {
		if entry.IsDir() {
			slots = append(slots, entry.Name())
		}
	}

	return slots, nil
}

// Reload ensures all worktrees exist for a slot
func (m *Manager) Reload(name string, cfg *config.Config) error {
	slotPath := m.getSlotPath(name)
	if _, err := os.Stat(slotPath); os.IsNotExist(err) {
		return errors.SlotNotFound(name)
	}

	// Check each repository
	for _, repo := range cfg.Repositories {
		bareRepoPath := filepath.Join(m.projectRoot, "repos", repo.BareRepoName())
		worktreePath := filepath.Join(slotPath, repo.Name)

		// Check if worktree exists
		if _, err := os.Stat(worktreePath); os.IsNotExist(err) {
			// Get default branch for missing worktree
			branch, err := git.GetDefaultBranch(bareRepoPath)
			if err != nil {
				return fmt.Errorf("failed to determine default branch for %s: %w", repo.Name, err)
			}

			// Create missing worktree
			if err := git.CreateWorktree(bareRepoPath, worktreePath, branch); err != nil {
				return fmt.Errorf("failed to create worktree for %s: %w", repo.Name, err)
			}
		}
	}

	// Run post-reload hook
	// Build repository names list
	repoNames := make([]string, len(cfg.Repositories))
	for i, repo := range cfg.Repositories {
		repoNames[i] = repo.Name
	}

	hookEnv := map[string]string{
		"DEVSLOT_REPOSITORIES": strings.Join(repoNames, " "),
	}

	if err := m.hookRunner.Run(hook.PostReload, name, hookEnv); err != nil {
		return fmt.Errorf("post-reload hook failed: %w", err)
	}

	return nil
}

// getSlotPath returns the path for a slot
func (m *Manager) getSlotPath(name string) string {
	return filepath.Join(m.projectRoot, "slots", name)
}

// validateSlotName validates the slot name
func (m *Manager) validateSlotName(name string) error {
	if name == "" {
		return stderrors.New("slot name cannot be empty")
	}

	if strings.Contains(name, "/") || strings.Contains(name, "\\") {
		return stderrors.New("slot name cannot contain path separators")
	}

	if name == "." || name == ".." {
		return stderrors.New("invalid slot name")
	}

	return nil
}
