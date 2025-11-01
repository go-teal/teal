package commands

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-teal/teal/internal/application"
	"github.com/go-teal/teal/internal/domain/services"
)

func NewUICommand(app *application.Application) *UICommand {
	uiCommand := &UICommand{
		fs:  flag.NewFlagSet("ui", flag.ContinueOnError),
		app: app,
	}

	uiCommand.fs.IntVar(&uiCommand.port, "port", 8080, "Port for UI server")
	uiCommand.fs.StringVar(&uiCommand.logLevel, "log-level", "debug", "Log level (debug, info, warn, error)")
	uiCommand.fs.StringVar(&uiCommand.projectPath, "project-path", ".", "Project directory path")

	return uiCommand
}

type UICommand struct {
	fs          *flag.FlagSet
	port        int
	logLevel    string
	projectPath string
	app         *application.Application
}

func (uiCommand *UICommand) Name() string {
	return uiCommand.fs.Name()
}

func (uiCommand *UICommand) Init(args []string) error {
	return uiCommand.fs.Parse(args)
}

func (uiCommand *UICommand) Run() error {
	// Load profile to get project name
	configService := uiCommand.app.GetConfigService()
	profile, err := configService.GetProfileProfile(uiCommand.projectPath)
	if err != nil {
		return fmt.Errorf("failed to load profile.yaml: %w", err)
	}

	if profile.Name == "" {
		return fmt.Errorf("project name not found in profile.yaml")
	}

	fmt.Printf("Starting Teal UI observer\n")
	fmt.Printf("  Project: %s\n", profile.Name)
	fmt.Printf("  Port: %d\n", uiCommand.port)
	fmt.Printf("  Log level: %s\n", uiCommand.logLevel)
	fmt.Printf("  Project path: %s\n", uiCommand.projectPath)
	fmt.Printf("\n")

	// Create asset observer
	observer := services.NewAssetObserver(
		uiCommand.projectPath,
		uiCommand.port,
		uiCommand.logLevel,
		profile.Name,
		uiCommand.app,
	)

	// Start observer
	if err := observer.Start(); err != nil {
		return fmt.Errorf("failed to start observer: %w", err)
	}

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	fmt.Printf("Watching for changes... Press Ctrl+C to stop\n\n")

	// Wait for interrupt signal
	<-sigChan

	fmt.Printf("\nReceived shutdown signal\n")

	// Stop observer (this will also stop the UI process)
	if err := observer.Stop(); err != nil {
		fmt.Fprintf(os.Stderr, "Error during shutdown: %v\n", err)
		return err
	}

	fmt.Printf("Shutdown complete\n")
	return nil
}
