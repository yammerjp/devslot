package command

// Version is set by the main package
var Version = "dev"

type VersionCmd struct{}

func (c *VersionCmd) Run(ctx *Context) error {
	ctx.Printf("devslot version %s\n", Version)
	ctx.LogInfo("version requested", "version", Version)
	return nil
}
