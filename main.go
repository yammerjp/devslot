package main

import (
	"fmt"
	"io"
	"os"

	"github.com/alecthomas/kong"
)

type Context struct {
	Writer io.Writer
}

type CLI struct {
	Boilerplate BoilerplateCmd `cmd:"" help:"Generate initial project structure in the specified directory"`
	Init        InitCmd        `cmd:"" help:"Sync bare repositories defined in devslot.yaml into repos/"`
	Create      CreateCmd      `cmd:"" help:"Create a new slot (multi-repo worktree environment)"`
	Destroy     DestroyCmd     `cmd:"" help:"Remove the specified slot (runs pre-destroy hook first)"`
	Reload      ReloadCmd      `cmd:"" help:"Ensure all worktrees exist for the slot and run post-reload hook"`
	List        ListCmd        `cmd:"" help:"List all existing slots"`
	Doctor      DoctorCmd      `cmd:"" help:"Check consistency of project structure and repositories"`
	Version     VersionCmd     `cmd:"" help:"Show devslot version"`

	VersionFlag kong.VersionFlag `short:"v" name:"version" help:"Show version (alias for 'version')"`
}

type BoilerplateCmd struct {
	Dir string `arg:"" help:"Target directory"`
}

// Run is defined in boilerplate.go

type InitCmd struct {
	AllowDelete bool `help:"Delete repositories no longer listed in devslot.yaml"`
}

func (cmd *InitCmd) Run(ctx *Context) error {
	return fmt.Errorf("not implemented")
}

type CreateCmd struct {
	Slot string `arg:"" help:"Slot name"`
}

func (cmd *CreateCmd) Run(ctx *Context) error {
	return fmt.Errorf("not implemented")
}

type DestroyCmd struct {
	Slot string `arg:"" help:"Slot name"`
}

func (cmd *DestroyCmd) Run(ctx *Context) error {
	return fmt.Errorf("not implemented")
}

type ReloadCmd struct {
	Slot string `arg:"" help:"Slot name"`
}

func (cmd *ReloadCmd) Run(ctx *Context) error {
	return fmt.Errorf("not implemented")
}

type ListCmd struct{}

func (cmd *ListCmd) Run(ctx *Context) error {
	return fmt.Errorf("not implemented")
}

type DoctorCmd struct{}

func (cmd *DoctorCmd) Run(ctx *Context) error {
	return fmt.Errorf("not implemented")
}

var version = "dev"

type VersionCmd struct{}

func (cmd *VersionCmd) Run(ctx *Context) error {
	fmt.Fprintf(ctx.Writer, "devslot version %s\n", version)
	return nil
}

type App struct {
	parser *kong.Kong
	writer io.Writer
}

func NewApp(w io.Writer) *App {
	cli := &CLI{}
	parser, _ := kong.New(cli,
		kong.Name("devslot"),
		kong.Description("Development environment manager for multi-repo worktrees"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
		}),
		kong.Writers(w, w),
		kong.Exit(func(int) {}),
		kong.Bind(&Context{Writer: w}),
		kong.Vars{
			"version": fmt.Sprintf("devslot version %s", version),
		},
	)
	return &App{parser: parser, writer: w}
}

func (app *App) Run(args []string) error {
	// Display help if no args are provided
	if len(args) == 0 {
		args = append(args, "--help")
	}

	ctx, err := app.parser.Parse(args)
	if err != nil {
		return err
	}

	return ctx.Run(&Context{Writer: app.writer})
}

func main() {
	app := NewApp(os.Stdout)
	args := os.Args[1:]

	// Special case: no arguments should show help with exit 0
	if len(args) == 0 {
		_ = app.Run([]string{"--help"})
		return
	}

	if err := app.Run(args); err != nil {
		os.Exit(1)
	}
}
