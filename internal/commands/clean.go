package commands

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/go-teal/teal/internal/application"
)

func NewCleanCommand(app *application.Application) *CleanCommand {
	cleanCommand := &CleanCommand{
		fs:  flag.NewFlagSet("clean", flag.ContinueOnError),
		app: app,
	}

	cleanCommand.fs.StringVar(&cleanCommand.projectPath, "project-path", ".", "Project path dir")
	cleanCommand.fs.StringVar(&cleanCommand.models, "model", "", "models for cleaning")
	cleanCommand.fs.BoolVar(&cleanCommand.cleanMain, "clean-main", false, "delete main.go")

	return cleanCommand
}

type CleanCommand struct {
	fs          *flag.FlagSet
	models      string
	projectPath string
	cleanMain   bool
	app         *application.Application
}

func (cleanCommand *CleanCommand) Name() string {
	return cleanCommand.fs.Name()
}

func (cleanCommand *CleanCommand) Init(args []string) error {
	cleanCommand.projectPath = "."
	cleanCommand.models = "*"
	return cleanCommand.fs.Parse(args)
}

func (cleanCommand *CleanCommand) Run() error {
	fmt.Println("Models:", cleanCommand.models)
	fmt.Println("Dir:", cleanCommand.projectPath)
	fmt.Println("Clean Main:", cleanCommand.cleanMain)

	// config, err := cleanCommand.app.GetConfigService().GetConfig(cleanCommand.dir)
	// if err != nil {
	// 	panic(err)
	// }

	profile, err := cleanCommand.app.GetConfigService().GetProfileProfile(cleanCommand.projectPath)
	if err != nil {
		panic(err)
	}

	if cleanCommand.models == "*" {
		fmt.Print("Models are not specifed, are you sure you want to clean them all? [y/N] ")
		flag := "n"
		fmt.Scanln(&flag)
		if strings.ToLower(flag) != "y" {
			fmt.Printf("canceled\n")
		}
		// TODO: Delegate to the cleaner
		cleanErr := os.RemoveAll(cleanCommand.projectPath + "/internal/assets")
		if cleanErr != nil {
			fmt.Println(cleanErr)
		}

	}

	if cleanCommand.cleanMain {
		cleanErr := os.Remove(cleanCommand.projectPath + "/cmd/" + profile.Name + "/" + profile.Name + ".go")
		if cleanErr != nil {
			fmt.Println(cleanErr)
		}
	}

	return nil
}
