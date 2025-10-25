package generators

import (
	_ "embed"
	"os"

	pongo2 "github.com/flosch/pongo2/v6"
	internalmodels "github.com/go-teal/teal/internal/domain/internal_models"
	"github.com/go-teal/teal/internal/domain/utils"
	"github.com/go-teal/teal/pkg/configs"
)

//go:embed templates/dwh_raw_model_asset.tmpl
var dwhRawModelTemplate string

type GenRawModelAsset struct {
	config         *configs.Config
	projectProfile *configs.ProjectProfile
	modelConfig    *internalmodels.ModelConfig
}

func InitGenModelRawAsset(
	config *configs.Config,
	projectProfile *configs.ProjectProfile,
	modelConfig *internalmodels.ModelConfig,

) Generator {
	return &GenRawModelAsset{
		config:         config,
		projectProfile: projectProfile,
		modelConfig:    modelConfig,
	}
}

// GetFileName implements Generator.
func (g *GenRawModelAsset) GetFileName() string {
	return g.modelConfig.ModelName
}

// GetFullPath implements Generator.
func (g *GenRawModelAsset) GetFullPath() string {
	return g.config.ProjectPath + "/internal/assets/" + g.GetFileName() + ".go"
}

func (g *GenRawModelAsset) RenderToFile() error {

	dirName := g.config.ProjectPath + "/internal/assets/"
	utils.CreateDir(dirName)

	goTempl, err := pongo2.FromString(dwhRawModelTemplate)
	if err != nil {
		return err
	}

	g.modelConfig.ModelFieldsFunc = "{{ ModelFields }}"

	output, err := goTempl.Execute(pongo2.Context{
		"ModelName":    g.modelConfig.ModelName,
		"GoName":       g.modelConfig.GoName,
		"ModelProfile": g.modelConfig.ModelProfile,
		"Upstreams":    g.modelConfig.Upstreams,
		"Downstreams":  g.modelConfig.Downstreams,
	})
	if err != nil {
		return err
	}

	file, err := os.Create(g.GetFullPath())

	if err != nil {
		panic(err)
	}

	defer file.Close()

	_, err = file.WriteString(output)
	return err
}
