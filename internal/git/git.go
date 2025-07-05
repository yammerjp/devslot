package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// CloneBare clones a repository as a bare repository
func CloneBare(url, destPath string) error {
	cmd := exec.Command("git", "clone", "--bare", url, destPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// CreateWorktree creates a new worktree for a bare repository
func CreateWorktree(bareRepoPath, worktreePath, branch string) error {
	cmd := exec.Command("git", "-C", bareRepoPath, "worktree", "add", worktreePath, branch)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// RemoveWorktree removes a worktree
func RemoveWorktree(bareRepoPath, worktreePath string) error {
	cmd := exec.Command("git", "-C", bareRepoPath, "worktree", "remove", worktreePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// ListWorktrees lists all worktrees for a bare repository
func ListWorktrees(bareRepoPath string) ([]string, error) {
	cmd := exec.Command("git", "-C", bareRepoPath, "worktree", "list", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	// Parse the output to extract worktree paths
	// This is a simplified implementation
	worktrees := []string{}
	lines := string(output)
	// TODO: Implement proper parsing
	_ = lines
	return worktrees, nil
}

// IsValidRepository checks if a path is a valid git repository
func IsValidRepository(path string) bool {
	gitDir := filepath.Join(path, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		return true
	}

	// Check if it's a bare repository
	cmd := exec.Command("git", "-C", path, "rev-parse", "--is-bare-repository")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	return string(output) == "true\n"
}

// GetCurrentBranch returns the current branch name
func GetCurrentBranch(repoPath string) (string, error) {
	cmd := exec.Command("git", "-C", repoPath, "branch", "--show-current")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}

	branch := string(output)
	if len(branch) > 0 && branch[len(branch)-1] == '\n' {
		branch = branch[:len(branch)-1]
	}

	return branch, nil
}