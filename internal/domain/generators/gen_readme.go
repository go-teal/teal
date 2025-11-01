package generators

import (
	_ "embed"
	"os"

	pongo2 "github.com/flosch/pongo2/v6"
	internalmodels "github.com/go-teal/teal/internal/domain/internal_models"
	"github.com/go-teal/teal/internal/domain/utils"
	"github.com/go-teal/teal/pkg/configs"
)

//go:embed templates/readme.md.tmpl
var readmeTemplate string

const README_FILENAME = "README.md"

type GenReadme struct {
	config        *configs.Config
	profile       *configs.ProjectProfile
	modelsConfigs []*internalmodels.ModelConfig
}

func InitGenReadme(config *configs.Config, profile *configs.ProjectProfile, modelsConfigs []*internalmodels.ModelConfig) Generator {
	return &GenReadme{
		config:        config,
		profile:       profile,
		modelsConfigs: modelsConfigs,
	}
}

// GetFileName implements Generator.
func (g *GenReadme) GetFileName() string {
	return README_FILENAME
}

// GetFullPath implements Generator.
func (g *GenReadme) GetFullPath() string {
	return g.config.ProjectPath + "/docs/" + README_FILENAME
}

func (g *GenReadme) RenderToFile() error {

	utils.CreateDir(g.config.ProjectPath + "/docs")

	// Build stages list
	stages := make([]struct {
		Name string
	}, len(g.profile.Models.Stages))
	for i, stage := range g.profile.Models.Stages {
		stages[i].Name = stage.Name
	}

	// Separate SQL and raw assets
	sqlAssets := []*internalmodels.ModelConfig{}
	rawAssets := []*internalmodels.ModelConfig{}

	for _, asset := range g.modelsConfigs {
		if asset.ModelType == internalmodels.DATABASE {
			sqlAssets = append(sqlAssets, asset)
		} else if asset.ModelType == internalmodels.SOURCE {
			rawAssets = append(rawAssets, asset)
		}
	}

	templ, err := pongo2.FromString(readmeTemplate)
	if err != nil {
		return err
	}

	output, err := templ.Execute(pongo2.Context{
		"ProjectName": g.profile.Name,
		"Module":      g.config.Module,
		"Version":     g.config.Version,
		"Stages":      stages,
		"Assets":      g.modelsConfigs,
		"RawAssets":   rawAssets,
		"Connections": g.config.Connections,
	})
	if err != nil {
		return err
	}

	file, err := os.Create(g.GetFullPath())
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(output)
	return err
}
