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

func getHelpMessage() string {
	return `
Usage:
	teal [command] [flags]

Commands:
	init      Create basic teal project structure
	          No flags required

	gen       Generate Go code from asset model files
	          Flags:
	            --project-path string    Project directory (default ".")
	            --config-file string     Path to config.yaml (default "config.yaml")
	            --model string          Name of target model (optional)

	clean     Clean generated files
	          Flags:
	            --project-path string    Project directory (default ".")
	            --model string          Models for cleaning (default "*")
	            --clean-main            Delete production main.go
	            --clean-main-ui         Delete UI debug main.go
	            --clean-dockerfile      Delete Dockerfile
	            --clean-go-mod          Delete go.mod and go.sum
	            --clean-all             Delete ALL generated files

	ui        Start UI development server with hot-reload
	          Flags:
	            --port int              Port for UI server (default 8080)
	            --log-level string      Log level: debug, info, warn, error (default "debug")
	            --project-path string    Project directory (default ".")

	version   Show teal version
	          No flags required

Global Flags:
	--help, -h    Show this help message

Examples:
	teal init
	teal gen --project-path ./my-project
	teal clean --clean-main
	teal ui --port 9090
	teal version

For more information, visit: https://github.com/go-teal/teal
`
}

func getCommandHelp(command string) string {
	helpTexts := map[string]string{
		"init": `
Usage: teal init

Creates basic teal project structure with default configuration files.

This command initializes a new Teal project with:
  - config.yaml (database connections)
  - profile.yaml (project configuration)
  - assets/ directory structure

No flags required.
`,
		"gen": `
Usage: teal gen [flags]

Generates Go code from asset model files.

Flags:
  --project-path string    Project directory (default ".")
  --config-file string     Path to config.yaml (default "config.yaml")
  --model string          Name of target model to generate (optional, generates all if not specified)

Examples:
  teal gen
  teal gen --project-path ./my-project
  teal gen --model staging.customers
  teal gen --config-file custom-config.yaml
`,
		"clean": `
Usage: teal clean [flags]

Cleans generated files from the project.

Flags:
  --project-path string    Project directory (default ".")
  --model string          Models for cleaning (default "*" for all)
  --clean-main            Delete production main.go in cmd/<project-name>/
  --clean-main-ui         Delete UI debug main.go in cmd/<project-name>-ui/
  --clean-dockerfile      Delete Dockerfile
  --clean-go-mod          Delete go.mod and go.sum
  --clean-all             Delete ALL generated files (prompts for confirmation)

Examples:
  teal clean                               # Clean all models (prompts for confirmation)
  teal clean --model staging.customers     # Clean specific model
  teal clean --clean-main                  # Clean production main.go only
  teal clean --clean-main-ui               # Clean UI main.go only
  teal clean --clean-dockerfile            # Clean Dockerfile only
  teal clean --clean-go-mod                # Clean go.mod and go.sum
  teal clean --clean-all                   # Clean everything (prompts for confirmation)
  teal clean --project-path ./my-project   # Clean in specific directory

Note:
- When cleaning all models (*), you will be prompted for confirmation
- --clean-all will delete ALL generated files including go.mod, Dockerfile, and main files
- Files not overwritten during 'teal gen': Dockerfile, go.mod, production main.go, UI main.go
`,
		"ui": `
Usage: teal ui [flags]

Starts UI development server with hot-reload for debugging and monitoring.

Flags:
  --port int              Port for UI server (default 8080)
  --log-level string      Log level: debug, info, warn, error (default "debug")
  --project-path string    Project directory (default ".")

Examples:
  teal ui
  teal ui --port 9090
  teal ui --log-level info
  teal ui --project-path ./my-project

The UI provides:
  - DAG visualization
  - Real-time execution monitoring
  - Test result tracking
  - Log viewing
  - Data inspection
`,
		"version": `
Usage: teal version

Shows the current version of Teal CLI.

No flags required.
`,
	}

	if help, ok := helpTexts[command]; ok {
		return help
	}
	return fmt.Sprintf("No detailed help available for command: %s\n", command)
}

func root(args []string) error {
	if len(args) < 1 {
		return errors.New(getHelpMessage())
	}

	// Handle --help or -h flag
	if args[0] == "--help" || args[0] == "-h" {
		fmt.Println(getHelpMessage())
		os.Exit(0)
	}

	app := application.InitApplication()
	cmds := []Runner{
		commands.NewGenCommand(app),
		commands.NewCleanCommand(app),
		commands.NewVersionCommand(app),
		commands.NewInitCommand(app),
		commands.NewUICommand(app),
	}

	subcommand := os.Args[1]

	for _, cmd := range cmds {
		if cmd.Name() == subcommand {
			// Check if user wants help for this specific command
			if len(os.Args) > 2 && (os.Args[2] == "--help" || os.Args[2] == "-h") {
				// Initialize without running to trigger flag parsing help
				if err := cmd.Init([]string{"--help"}); err != nil {
					// Flag package prints help and returns ErrHelp
					fmt.Println(getCommandHelp(subcommand))
				}
				os.Exit(0)
			}
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
