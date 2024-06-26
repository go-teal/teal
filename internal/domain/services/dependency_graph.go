package services

import (
	"fmt"

	internalmodels "github.com/go-teal/teal/internal/domain/internal_models"
)

type DependnacyGraph struct {
	// config  *entities.Config
	// profile *entities.Profile
}

func InitDependnacyGraph() *DependnacyGraph {
	return &DependnacyGraph{
		// config:  config,
		// profile: profile,
	}
}

func (dg *DependnacyGraph) Build(modelsConfigs []*internalmodels.ModelConfig) [][]string {
	modelConfigMap := make(map[string]*internalmodels.ModelConfig, len(modelsConfigs))
	priorityMap := make(map[string]int, len(modelsConfigs))
	for _, modelConfig := range modelsConfigs {
		modelConfigMap[modelConfig.ModelName] = modelConfig
		priorityMap[modelConfig.ModelName] = -1
	}

	for ref, modelConfig := range modelConfigMap {
		for _, upstreamName := range modelConfig.Upstreams {
			modelConfigMap[upstreamName].Downstreams = append(modelConfigMap[upstreamName].Downstreams, ref)
		}
	}
	maxPriority := 0
	for _, modelConfig := range modelsConfigs {
		propagatePriotiry(modelConfig.ModelName, 0, modelConfigMap, priorityMap, &maxPriority)
	}
	fmt.Println("maxPriority:", maxPriority)
	priorityGroups := make([][]string, maxPriority+1)
	for modelName, priority := range priorityMap {
		priorityGroups[priority] = append(priorityGroups[priority], modelName)
	}
	return priorityGroups
}

func propagatePriotiry(
	modelName string,
	newPriority int,
	modelConfigs map[string]*internalmodels.ModelConfig,
	priorityMap map[string]int,
	maxPriority *int) {
	if newPriority > priorityMap[modelName] {
		if newPriority > *maxPriority {
			*maxPriority = newPriority
		}
		priorityMap[modelName] = newPriority
		modelConfigs[modelName].Priority = newPriority
		for _, downstreamName := range modelConfigs[modelName].Downstreams {
			propagatePriotiry(downstreamName, newPriority+1, modelConfigs, priorityMap, maxPriority)
		}
	}
}
