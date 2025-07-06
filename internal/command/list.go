package command

import (
	"fmt"
	"os"

	"github.com/yammerjp/devslot/internal/config"
	"github.com/yammerjp/devslot/internal/slot"
)

type ListCmd struct{}

func (c *ListCmd) Run(ctx *Context) error {
	// Find project root
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	projectRoot, err := config.FindProjectRoot(currentDir)
	if err != nil {
		return fmt.Errorf("not in a devslot project: %w", err)
	}

	// List slots
	mgr := slot.NewManager(projectRoot)
	slots, err := mgr.List()
	if err != nil {
		return fmt.Errorf("failed to list slots: %w", err)
	}

	if len(slots) == 0 {
		ctx.Println("No slots found.")
		ctx.Println("Create a new slot with 'devslot create <slot-name>'")
		ctx.LogInfo("no slots found")
		return nil
	}

	ctx.Println("Available slots:")
	ctx.LogInfo("listing slots", "count", len(slots))
	for _, slotName := range slots {
		ctx.Printf("  - %s\n", slotName)
	}

	return nil
}
