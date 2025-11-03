package application

import (
	"fmt"
	"strings"

	"github.com/go-teal/teal/internal/domain/generators"
	internalmodels "github.com/go-teal/teal/internal/domain/internal_models"
	"github.com/go-teal/teal/internal/domain/services"
	"github.com/go-teal/teal/pkg/configs"
)

func (app *Application) GetConfigService() *configs.ConfigService {
	return app.configService
}

func (app *Application) GenegateAssets(pojectPath string, configFilePath string, models string) error {

	config, err := app.configService.GetConfig(configFilePath, pojectPath)
	if err != nil {
		return err
	}
	projectProfile, err := app.configService.GetProfileProfile(pojectPath)
	if err != nil {
		return err
	}

	var generatorsList []generators.Generator = []generators.Generator{
		generators.InitGenMain(config, projectProfile),   // Production main.go
		generators.InitGenMainUI(config, projectProfile), // UI debugging main.go
		generators.InitGenGoMod(config, projectProfile),
		generators.InitGenMakefile(config, projectProfile),   // Makefile
		generators.InitGenDockerfile(config, projectProfile), // Dockerfile
	}

	services.CombineProfiles(config, projectProfile)
	modelConfigs, err := services.InitSQLModelConfigs(config, projectProfile)
	if err != nil {
		fmt.Printf("can not create a configuration for SQL models %v\n", err)
		return err
	}

	modelConfigs2, err := services.InitRawModelConfig(config, projectProfile)
	if err != nil {
		fmt.Printf("can not create a configuration for RAW models %v\n", err)
		return err
	}

	modelConfigs = append(modelConfigs, modelConfigs2...)

	for _, modelConfig := range modelConfigs {
		fmt.Printf("%s <- %v\n", modelConfig.ModelName, modelConfig.Upstreams)
		switch modelConfig.ModelType {
		case internalmodels.DATABASE:
			generatorsList = append(generatorsList, generators.InitGenModelSQLAsset(config, projectProfile, modelConfig))
		case internalmodels.SOURCE:
			generatorsList = append(generatorsList, generators.InitGenModelRawAsset(config, projectProfile, modelConfig))
		default:
			panic("unknown model type")
		}
		// generatorsList = append(generatorsList, generators.InitGenModelAsset(config, projectProfile, modelConfig))
	}
	priorityGroups := app.dependnacyGraph.Build(modelConfigs)
	fmt.Println(priorityGroups)

	for _, modelConfig := range modelConfigs {
		fmt.Printf("%s -> %v\n", modelConfig.ModelName, modelConfig.Downstreams)
		// generatorsList = append(generatorsList, generators.InitGenModelAsset(config, profile, modelConfig))
	}

	testConfigs, err := services.InitSQLTestsConfigs(config, projectProfile)

	for _, testConfig := range testConfigs {
		generatorsList = append(generatorsList, generators.InitGenSQLModelTest(config, projectProfile, testConfig))
	}

	generatorsList = append(generatorsList, generators.InitGenAssetsConfig(config, projectProfile, modelConfigs, priorityGroups))
	generatorsList = append(generatorsList, generators.InitGenGraph(config, projectProfile, modelConfigs))
	generatorsList = append(generatorsList, generators.InitGenReadme(config, projectProfile, modelConfigs))
	generatorsList = append(generatorsList, generators.InitGenTestConfig(config, projectProfile, testConfigs))

	fmt.Printf("Files %d\n", len(generatorsList))

	for _, g := range generatorsList {
		err, skipped := g.RenderToFile()
		if err != nil {
			fmt.Printf("\033[39m%s %s \033[91m[FAIL]\n%v\n", g.GetFullPath(), strings.Repeat(".", 70-len(g.GetFullPath())), err)
		} else if skipped {
			fmt.Printf("\033[39m%s %s [SKIPPED]\n", g.GetFullPath(), strings.Repeat(".", 70-len(g.GetFullPath())))
		} else {
			fmt.Printf("\033[39m%s %s \033[92m[OK]\n", g.GetFullPath(), strings.Repeat(".", 70-len(g.GetFullPath())))
		}
	}

	return nil
}
