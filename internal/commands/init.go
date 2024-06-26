package commands

import (
	"flag"

	"github.com/go-teal/teal/internal/application"
)

func NewInitCommand(app *application.Application) *InitCommand {
	initCommand := &InitCommand{
		fs:  flag.NewFlagSet("init", flag.ContinueOnError),
		app: app,
	}

	return initCommand
}

type InitCommand struct {
	fs  *flag.FlagSet
	app *application.Application
}

func (genCommand *InitCommand) Name() string {
	return genCommand.fs.Name()
}

func (genCommand *InitCommand) Init(args []string) error {
	return genCommand.fs.Parse(args)
}

func (genCommand *InitCommand) Run() error {
	return genCommand.app.InitProject()
}
