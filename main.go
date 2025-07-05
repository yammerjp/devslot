package main

import (
	"fmt"
	"io"
	"os"

	"github.com/alecthomas/kong"
)

type CLI struct {
	Boilerplate BoilerplateCmd `cmd:"" help:"Generate initial project structure in the specified directory"`
	Init        InitCmd        `cmd:"" help:"Sync bare repositories defined in devslot.yaml into repos/"`
	Create      CreateCmd      `cmd:"" help:"Create a new slot (multi-repo worktree environment)"`
	Destroy     DestroyCmd     `cmd:"" help:"Remove the specified slot (runs pre-destroy hook first)"`
	Reload      ReloadCmd      `cmd:"" help:"Ensure all worktrees exist for the slot and run post-reload hook"`
	List        ListCmd        `cmd:"" help:"List all existing slots"`
	Doctor      DoctorCmd      `cmd:"" help:"Check consistency of project structure and repositories"`
	Version     VersionCmd     `cmd:"" help:"Show devslot version"`

	VersionFlag bool `short:"v" help:"Show version (alias for 'version')"`
}

type BoilerplateCmd struct {
	Dir string `arg:"" help:"Target directory"`
}

func (cmd *BoilerplateCmd) Run() error {
	return fmt.Errorf("not implemented")
}

type InitCmd struct {
	AllowDelete bool `help:"Delete repositories no longer listed in devslot.yaml"`
}

func (cmd *InitCmd) Run() error {
	return fmt.Errorf("not implemented")
}

type CreateCmd struct {
	Slot string `arg:"" help:"Slot name"`
}

func (cmd *CreateCmd) Run() error {
	return fmt.Errorf("not implemented")
}

type DestroyCmd struct {
	Slot string `arg:"" help:"Slot name"`
}

func (cmd *DestroyCmd) Run() error {
	return fmt.Errorf("not implemented")
}

type ReloadCmd struct {
	Slot string `arg:"" help:"Slot name"`
}

func (cmd *ReloadCmd) Run() error {
	return fmt.Errorf("not implemented")
}

type ListCmd struct{}

func (cmd *ListCmd) Run() error {
	return fmt.Errorf("not implemented")
}

type DoctorCmd struct{}

func (cmd *DoctorCmd) Run() error {
	return fmt.Errorf("not implemented")
}

type VersionCmd struct{}

func (cmd *VersionCmd) Run() error {
	return fmt.Errorf("not implemented")
}

type App struct {
	parser *kong.Kong
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
	)
	return &App{parser: parser}
}

func (app *App) Run(args []string) error {
	_, err := app.parser.Parse(args)
	return err
}

func main() {
	app := NewApp(os.Stdout)
	if err := app.Run(os.Args[1:]); err != nil {
		os.Exit(1)
	}
}