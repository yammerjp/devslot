package errors

import (
	"fmt"
)

// UserError wraps an error with a user-friendly message and suggestion
type UserError struct {
	Err        error
	Message    string
	Suggestion string
}

// Error implements the error interface
func (e *UserError) Error() string {
	if e.Suggestion != "" {
		return fmt.Sprintf("%s: %v\n%s", e.Message, e.Err, e.Suggestion)
	}
	return fmt.Sprintf("%s: %v", e.Message, e.Err)
}

// Unwrap returns the wrapped error
func (e *UserError) Unwrap() error {
	return e.Err
}

// WithSuggestion creates a new UserError with a suggestion
func WithSuggestion(err error, message, suggestion string) error {
	return &UserError{
		Err:        err,
		Message:    message,
		Suggestion: suggestion,
	}
}

// NotInProject returns an error indicating the user is not in a devslot project
func NotInProject(err error) error {
	return WithSuggestion(err,
		"not in a devslot project",
		"Run 'devslot boilerplate .' to create a new project")
}

// SlotAlreadyExists returns an error indicating a slot already exists
func SlotAlreadyExists(name string) error {
	return WithSuggestion(fmt.Errorf("slot already exists"),
		fmt.Sprintf("slot %s already exists", name),
		fmt.Sprintf("Use a different name or destroy the existing slot with 'devslot destroy %s'", name))
}

// SlotNotFound returns an error indicating a slot doesn't exist
func SlotNotFound(name string) error {
	return WithSuggestion(fmt.Errorf("slot not found"),
		fmt.Sprintf("slot %s does not exist", name),
		"Run 'devslot list' to see available slots")
}

// LockFailed returns an error indicating lock acquisition failed
func LockFailed(err error) error {
	return WithSuggestion(err,
		"failed to acquire lock",
		"Another devslot command may be running")
}

// CloneFailed returns an error indicating repository cloning failed
func CloneFailed(repoName string, err error) error {
	return WithSuggestion(err,
		fmt.Sprintf("failed to clone %s", repoName),
		"Check your network connection and repository URL")
}

// FetchFailed returns an error indicating fetch failed
func FetchFailed(err error) error {
	return WithSuggestion(err,
		"failed to fetch latest changes",
		"Check your network connection and repository access")
}

// HookNotExecutable returns an error indicating a hook is not executable
func HookNotExecutable(hookName string) error {
	return WithSuggestion(fmt.Errorf("permission denied"),
		fmt.Sprintf("hook %s is not executable", hookName),
		fmt.Sprintf("Run 'chmod +x hooks/%s' to fix", hookName))
}

// HookFailed returns an error indicating a hook execution failed
func HookFailed(hookName string, err error) error {
	return WithSuggestion(err,
		fmt.Sprintf("hook %s failed", hookName),
		fmt.Sprintf("Check the hook script at hooks/%s for errors", hookName))
}

// WorktreeFailed returns an error indicating worktree creation failed
func WorktreeFailed(repoName string, err error) error {
	return WithSuggestion(err,
		fmt.Sprintf("failed to create worktree for %s", repoName),
		"Ensure the branch exists or try 'devslot init' to update repositories")
}

// ConfigNotFound returns an error indicating devslot.yaml was not found
func ConfigNotFound() error {
	return WithSuggestion(fmt.Errorf("configuration not found"),
		"devslot.yaml not found in any parent directory",
		"Run 'devslot boilerplate .' to create a new project")
}

// YAMLParseFailed returns an error indicating YAML parsing failed
func YAMLParseFailed(err error) error {
	return WithSuggestion(err,
		"failed to parse YAML",
		"Check the devslot.yaml syntax")
}

// UnsupportedVersion returns an error indicating unsupported config version
func UnsupportedVersion(version int) error {
	return WithSuggestion(fmt.Errorf("unsupported version"),
		fmt.Sprintf("unsupported config version: %d", version),
		"Only version 1 is supported")
}

// NoBranchesFound returns an error indicating no branches in repository
func NoBranchesFound() error {
	return WithSuggestion(fmt.Errorf("no branches"),
		"no branches found in repository",
		"The repository may be empty or corrupted")
}