package services

import (
	internalmodels "github.com/go-teal/teal/internal/domain/internal_models"
	"github.com/go-teal/teal/internal/domain/utils"
	"github.com/go-teal/teal/pkg/configs"
)

func InitRawModelConfig(config *configs.Config, profiles *configs.ProjectProfile) ([]*internalmodels.ModelConfig, error) {

	var modelsConfigs []*internalmodels.ModelConfig

	for _, stage := range profiles.Models.Stages {
		for _, model := range stage.Models {
			if model.Materialization != configs.MAT_RAW {
				continue
			}
			goModelName, refName := utils.CreateModelName(stage.Name, model.Name)
			model.Stage = stage.Name
			modelsConfigs = append(modelsConfigs, &internalmodels.ModelConfig{
				Stage:        stage.Name,
				GoName:       goModelName,
				ModelName:    refName,
				ModelType:    internalmodels.SOURCE,
				Upstreams:    model.RawUpstreams,
				Config:       config,
				Profile:      profiles,
				ModelProfile: model,
			})
		}
	}

	return modelsConfigs, nil
}
