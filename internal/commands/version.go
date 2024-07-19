package commands

import (
	"flag"
	"fmt"

	"github.com/go-teal/teal/internal/application"
	"github.com/go-teal/teal/pkg/configs"
)

func NewVersionCommand(app *application.Application) *VersionCommand {
	versionCommand := &VersionCommand{
		fs:  flag.NewFlagSet("version", flag.ContinueOnError),
		app: app,
	}

	return versionCommand
}

type VersionCommand struct {
	fs  *flag.FlagSet
	app *application.Application
}

func (versionCommand *VersionCommand) Name() string {
	return versionCommand.fs.Name()
}

func (versionCommand *VersionCommand) Init(args []string) error {
	return versionCommand.fs.Parse(args)
}

func (genCommand *VersionCommand) Run() error {
	fmt.Println(configs.TEAL_VERSION)
	return nil
}
