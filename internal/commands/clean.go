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
	cleanCommand.fs.BoolVar(&cleanCommand.cleanMain, "clean-main", false, "delete production main.go")
	cleanCommand.fs.BoolVar(&cleanCommand.cleanMainUI, "clean-main-ui", false, "delete UI debug main.go")
	cleanCommand.fs.BoolVar(&cleanCommand.cleanDockerfile, "clean-dockerfile", false, "delete Dockerfile")
	cleanCommand.fs.BoolVar(&cleanCommand.cleanGoMod, "clean-go-mod", false, "delete go.mod and go.sum")
	cleanCommand.fs.BoolVar(&cleanCommand.cleanAll, "clean-all", false, "delete all generated files")

	return cleanCommand
}

type CleanCommand struct {
	fs               *flag.FlagSet
	models           string
	projectPath      string
	cleanMain        bool
	cleanMainUI      bool
	cleanDockerfile  bool
	cleanGoMod       bool
	cleanAll         bool
	app              *application.Application
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

	profile, err := cleanCommand.app.GetConfigService().GetProfileProfile(cleanCommand.projectPath)
	if err != nil {
		panic(err)
	}

	// Handle --clean-all flag
	if cleanCommand.cleanAll {
		fmt.Print("This will delete ALL generated files including go.mod, Dockerfile, and main files. Are you sure? [y/N] ")
		flag := "n"
		fmt.Scanln(&flag)
		if strings.ToLower(flag) != "y" {
			fmt.Printf("Canceled\n")
			return nil
		}

		// Set all clean flags to true
		cleanCommand.models = "*"
		cleanCommand.cleanMain = true
		cleanCommand.cleanMainUI = true
		cleanCommand.cleanDockerfile = true
		cleanCommand.cleanGoMod = true
	}

	// Determine if user wants to clean models
	cleanModels := cleanCommand.models == "*" || cleanCommand.models != ""

	// Determine if any specific file flags are set
	hasSpecificFileFlags := cleanCommand.cleanMain || cleanCommand.cleanMainUI ||
		cleanCommand.cleanDockerfile || cleanCommand.cleanGoMod

	// Clean models
	if cleanModels {
		shouldCleanModels := false

		if cleanCommand.cleanAll {
			// cleanAll is set, clean without asking
			shouldCleanModels = true
		} else if !hasSpecificFileFlags {
			// No specific file flags, ask for confirmation
			fmt.Print("Models are not specified, are you sure you want to clean them all? [y/N] ")
			flag := "n"
			fmt.Scanln(&flag)
			if strings.ToLower(flag) == "y" {
				shouldCleanModels = true
			} else {
				fmt.Printf("Model cleaning canceled\n")
			}
		} else {
			// Specific file flags are set, ask for confirmation
			fmt.Print("Do you want to clean models? [y/N] ")
			flag := "n"
			fmt.Scanln(&flag)
			if strings.ToLower(flag) == "y" {
				shouldCleanModels = true
			}
		}

		if shouldCleanModels {
			fmt.Println("Cleaning internal/assets...")
			cleanErr := os.RemoveAll(cleanCommand.projectPath + "/internal/assets")
			if cleanErr != nil {
				fmt.Fprintf(os.Stderr, "Error cleaning assets: %v\n", cleanErr)
			}

			fmt.Println("Cleaning internal/model_tests...")
			cleanErr = os.RemoveAll(cleanCommand.projectPath + "/internal/model_tests")
			if cleanErr != nil {
				fmt.Fprintf(os.Stderr, "Error cleaning model_tests: %v\n", cleanErr)
			}

			fmt.Println("Cleaning docs/...")
			cleanErr = os.Remove(cleanCommand.projectPath + "/docs/graph.mmd")
			if cleanErr != nil && !os.IsNotExist(cleanErr) {
				fmt.Fprintf(os.Stderr, "Error cleaning graph.mmd: %v\n", cleanErr)
			}
			cleanErr = os.Remove(cleanCommand.projectPath + "/docs/README.md")
			if cleanErr != nil && !os.IsNotExist(cleanErr) {
				fmt.Fprintf(os.Stderr, "Error cleaning docs/README.md: %v\n", cleanErr)
			}
		}
	}

	// Clean production main.go
	if cleanCommand.cleanMain {
		fmt.Printf("Cleaning production main.go: cmd/%s/%s.go\n", profile.Name, profile.Name)
		cleanErr := os.Remove(cleanCommand.projectPath + "/cmd/" + profile.Name + "/" + profile.Name + ".go")
		if cleanErr != nil && !os.IsNotExist(cleanErr) {
			fmt.Fprintf(os.Stderr, "Error: %v\n", cleanErr)
		}

		cleanErr = os.RemoveAll(cleanCommand.projectPath + "/cmd/" + profile.Name)
		if cleanErr != nil && !os.IsNotExist(cleanErr) {
			fmt.Fprintf(os.Stderr, "Error: %v\n", cleanErr)
		}
	}

	// Clean UI main.go
	if cleanCommand.cleanMainUI {
		fmt.Printf("Cleaning UI main.go: cmd/%s-ui/%s-ui.go\n", profile.Name, profile.Name)
		cleanErr := os.Remove(cleanCommand.projectPath + "/cmd/" + profile.Name + "-ui/" + profile.Name + "-ui.go")
		if cleanErr != nil && !os.IsNotExist(cleanErr) {
			fmt.Fprintf(os.Stderr, "Error: %v\n", cleanErr)
		}

		cleanErr = os.RemoveAll(cleanCommand.projectPath + "/cmd/" + profile.Name + "-ui")
		if cleanErr != nil && !os.IsNotExist(cleanErr) {
			fmt.Fprintf(os.Stderr, "Error: %v\n", cleanErr)
		}
	}

	// Clean Dockerfile
	if cleanCommand.cleanDockerfile {
		fmt.Println("Cleaning Dockerfile...")
		cleanErr := os.Remove(cleanCommand.projectPath + "/Dockerfile")
		if cleanErr != nil && !os.IsNotExist(cleanErr) {
			fmt.Fprintf(os.Stderr, "Error: %v\n", cleanErr)
		} else if cleanErr == nil {
			fmt.Println("Dockerfile deleted")
		}
	}

	// Clean go.mod and go.sum
	if cleanCommand.cleanGoMod {
		fmt.Println("Cleaning go.mod and go.sum...")
		cleanErr := os.Remove(cleanCommand.projectPath + "/go.mod")
		if cleanErr != nil && !os.IsNotExist(cleanErr) {
			fmt.Fprintf(os.Stderr, "Error cleaning go.mod: %v\n", cleanErr)
		} else if cleanErr == nil {
			fmt.Println("go.mod deleted")
		}

		cleanErr = os.Remove(cleanCommand.projectPath + "/go.sum")
		if cleanErr != nil && !os.IsNotExist(cleanErr) {
			fmt.Fprintf(os.Stderr, "Error cleaning go.sum: %v\n", cleanErr)
		} else if cleanErr == nil {
			fmt.Println("go.sum deleted")
		}
	}

	fmt.Println("Clean completed")
	return nil
}
