package services

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-teal/teal/internal/domain/utils"
	"github.com/go-teal/teal/pkg/configs"
	"gopkg.in/yaml.v2"
)

func CombineProfiles(config *configs.Config, projectProfile *configs.ProjectProfile) {
	fmt.Println("reading model profiles...")
	modelsProjetDir := config.ProjectPath + "/" + MODEL_DIR

	for _, stage := range projectProfile.Models.Stages {
		fmt.Printf("Stage: %s\n", stage.Name)
		modelProfilesMap := make(map[string]*configs.ModelProfile)
		for _, modelProfile := range stage.Models {
			fmt.Printf("Model: %s.%s\n", stage.Name, modelProfile.Name)
			modelProfilesMap[stage.Name+"."+modelProfile.Name] = modelProfile
		}

		modelFileNames, err := os.ReadDir(modelsProjetDir + "/" + stage.Name)
		if err != nil {
			fmt.Printf("Error reading directory: %v", err)
			panic(err)
		}
		for _, modelFileNameEntry := range modelFileNames {
			if !modelFileNameEntry.IsDir() {
				modelFileName := modelFileNameEntry.Name()
				fmt.Printf("File: %s\n", modelFileName)
				_, refName := utils.CreateModelName(stage.Name, modelFileNameEntry.Name())
				modelFileByte, err := os.ReadFile(modelsProjetDir + "/" + stage.Name + "/" + modelFileName)
				if err != nil {
					panic(err)
				}
				preparedTemplate, _, err := prepareModelTemplate(modelFileByte, refName, modelsProjetDir, projectProfile)
				if err != nil {
					fmt.Printf("can not parse model profile %s\n", string(modelFileByte))
					continue
				}
				var newModelProfile configs.ModelProfile

				// Extract profile.yaml from define block if present
				profileYAML := preparedTemplate.GetProfileYAML()
				if profileYAML != "" {
					inlineProfileBytes := []byte(profileYAML)
					fmt.Printf("Overriding profile: %s\n", refName)
					err = yaml.Unmarshal(inlineProfileBytes, &newModelProfile)
					if err != nil {
						fmt.Printf("can not unmarshal parse model profile")
						continue
					}
					newModelProfile.Name = strings.Replace(modelFileName, ".sql", "", -1)
					profile, ok := modelProfilesMap[refName]
					// TODO: simplify
					if !ok {
						// If profile is not defined in the profile.yaml file, we read it from the model file
						modelProfilesMap[refName] = &newModelProfile
					} else {
						// If profile defined in the profile.yaml file, we try to merge profiles
						if newModelProfile.Connection != "" {
							profile.Connection = newModelProfile.Connection
						}
						if newModelProfile.Materialization != "" {
							profile.Materialization = newModelProfile.Materialization
						}

						if len(newModelProfile.PrimaryKeyFields) > 0 {
							profile.PrimaryKeyFields = newModelProfile.PrimaryKeyFields
						}
						if len(newModelProfile.Indexes) > 0 {
							profile.Indexes = newModelProfile.Indexes
						}

						if len(newModelProfile.Tests) > 0 {
							profile.Tests = newModelProfile.Tests
							for _, testProfile := range profile.Tests {
								if testProfile.Connection == "" {
									testProfile.Connection = profile.Connection
								}
								testProfile.Stage = profile.Stage
							}
						}
						if profile.Connection == "" {
							profile.Connection = "default"
						}
						profile.IsDataFramed = newModelProfile.IsDataFramed || profile.IsDataFramed
						profile.PersistInputs = newModelProfile.PersistInputs || profile.PersistInputs
						modelProfilesMap[refName] = profile
					}
				}
			}
		}

		stage.Models = make([]*configs.ModelProfile, len(modelProfilesMap))
		var idx int
		for _, v := range modelProfilesMap {
			stage.Models[idx] = v
			idx++
		}
	}
}
