package command

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/yammerjp/devslot/internal/config"
	"github.com/yammerjp/devslot/internal/lock"
	"github.com/yammerjp/devslot/internal/slot"
)

type ReloadCmd struct {
	SlotName string `arg:"" help:"Name of the slot to reload"`
}

func (c *ReloadCmd) Help() string {
	return `Ensures all repositories are checked out as worktrees for the slot.

Automatically creates any missing worktrees (useful after adding new
repositories to devslot.yaml). Runs post-reload hook if it exists.`
}

func (c *ReloadCmd) Run(ctx *Context) error {
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
	lockFile := lock.New(filepath.Join(projectRoot, ".devslot.lock"))
	if err := lockFile.Acquire(); err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}
	defer func() {
		if err := lockFile.Release(); err != nil {
			ctx.Printf("Warning: failed to release lock: %v\n", err)
			ctx.LogWarn("failed to release lock", "error", err)
		}
	}()

	// Load configuration
	cfg, err := config.Load(projectRoot)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Reload slot
	mgr := slot.NewManager(projectRoot)
	ctx.Printf("Reloading slot '%s'...\n", c.SlotName)
	ctx.LogInfo("reloading slot", "slot", c.SlotName)

	if err := mgr.Reload(c.SlotName, cfg); err != nil {
		return fmt.Errorf("failed to reload slot: %w", err)
	}

	ctx.Printf("Slot '%s' reloaded successfully!\n", c.SlotName)
	ctx.LogInfo("slot reloaded", "slot", c.SlotName)

	return nil
}
