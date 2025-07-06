package command

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/yammerjp/devslot/internal/config"
	"github.com/yammerjp/devslot/internal/lock"
	"github.com/yammerjp/devslot/internal/slot"
)

type DestroyCmd struct {
	SlotName string `arg:"" help:"Name of the slot to destroy"`
}

func (c *DestroyCmd) Run(ctx *Context) error {
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

	// Destroy slot
	mgr := slot.NewManager(projectRoot)
	ctx.Printf("Destroying slot '%s'...\n", c.SlotName)
	ctx.LogInfo("destroying slot", "slot", c.SlotName)

	if err := mgr.Destroy(c.SlotName, cfg); err != nil {
		return fmt.Errorf("failed to destroy slot: %w", err)
	}

	ctx.Printf("Slot '%s' destroyed successfully!\n", c.SlotName)
	ctx.LogInfo("slot destroyed", "slot", c.SlotName)

	return nil
}
