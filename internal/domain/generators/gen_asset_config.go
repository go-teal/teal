package generators

import (
	_ "embed"
	"os"
	"text/template"

	internalmodels "github.com/go-teal/teal/internal/domain/internal_models"
	"github.com/go-teal/teal/internal/domain/utils"
	"github.com/go-teal/teal/pkg/configs"
)

//go:embed templates/configs.go.tmpl
var goQuerySetConfigGlobalDeclarationTemplate string

const GO_ASSETS_CONFIG_FILE_NAME = "configs.go"

type GenAssetsConfig struct {
	config         *configs.Config
	profile        *configs.Profile
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
	profile *configs.Profile,
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
	templ, err := template.New(GO_ASSETS_CONFIG_FILE_NAME).Parse(goQuerySetConfigGlobalDeclarationTemplate)
	if err != nil {
		panic(err)
	}
	file, err := os.Create(g.GetFullPath())

	if err != nil {
		panic(err)
	}

	defer file.Close()

	data := struct {
		Config         *configs.Config
		Assets         []*internalmodels.ModelConfig
		PriorityGroups [][]string
	}{
		Config:         g.config,
		Assets:         g.modelsConfig,
		PriorityGroups: g.priorityGroups,
	}
	err = templ.Execute(file, data)
	return err
}
