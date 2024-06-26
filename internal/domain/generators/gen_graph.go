package generators

import (
	_ "embed"
	"os"
	"text/template"

	internalmodels "github.com/go-teal/teal/internal/domain/internal_models"
	"github.com/go-teal/teal/internal/domain/utils"
	"github.com/go-teal/teal/pkg/configs"
)

//go:embed templates/graph.wsd.tmpl
var graphTemplate string

const GRAPH_FILENAME = "graph.wsd"

type GenGraph struct {
	config        *configs.Config
	profile       *configs.Profile
	modelsConfigs []*internalmodels.ModelConfig
}

func InitGenGraph(config *configs.Config, profile *configs.Profile, modelsConfigs []*internalmodels.ModelConfig) Generator {
	return &GenGraph{
		config:        config,
		profile:       profile,
		modelsConfigs: modelsConfigs,
	}
}

// GetFileName implements Generator.
func (g *GenGraph) GetFileName() string {
	return GRAPH_FILENAME
}

// GetFullPath implements Generator.
func (g *GenGraph) GetFullPath() string {
	return g.config.ProjectPath + "/docs/" + GRAPH_FILENAME
}

func (g *GenGraph) RenderToFile() error {

	utils.CreateDir(g.config.ProjectPath + "/docs")
	stages := make([]string, len(g.profile.Models.Stages))
	for i, stage := range g.profile.Models.Stages {
		stages[i] = stage.Name
	}

	templ, err := template.New(GRAPH_FILENAME).Parse(graphTemplate)
	if err != nil {
		panic(err)
	}
	file, err := os.Create(g.GetFullPath())

	if err != nil {
		panic(err)
	}

	defer file.Close()

	data := struct {
		ProjectName string
		Stages      []string
		Assets      []*internalmodels.ModelConfig
	}{
		ProjectName: g.profile.Name,
		Stages:      stages,
		Assets:      g.modelsConfigs,
	}
	err = templ.Execute(file, data)
	return err
}
