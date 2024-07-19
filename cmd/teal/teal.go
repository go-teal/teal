package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/go-teal/teal/internal/application"
	"github.com/go-teal/teal/internal/commands"
)

type Runner interface {
	Init([]string) error
	Run() error
	Name() string
}

func root(args []string) error {
	if len(args) < 1 {
		message := `
Usage:
	teal [command]

Commands:
	init 	creates basic teal project structure
	gen		generates GO code from asset model files
		`
		return errors.New(message)
	}

	app := application.InitApplication()
	cmds := []Runner{
		commands.NewGenCommand(app),
		commands.NewCleanCommand(app),
		commands.NewVersionCommand(app),
		commands.NewInitCommand(app),
	}

	subcommand := os.Args[1]

	for _, cmd := range cmds {
		if cmd.Name() == subcommand {
			cmd.Init(os.Args[2:])
			return cmd.Run()
		}
	}

	return fmt.Errorf("unknown subcommand: %s", subcommand)
}

func main() {
	if err := root(os.Args[1:]); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
