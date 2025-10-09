package main

import (

	_ "github.com/marcboeker/go-duckdb/v2"


	"context"
	"flag"
	"os"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	modeltests "github.com/you_git_user/your_project/internal/model_tests"
	"github.com/go-teal/teal/pkg/core"
	"github.com/go-teal/teal/pkg/dags"
	"github.com/go-teal/teal/pkg/services/logwriter"
	"github.com/go-teal/teal/pkg/ui"
	"github.com/you_git_user/your_project/internal/assets"
)

func main() {
	// Parse command line flags
	port := flag.Int("port", 8080, "Port for debug UI server")
	logOutput := flag.String("log-output", "raw", "Log output format: json or raw")
	logLevel := flag.String("log-level", "info", "Log level: panic, fatal, error, warn, info, debug, trace")
	flag.Parse()

	// Create a context for the application
	ctx := context.Background()
	
	// Always use StoringConsoleWriter for UI mode to capture logs per task
	storingWriter := logwriter.NewStoringConsoleWriter(ctx, os.Stderr)
	
	// Configure output format
	if *logOutput == "json" {
		storingWriter.SetNoColor(true)
		storingWriter.SetTimeFormat("")
	}
	
	// Set the global logger to use our storing writer
	log.Logger = log.Output(storingWriter)

	// Set log level
	logLevels := map[string]zerolog.Level{
		"panic": zerolog.PanicLevel,
		"fatal": zerolog.FatalLevel,
		"error": zerolog.ErrorLevel,
		"warn":  zerolog.WarnLevel,
		"info":  zerolog.InfoLevel,
		"debug": zerolog.DebugLevel,
		"trace": zerolog.TraceLevel,
	}
	
	if level, ok := logLevels[*logLevel]; ok {
		zerolog.SetGlobalLevel(level)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	log.Info().Int("port", *port).Msg("Starting hello-world in UI debug mode")
	
	// Initialize core
	core.GetInstance().Init("config.yaml", ".")
	config := core.GetInstance().Config
	
	// Create DebugDag for UI mode
	dag := dags.InitDebugDag(assets.DAG, assets.ProjectAssets, modeltests.ProjectTests, config, "hello-world")
	
	// Start UI server with DebugDag and log writer
	server := ui.NewUIServerWithLogWriter("hello-world", "github.com/you_git_user/your_project", *port, dag, storingWriter)
	if err := server.Start(); err != nil {
		log.Fatal().Err(err).Msg("Failed to start UI server")
	}
}