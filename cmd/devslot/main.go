package main

import (
	"io"
	"log/slog"
	"os"

	"github.com/alecthomas/kong"
	"github.com/yammerjp/devslot/internal/command"
	"github.com/yammerjp/devslot/internal/logger"
)

type CLI struct {
	Verbose     bool                   `short:"v" help:"Enable verbose logging"`
	Boilerplate command.BoilerplateCmd `cmd:"" help:"Create boilerplate structure for a new project"`
	Init        command.InitCmd        `cmd:"" help:"Initialize the project by syncing bare repositories"`
	Create      command.CreateCmd      `cmd:"" help:"Create a new slot"`
	Destroy     command.DestroyCmd     `cmd:"" help:"Destroy an existing slot"`
	Reload      command.ReloadCmd      `cmd:"" help:"Reload a slot to ensure all worktrees exist"`
	List        command.ListCmd        `cmd:"" help:"List all existing slots"`
	Doctor      command.DoctorCmd      `cmd:"" help:"Check project consistency and show diagnostics"`
	Version     command.VersionCmd     `cmd:"" help:"Show version information"`
}

type App struct {
	parser *kong.Kong
	writer io.Writer
	cli    *CLI
}

func NewApp(writer io.Writer) *App {
	cli := &CLI{}
	parser, err := kong.New(cli,
		kong.Name("devslot"),
		kong.Description("A development environment manager for multi-repository worktrees"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
		}),
		kong.Writers(writer, writer),
		kong.Exit(func(int) {}), // Override exit for testing
	)
	if err != nil {
		panic(err)
	}

	return &App{
		parser: parser,
		writer: writer,
		cli:    cli,
	}
}

func (app *App) Run(args []string) error {
	ctx, err := app.parser.Parse(args)
	if err != nil {
		return err
	}

	// Access the CLI struct to get verbose flag

	// Create logger with appropriate log level
	logOpts := logger.DefaultOptions()
	logOpts.Writer = os.Stderr // Log to stderr to keep stdout clean
	if app.cli.Verbose {
		logOpts.Level = slog.LevelDebug
	}
	log := logger.New(logOpts)

	cmdCtx := &command.Context{
		Writer: app.writer,
		Logger: log,
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
