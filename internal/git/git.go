package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/yammerjp/devslot/internal/errors"
)

// CloneBare clones a repository as a bare repository
func CloneBare(url, destPath string) error {
	cmd := exec.Command("git", "clone", "--bare", url, destPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// CloneBareShallow clones a repository as a shallow bare repository
func CloneBareShallow(url, destPath string, depth int) error {
	args := []string{"clone", "--bare", "--depth", fmt.Sprintf("%d", depth), url, destPath}
	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Unshallow converts a shallow repository to a complete one
func Unshallow(bareRepoPath string) error {
	cmd := exec.Command("git", "-C", bareRepoPath, "fetch", "--unshallow")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// IsShallow checks if a repository is shallow
func IsShallow(bareRepoPath string) (bool, error) {
	shallowFile := filepath.Join(bareRepoPath, "shallow")
	_, err := os.Stat(shallowFile)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// CreateWorktree creates a new worktree for a bare repository
func CreateWorktree(bareRepoPath, worktreePath, branch string) error {
	// First, check if the branch exists
	checkCmd := exec.Command("git", "-C", bareRepoPath, "show-ref", "--verify", "--quiet", fmt.Sprintf("refs/heads/%s", branch))
	if err := checkCmd.Run(); err != nil {
		// Branch doesn't exist, create worktree with a new branch
		cmd := exec.Command("git", "-C", bareRepoPath, "worktree", "add", "-b", branch, worktreePath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	// Branch exists, create worktree tracking the existing branch
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

// GetDefaultBranch returns the default branch name for a repository
func GetDefaultBranch(bareRepoPath string) (string, error) {
	// Try to get the symbolic ref for HEAD
	cmd := exec.Command("git", "-C", bareRepoPath, "symbolic-ref", "refs/remotes/origin/HEAD")
	output, err := cmd.Output()
	if err == nil {
		// Extract branch name from refs/remotes/origin/main
		branch := string(output)
		if len(branch) > 0 && branch[len(branch)-1] == '\n' {
			branch = branch[:len(branch)-1]
		}
		// Remove the refs/remotes/origin/ prefix
		const prefix = "refs/remotes/origin/"
		if len(branch) > len(prefix) {
			return branch[len(prefix):], nil
		}
	}

	// Fallback: check common default branch names
	for _, branch := range []string{"main", "master"} {
		checkCmd := exec.Command("git", "-C", bareRepoPath, "show-ref", "--verify", "--quiet", fmt.Sprintf("refs/heads/%s", branch))
		if err := checkCmd.Run(); err == nil {
			return branch, nil
		}
	}

	// Last resort: get the first branch
	cmd = exec.Command("git", "-C", bareRepoPath, "for-each-ref", "--format=%(refname:short)", "--count=1", "refs/heads/")
	output, err = cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to find any branch: %w", err)
	}

	branch := string(output)
	if len(branch) > 0 && branch[len(branch)-1] == '\n' {
		branch = branch[:len(branch)-1]
	}

	if branch == "" {
		return "", errors.NoBranchesFound()
	}

	return branch, nil
}

// Fetch fetches updates from origin
func Fetch(bareRepoPath string) error {
	cmd := exec.Command("git", "-C", bareRepoPath, "fetch", "origin", "+refs/heads/*:refs/remotes/origin/*")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// GetBranchPrefix returns the branch prefix for new branches
func GetBranchPrefix() string {
	// 1. Environment variable (for temporary override)
	if prefix := os.Getenv("DEVSLOT_BRANCH_PREFIX"); prefix != "" {
		return prefix
	}

	// 2. Git config (persistent setting)
	if prefix := getGitConfig("devslot.branchPrefix"); prefix != "" {
		return prefix
	}

	// 3. Git email local part (default)
	if localPart := getGitEmailLocalPart(); localPart != "" {
		return fmt.Sprintf("devslot/%s/", localPart)
	}

	// 4. Fallback
	return "devslot/user/"
}

// getGitConfig reads a git config value
func getGitConfig(key string) string {
	cmd := exec.Command("git", "config", "--get", key)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// getGitEmailLocalPart extracts the local part from git user.email
func getGitEmailLocalPart() string {
	email := getGitConfig("user.email")
	if email == "" {
		return ""
	}

	// Extract local part before @
	parts := strings.Split(email, "@")
	if len(parts) == 0 {
		return ""
	}

	// Sanitize for branch name
	return SanitizeBranchComponent(parts[0])
}

// SanitizeBranchComponent ensures the string is safe for git branch names
func SanitizeBranchComponent(name string) string {
	// Convert to lowercase
	name = strings.ToLower(name)

	// Replace unsafe characters with hyphens
	reg := regexp.MustCompile(`[^a-z0-9\-_]+`)
	name = reg.ReplaceAllString(name, "-")

	// Replace multiple hyphens with single hyphen
	reg = regexp.MustCompile(`-+`)
	name = reg.ReplaceAllString(name, "-")

	// Trim hyphens and underscores from start and end
	name = strings.Trim(name, "-_")

	// Return "user" if empty
	if name == "" {
		return "user"
	}

	return name
}

// CreateWorktreeWithFetch creates a new worktree after fetching latest changes
func CreateWorktreeWithFetch(bareRepoPath, worktreePath, slotName string) error {
	// Check if remote origin exists
	checkRemoteCmd := exec.Command("git", "-C", bareRepoPath, "remote", "get-url", "origin")
	if err := checkRemoteCmd.Run(); err != nil {
		// No remote origin, create without fetch (for tests)
		return CreateWorktreeWithoutFetch(bareRepoPath, worktreePath, slotName)
	}

	// 1. Fetch latest changes
	if err := Fetch(bareRepoPath); err != nil {
		return errors.FetchFailed(err)
	}

	// 2. Get default branch
	defaultBranch, err := GetDefaultBranch(bareRepoPath)
	if err != nil {
		return err
	}

	// 3. Generate branch name
	prefix := GetBranchPrefix()
	branchName := fmt.Sprintf("%s%s", prefix, slotName)

	// 4. Create worktree with new branch from origin/defaultBranch
	cmd := exec.Command("git", "-C", bareRepoPath,
		"worktree", "add", "-b", branchName,
		worktreePath,
		fmt.Sprintf("origin/%s", defaultBranch))

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// CreateWorktreeWithoutFetch creates a new worktree without fetching (for local/test repos)
func CreateWorktreeWithoutFetch(bareRepoPath, worktreePath, slotName string) error {
	// Get default branch
	defaultBranch, err := GetDefaultBranch(bareRepoPath)
	if err != nil {
		return err
	}

	// Generate branch name
	prefix := GetBranchPrefix()
	branchName := fmt.Sprintf("%s%s", prefix, slotName)

	// Create worktree with new branch from local defaultBranch
	cmd := exec.Command("git", "-C", bareRepoPath,
		"worktree", "add", "-b", branchName,
		worktreePath,
		defaultBranch)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
