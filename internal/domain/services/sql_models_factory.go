package services

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"os"
	"strings"

	internalmodels "github.com/go-teal/teal/internal/domain/internal_models"
	"github.com/go-teal/teal/internal/domain/utils"
	"github.com/go-teal/teal/pkg/configs"
)

const MODEL_DIR = "assets/models"

func InitSQLModelConfigs(config *configs.Config, profiles *configs.ProjectProfile) ([]*internalmodels.ModelConfig, error) {
	modelsProjetDir := config.ProjectPath + "/" + MODEL_DIR
	var modelsConfigs []*internalmodels.ModelConfig

	for _, stage := range profiles.Models.Stages {
		stageName := stage.Name
		// Read the directory contents
		modelFileNames, err := os.ReadDir(modelsProjetDir + "/" + stageName)
		if err != nil {
			fmt.Printf("Error reading directory: %v", err)
			panic(err)
		}

		// Iterate through the directory entries and print the file names
		for _, modelFileNameEntry := range modelFileNames {
			if !modelFileNameEntry.IsDir() {
				originalName := modelFileNameEntry.Name()
				nameWithoutStageName := strings.Replace(originalName, ".sql", "", -1)
				fmt.Printf("Building: %s.%s\n", stageName, modelFileNameEntry.Name())
				goModelName, refName := utils.CreateModelName(stageName, modelFileNameEntry.Name())
				modelProfile := profiles.GetModelProfile(stageName, nameWithoutStageName)

				modelFileByte, err := os.ReadFile(modelsProjetDir + "/" + stageName + "/" + originalName)
				if err != nil {
					panic(err)
				}
				modelFileFinalTemplate, uniqueRefs, err := prepareModelTemplate(modelFileByte, refName, modelsProjetDir, profiles)
				if err != nil {
					fmt.Printf("can not parse model profle %s\n", string(modelFileByte))
					panic(err)
				}

				var sqlByteBuffer bytes.Buffer
				err = modelFileFinalTemplate.Execute(io.Writer(&sqlByteBuffer), nil)
				if err != nil {
					return nil, err
				}

				data := &internalmodels.ModelConfig{
					ModelName:     refName,
					GoName:        goModelName,
					Stage:         stage.Name,
					NameUpperCase: fmt.Sprintf("%s_%s", strings.ToUpper(stageName), strings.ToUpper(strings.ReplaceAll(originalName, ".sql", ""))),
					SqlByteBuffer: sqlByteBuffer,
					Config:        config,
					Profile:       profiles,
					Upstreams:     *uniqueRefs,
					ModelProfile:  modelProfile,
					ModelType:     internalmodels.DATABASE,
				}
				modelsConfigs = append(modelsConfigs, data)
			}
		}
	}

	return modelsConfigs, nil
}
