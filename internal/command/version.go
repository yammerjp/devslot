package command

import "fmt"

var version = "dev"

type VersionCmd struct{}

func (c *VersionCmd) Run(ctx *Context) error {
	fmt.Fprintf(ctx.Writer, "devslot version %s\n", version)
	return nil
}