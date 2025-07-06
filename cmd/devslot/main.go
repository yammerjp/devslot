package main

import (
	"io"
	"os"

	"github.com/alecthomas/kong"
	"github.com/yammerjp/devslot/internal/command"
)

type CLI struct {
	Boilerplate command.BoilerplateCmd `cmd:"" help:"Generate initial project structure in current directory"`
	Init        command.InitCmd        `cmd:"" help:"Sync bare repositories defined in devslot.yaml into repos/"`
	Create      command.CreateCmd      `cmd:"" help:"Create a new slot (multi-repo worktree environment)"`
	Destroy     command.DestroyCmd     `cmd:"" help:"Remove the specified slot (runs pre-destroy hook first)"`
	Reload      command.ReloadCmd      `cmd:"" help:"Ensure all worktrees exist for the slot and run post-reload hook"`
	List        command.ListCmd        `cmd:"" help:"List all existing slots"`
	Doctor      command.DoctorCmd      `cmd:"" help:"Check consistency of project structure and repositories"`
	Version     command.VersionCmd     `cmd:"" help:"Show devslot version"`

	VersionFlag kong.VersionFlag `short:"v" name:"version" help:"Show version"`
}

type App struct {
	parser      *kong.Kong
	writer      io.Writer
	exitHandler func(int)
}

func NewApp(writer io.Writer) *App {
	app := &App{
		writer:      writer,
		exitHandler: os.Exit,
	}
	
	cli := &CLI{}
	parser, err := kong.New(cli,
		kong.Name("devslot"),
		kong.Description("Development environment manager for multi-repo worktrees"),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
		}),
		kong.Writers(writer, writer),
		kong.Exit(func(code int) {
			app.exitHandler(code)
		}),
		kong.Vars{
			"version": "dev", // This should be set by ldflags during build
		},
	)
	if err != nil {
		panic(err)
	}

	app.parser = parser
	return app
}

// SetExitHandler sets a custom exit handler for testing
func (app *App) SetExitHandler(handler func(int)) {
	app.exitHandler = handler
}

func (app *App) Run(args []string) error {
	// Show help if no arguments provided
	if len(args) == 0 {
		_, err := app.parser.Parse([]string{"--help"})
		return err
	}

	ctx, err := app.parser.Parse(args)
	if err != nil {
		return err
	}

	cmdCtx := &command.Context{
		Writer: app.writer,
	}

	return ctx.Run(cmdCtx)
}

func main() {
	app := NewApp(os.Stdout)
	if err := app.Run(os.Args[1:]); err != nil {
		app.parser.FatalIfErrorf(err)
		os.Exit(1)
	}
}
