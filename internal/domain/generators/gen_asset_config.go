package generators

import (
	_ "embed"
	"os"
	"sort"

	pongo2 "github.com/flosch/pongo2/v6"
	internalmodels "github.com/go-teal/teal/internal/domain/internal_models"
	"github.com/go-teal/teal/internal/domain/utils"
	"github.com/go-teal/teal/pkg/configs"
)

//go:embed templates/configs.go.tmpl
var goQuerySetConfigGlobalDeclarationTemplate string

const GO_ASSETS_CONFIG_FILE_NAME = "configs.go"

type GenAssetsConfig struct {
	config         *configs.Config
	profile        *configs.ProjectProfile
	modelsConfig   []*internalmodels.ModelConfig
	priorityGroups [][]string
}

// GetFileName implements Generator.
func (g *GenAssetsConfig) GetFileName() string {
	return GO_ASSETS_CONFIG_FILE_NAME
}

// GetFullPath implements Generator.
func (g *GenAssetsConfig) GetFullPath() string {
	return g.config.ProjectPath + "/internal/assets/" + GO_ASSETS_CONFIG_FILE_NAME
}

func InitGenAssetsConfig(
	config *configs.Config,
	profile *configs.ProjectProfile,
	modelsConfig []*internalmodels.ModelConfig,
	priorityGroups [][]string,
) Generator {
	return &GenAssetsConfig{
		config:         config,
		profile:        profile,
		modelsConfig:   modelsConfig,
		priorityGroups: priorityGroups,
	}
}

func (g *GenAssetsConfig) RenderToFile() error {
	// fmt.Printf("Rendering: %s", g.GetFullPath())
	dirName := g.config.ProjectPath + "/internal/assets/"
	utils.CreateDir(dirName)
	templ, err := pongo2.FromString(goQuerySetConfigGlobalDeclarationTemplate)
	if err != nil {
		panic(err)
	}

	// Sort assets alphabetically by ModelName
	sortedAssets := make([]*internalmodels.ModelConfig, len(g.modelsConfig))
	copy(sortedAssets, g.modelsConfig)
	sort.Slice(sortedAssets, func(i, j int) bool {
		return sortedAssets[i].ModelName < sortedAssets[j].ModelName
	})

	// Sort model names within each priority group
	sortedPriorityGroups := make([][]string, len(g.priorityGroups))
	for i, group := range g.priorityGroups {
		sortedGroup := make([]string, len(group))
		copy(sortedGroup, group)
		sort.Strings(sortedGroup)
		sortedPriorityGroups[i] = sortedGroup
	}

	output, err := templ.Execute(pongo2.Context{
		"Config":         g.config,
		"Assets":         sortedAssets,
		"PriorityGroups": sortedPriorityGroups,
	})
	if err != nil {
		panic(err)
	}

	file, err := os.Create(g.GetFullPath())

	if err != nil {
		panic(err)
	}

	defer file.Close()

	_, err = file.WriteString(output)
	return err
}
