package application

import (
	"fmt"
	"strings"

	"github.com/go-teal/teal/internal/domain/generators"
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
	profile, err := app.configService.GetProfile(pojectPath)
	if err != nil {
		return err
	}

	var generatorsList []generators.Generator = []generators.Generator{
		generators.InitGenMain(config, profile),
		generators.InitGenGoMod(config, profile),
	}

	modelConfigs, err := services.InitSQLModelConfigs(config, profile)
	if err != nil {
		fmt.Printf("can not create a configuration for models %v\n", err)
		return err
	}

	for _, modelConfig := range modelConfigs {
		fmt.Printf("%s <- %v\n", modelConfig.ModelName, modelConfig.Upstreams)
		generatorsList = append(generatorsList, generators.InitGenModelAsset(config, profile, modelConfig))
	}
	priorityGroups := app.dependnacyGraph.Build(modelConfigs)
	fmt.Println(priorityGroups)

	for _, modelConfig := range modelConfigs {
		fmt.Printf("%s -> %v\n", modelConfig.ModelName, modelConfig.Downstreams)
		// generatorsList = append(generatorsList, generators.InitGenModelAsset(config, profile, modelConfig))
	}

	generatorsList = append(generatorsList, generators.InitGenAssetsConfig(config, profile, modelConfigs, priorityGroups))
	generatorsList = append(generatorsList, generators.InitGenGraph(config, profile, modelConfigs))

	fmt.Printf("Files %d\n", len(generatorsList))

	for _, g := range generatorsList {
		err := g.RenderToFile()
		if err != nil {
			fmt.Printf("\033[39m%s %s \033[91m[FAIL]\n%v\n", g.GetFullPath(), strings.Repeat(".", 70-len(g.GetFullPath())), err)
		} else {
			fmt.Printf("\033[39m%s %s \033[92m[OK]\n", g.GetFullPath(), strings.Repeat(".", 70-len(g.GetFullPath())))
		}
	}

	return nil
}
