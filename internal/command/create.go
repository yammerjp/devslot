package command

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/yammerjp/devslot/internal/config"
	"github.com/yammerjp/devslot/internal/errors"
	"github.com/yammerjp/devslot/internal/lock"
	"github.com/yammerjp/devslot/internal/slot"
)

type CreateCmd struct {
	SlotName string `arg:"" help:"Name of the slot to create"`
	Branch   string `short:"b" help:"Branch to checkout (if not specified, creates new branch named devslot/<git-email-localpart>/<slot-name>)"`
}

func (c *CreateCmd) Help() string {
	return `Creates a new slot with git worktrees for all repositories.

When -b/--branch is not specified, a new branch is created with the naming pattern:
  devslot/<prefix>/<slot-name>

The <prefix> is determined by (in order of precedence):
  1. DEVSLOT_BRANCH_PREFIX environment variable
  2. git config devslot.branchPrefix
  3. Local part of git user.email (e.g., "john.doe" from "john.doe@example.com")
  4. "user" (fallback)

Example: For user "john.doe@example.com" creating slot "feature-x":
  Branch name: devslot/john-doe/feature-x`
}

func (c *CreateCmd) Run(ctx *Context) error {
	// Find project root
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	projectRoot, err := config.FindProjectRoot(currentDir)
	if err != nil {
		return err // config.FindProjectRoot already returns a user-friendly error
	}

	// Acquire lock
	lockFile := lock.New(filepath.Join(projectRoot, ".devslot.lock"))
	if err := lockFile.Acquire(); err != nil {
		return errors.LockFailed(err)
	}
	defer func() {
		if err := lockFile.Release(); err != nil {
			ctx.LogWarn("failed to release lock", "error", err)
		}
	}()

	// Load configuration
	cfg, err := config.Load(projectRoot)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create slot
	mgr := slot.NewManager(projectRoot)
	ctx.Printf("Creating slot '%s'...\n", c.SlotName)
	ctx.LogInfo("creating slot", "name", c.SlotName, "branch", c.Branch)

	// Prepare options
	opts := &slot.CreateOptions{
		Branch: c.Branch,
	}

	// Show repositories that will be created
	ctx.LogDebug("repositories to create", "count", len(cfg.Repositories))
	for _, repo := range cfg.Repositories {
		ctx.Printf("  - %s\n", repo.Name)
	}

	if err := mgr.Create(c.SlotName, cfg, opts); err != nil {
		return fmt.Errorf("failed to create slot: %w", err)
	}

	ctx.Printf("\nSlot '%s' created successfully!\n", c.SlotName)
	ctx.Printf("You can now work in: %s/slots/%s\n", projectRoot, c.SlotName)
	ctx.LogInfo("slot created successfully", "name", c.SlotName, "path", filepath.Join(projectRoot, "slots", c.SlotName))

	return nil
}
