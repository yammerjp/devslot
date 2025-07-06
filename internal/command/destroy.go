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
			fmt.Fprintf(ctx.Writer, "Warning: failed to release lock: %v\n", err)
		}
	}()

	// Destroy slot
	mgr := slot.NewManager(projectRoot)
	fmt.Fprintf(ctx.Writer, "Destroying slot '%s'...\n", c.SlotName)

	if err := mgr.Destroy(c.SlotName); err != nil {
		return fmt.Errorf("failed to destroy slot: %w", err)
	}

	fmt.Fprintf(ctx.Writer, "Slot '%s' destroyed successfully!\n", c.SlotName)

	return nil
}
