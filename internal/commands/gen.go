package commands

import (
	"flag"
	"fmt"

	"github.com/go-teal/teal/internal/application"
)

const defaultConfig = "config.yaml"

func NewGenCommand(app *application.Application) *GenCommand {
	genCommand := &GenCommand{
		fs:  flag.NewFlagSet("gen", flag.ContinueOnError),
		app: app,
	}

	genCommand.fs.StringVar(&genCommand.projectPath, "project-path", ".", "Project dir")
	genCommand.fs.StringVar(&genCommand.configFile, "config-file", defaultConfig, "Path to config.yaml")
	genCommand.fs.StringVar(&genCommand.models, "model", "", "name of the target model")

	return genCommand
}

type GenCommand struct {
	fs          *flag.FlagSet
	models      string
	projectPath string
	configFile  string
	app         *application.Application
}

func (genCommand *GenCommand) Name() string {
	return genCommand.fs.Name()
}

func (genCommand *GenCommand) Init(args []string) error {
	genCommand.projectPath = "."
	return genCommand.fs.Parse(args)
}

func (genCommand *GenCommand) Run() error {
	if genCommand.configFile == defaultConfig {
		genCommand.configFile = genCommand.projectPath + "/" + genCommand.configFile
	}
	fmt.Println("Models:", genCommand.models)
	fmt.Println("project-path:", genCommand.projectPath)
	fmt.Println("config-file:", genCommand.configFile)

	return genCommand.app.GenegateAssets(genCommand.projectPath, genCommand.configFile, genCommand.models)
}
