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
	modelsProjectDir := config.ProjectPath + "/" + MODEL_DIR
	var modelsConfigs []*internalmodels.ModelConfig

	for _, stage := range profiles.Models.Stages {
		stageName := stage.Name
		// Read the directory contents
		modelFileNames, err := os.ReadDir(modelsProjectDir + "/" + stageName)
		if err != nil {
			fmt.Printf("Error reading directory: %v", err)
			panic(err)
		}

		// Iterate through the directory entries and print the file names
		for _, modelFileNameEntry := range modelFileNames {
			if !modelFileNameEntry.IsDir() {
				originalName := modelFileNameEntry.Name()
				splittedName := strings.Split(originalName, ".")
				fileExtenstion := splittedName[len(splittedName)-1]
				if strings.ToLower(fileExtenstion) != "sql" {
					continue
				}
				nameWithoutStageName := strings.Replace(originalName, ".sql", "", -1)
				fmt.Printf("Building: %s.%s\n", stageName, modelFileNameEntry.Name())
				goModelName, refName := utils.CreateModelName(stageName, modelFileNameEntry.Name())

				var modelProfile *configs.ModelProfile
				defaultModelProfile := profiles.GetModelProfile(stageName, nameWithoutStageName)
				var ok bool
				if modelProfile, ok = profiles.ToMap()[stageName+"."+nameWithoutStageName]; !ok {
					modelProfile = defaultModelProfile
				}

				if modelProfile.Connection == "" {
					modelProfile.Connection = defaultModelProfile.Connection
				}

				if modelProfile.Materialization == "" {
					modelProfile.Materialization = defaultModelProfile.Materialization
				}

				if modelProfile.Stage == "" {
					modelProfile.Stage = defaultModelProfile.Stage
				}

				modelFileByte, err := os.ReadFile(modelsProjectDir + "/" + stageName + "/" + originalName)
				if err != nil {
					panic(err)
				}
				modelFileFinalTemplate, uniqueRefs, err := prepareModelTemplate(modelFileByte, refName, modelsProjectDir, profiles)
				if err != nil {
					fmt.Printf("can not parse model profile %s\n", string(modelFileByte))
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
				if len(modelProfile.PrimaryKeyFields) > 0 {
					data.PrimaryKeyExpression = strings.Join(defaultModelProfile.PrimaryKeyFields, ", ")
				}
				modelsConfigs = append(modelsConfigs, data)
			}
		}
	}

	return modelsConfigs, nil
}
