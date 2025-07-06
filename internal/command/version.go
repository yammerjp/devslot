package command

var version = "dev"

type VersionCmd struct{}

func (c *VersionCmd) Run(ctx *Context) error {
	ctx.Printf("devslot version %s\n", version)
	ctx.LogInfo("version requested", "version", version)
	return nil
}
