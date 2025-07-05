package hook

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Type represents the type of hook
type Type string

const (
	PostCreate Type = "post-create"
	PreDestroy Type = "pre-destroy"
	PostReload Type = "post-reload"
)

// Runner executes hooks
type Runner struct {
	projectRoot string
}

// NewRunner creates a new hook runner
func NewRunner(projectRoot string) *Runner {
	return &Runner{
		projectRoot: projectRoot,
	}
}

// Run executes a hook if it exists
func (r *Runner) Run(hookType Type, slotName string, env map[string]string) error {
	hookPath := filepath.Join(r.projectRoot, "hooks", string(hookType))
	
	// Check if hook exists and is executable
	info, err := os.Stat(hookPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Hook doesn't exist, which is fine
			return nil
		}
		return fmt.Errorf("failed to stat hook %s: %w", hookType, err)
	}
	
	// Check if file is executable
	if info.Mode().Perm()&0111 == 0 {
		return fmt.Errorf("hook %s is not executable", hookType)
	}
	
	// Prepare command
	cmd := exec.Command(hookPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	// Set environment variables
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("DEVSLOT_SLOT=%s", slotName))
	cmd.Env = append(cmd.Env, fmt.Sprintf("DEVSLOT_PROJECT_ROOT=%s", r.projectRoot))
	
	// Add custom environment variables
	for k, v := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}
	
	// Execute hook
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("hook %s failed: %w", hookType, err)
	}
	
	return nil
}

// Exists checks if a hook exists
func (r *Runner) Exists(hookType Type) bool {
	hookPath := filepath.Join(r.projectRoot, "hooks", string(hookType))
	info, err := os.Stat(hookPath)
	if err != nil {
		return false
	}
	
	// Check if it's a regular file and executable
	return info.Mode().IsRegular() && info.Mode().Perm()&0111 != 0
}