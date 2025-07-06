package command

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/yammerjp/devslot/internal/config"
	"github.com/yammerjp/devslot/internal/lock"
	"github.com/yammerjp/devslot/internal/slot"
)

type CreateCmd struct {
	SlotName string `arg:"" help:"Name of the slot to create"`
}

func (c *CreateCmd) Run(ctx *Context) error {
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
	ctx.LogInfo("creating slot", "name", c.SlotName)

	if err := mgr.Create(c.SlotName, cfg); err != nil {
		return fmt.Errorf("failed to create slot: %w", err)
	}

	ctx.Printf("Slot '%s' created successfully!\n", c.SlotName)
	ctx.Printf("You can now work in: %s/slots/%s\n", projectRoot, c.SlotName)
	ctx.LogInfo("slot created successfully", "name", c.SlotName, "path", filepath.Join(projectRoot, "slots", c.SlotName))

	return nil
}
