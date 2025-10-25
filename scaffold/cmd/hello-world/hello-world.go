package main

import (

	_ "github.com/marcboeker/go-duckdb/v2"



	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	modeltests "github.com/you_git_user/your_project/internal/model_tests"
	"github.com/go-teal/teal/pkg/core"
	"github.com/go-teal/teal/pkg/dags"
	"github.com/you_git_user/your_project/internal/assets"
)

func main() {
	// Parse command line flags
	inputData := flag.String("input-data", "", "Input data in JSON format (optional)")
	logOutput := flag.String("log-output", "json", "Log output format: json or raw")
	logLevel := flag.String("log-level", "debug", "Log level: panic, fatal, error, warn, info, debug, trace")
	withTests := flag.Bool("with-tests", true, "Run with tests")
	customTaskName := flag.String("task-name", "", "Custom task name (optional, auto-generated if not provided)")
	flag.Parse()

	// Configure logger based on log output format
	if *logOutput == "raw" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	} else {
		log.Logger = log.Output(os.Stderr)
	}

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

	log.Info().Msg("Starting hello-world")
	core.GetInstance().Init("config.yaml", ".")
	core.GetInstance().ConnectAll()
	defer core.GetInstance().Shutdown()
	config := core.GetInstance().Config

	// Parse input data if provided
	var inputDataMap map[string]interface{}
	if *inputData != "" {
		if err := json.Unmarshal([]byte(*inputData), &inputDataMap); err != nil {
			log.Fatal().Err(err).Msg("Failed to parse input data JSON")
		}
	}

	// Generate unique task ID with timestamp or use custom name
	var taskId string
	if *customTaskName != "" {
		taskId = *customTaskName
		log.Info().Str("taskId", taskId).Msg("Using custom task ID")
	} else {
		taskId = fmt.Sprintf("hello-world_%d", time.Now().Unix())
		log.Info().Str("taskId", taskId).Msg("Generated task ID")
	}

	// Initialize DAG with or without tests based on flag
	var dag dags.DAG
	if *withTests {
		dag = dags.InitChannelDagWithTests(assets.DAG, assets.ProjectAssets, modeltests.ProjectTests, config, taskId)
	} else {
		dag = dags.InitChannelDag(assets.DAG, assets.ProjectAssets, config, taskId)
	}

	wg := dag.Run()
	result := <-dag.Push(taskId, inputDataMap, make(chan map[string]interface{}))
	log.Info().Str("taskId", taskId).Any("Result", result).Send()
	dag.Stop()
	wg.Wait()

	log.Info().Msg("Finishing hello-world")
}