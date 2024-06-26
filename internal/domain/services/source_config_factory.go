package services

import (
	"fmt"
	"strings"

	internalmodels "github.com/go-teal/teal/internal/domain/internal_models"
	"github.com/go-teal/teal/pkg/configs"
)

func InitSourceModelConfigs(config *configs.Config, profiles *configs.Profile) ([]*internalmodels.ModelConfig, error) {
	var modelConfigs []*internalmodels.ModelConfig

	// iterate over sources
	for _, source := range profiles.Sources {
		for _, tableName := range source.Tables {
			modelConfig := &internalmodels.ModelConfig{
				NameUpperCase: fmt.Sprintf("SRC_%s_%s", strings.ToUpper(source.Name), strings.ToUpper(tableName)),
				GoName:        fmt.Sprintf("source_%s_%s", source.Name, tableName),
				ModelType:     internalmodels.SOURCE,
				SourceProfile: &source,
				Stage:         strings.ToLower(source.Name),
				ModelName:     fmt.Sprintf("%s.%s", strings.ToLower(source.Name), strings.ToLower(tableName)),
				Config:        config,
			}
			modelConfigs = append(modelConfigs, modelConfig)
		}
	}
	return modelConfigs, nil
}
