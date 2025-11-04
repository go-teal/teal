package generators

import (
	_ "embed"
	"os"
	"strings"

	pongo2 "github.com/flosch/pongo2/v6"
	internalmodels "github.com/go-teal/teal/internal/domain/internal_models"
	"github.com/go-teal/teal/internal/domain/utils"
	"github.com/go-teal/teal/pkg/configs"
)

//go:embed templates/graph.mmd.tmpl
var graphTemplate string

const GRAPH_FILENAME = "graph.mmd"

type GenGraph struct {
	config        *configs.Config
	profile       *configs.ProjectProfile
	modelsConfigs []*internalmodels.ModelConfig
}

func InitGenGraph(config *configs.Config, profile *configs.ProjectProfile, modelsConfigs []*internalmodels.ModelConfig) Generator {
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

type GraphNode struct {
	NodeID          string
	ModelName       string
	Stage           string
	Materialization string
	Downstreams     []string
	DownstreamIDs   []string
}

func sanitizeNodeID(name string) string {
	return strings.ReplaceAll(name, ".", "_")
}

func (g *GenGraph) RenderToFile() (error, bool) {

	utils.CreateDir(g.config.ProjectPath + "/docs")
	stages := make([]string, len(g.profile.Models.Stages))
	for i, stage := range g.profile.Models.Stages {
		stages[i] = stage.Name
	}

	// Create graph nodes with sanitized IDs
	nodes := make([]*GraphNode, len(g.modelsConfigs))
	for i, model := range g.modelsConfigs {
		downstreamIDs := make([]string, len(model.Downstreams))
		for j, ds := range model.Downstreams {
			downstreamIDs[j] = sanitizeNodeID(ds)
		}

		materialization := ""
		if model.ModelProfile != nil {
			materialization = string(model.ModelProfile.Materialization)
		}

		nodes[i] = &GraphNode{
			NodeID:          sanitizeNodeID(model.ModelName),
			ModelName:       model.ModelName,
			Stage:           model.Stage,
			Materialization: materialization,
			Downstreams:     model.Downstreams,
			DownstreamIDs:   downstreamIDs,
		}
	}

	templ, err := pongo2.FromString(graphTemplate)
	if err != nil {
		panic(err)
	}

	output, err := templ.Execute(pongo2.Context{
		"ProjectName": g.profile.Name,
		"Stages":      stages,
		"Nodes":       nodes,
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
	return err, false
}
