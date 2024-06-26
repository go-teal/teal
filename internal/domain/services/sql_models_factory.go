package services

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"
	templ "text/template"

	internalmodels "github.com/go-teal/teal/internal/domain/internal_models"
	"github.com/go-teal/teal/internal/domain/utils"
	"github.com/go-teal/teal/pkg/configs"

	"gopkg.in/yaml.v2"
)

const MODEL_DIR = "assets/models"

func InitSQLModelConfigs(config *configs.Config, profiles *configs.Profile) ([]*internalmodels.ModelConfig, error) {
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

				fmt.Printf("Building: %s.%s\n", stageName, modelFileNameEntry.Name())
				goModelName, refName := utils.CreateModelName(stageName, modelFileNameEntry.Name())

				var uniqueRefs utils.UpstreamDependencies

				sqlTemplateByte, err := os.ReadFile(modelsProjetDir + "/" + stageName + "/" + originalName)
				if err != nil {
					panic(err)
				}
				inlineFunctions := utils.ExtractStringsBetweenBraces(sqlTemplateByte)
				stringTemplate := string(sqlTemplateByte)
				stubs := make(map[string]string, len(inlineFunctions))

				funcMap := templ.FuncMap{
					"Ref": func(ref string) string {
						isExists, err := utils.CheckModelExists(modelsProjetDir, ref, "sql")
						if !isExists {
							fmt.Println(err)
							panic(fmt.Sprintf("Model %s not found", ref))
						}
						for _, r := range uniqueRefs {
							if r == ref {
								return ref
							}
						}
						uniqueRefs = append(uniqueRefs, ref)
						return ref
					},
					"Source": func(ref string) string {
						for _, r := range uniqueRefs {
							if r == ref {
								return ref
							}
						}
						uniqueRefs = append(uniqueRefs, ref)
						return ref
					},
					"ENV": func(arg1 string, arg2 string) string {
						return fmt.Sprintf("{{ ENV \"%s\" \"%s\" }}", arg1, arg2)
					},
					"this": func() string {
						return refName
					},

					"STAB": func(stubHash string) string {
						result := strings.ReplaceAll(stubs[stubHash], "{{{", "")
						result = strings.ReplaceAll(result, "}}}", "")
						return "{{" + result + "}}"
					},
				}

				fmt.Println(strings.Join(inlineFunctions, ","))

				for _, inlineFunctionCall := range inlineFunctions {
					stubHash := utils.GetMD5Hash(inlineFunctionCall)
					stubs[stubHash] = inlineFunctionCall
					stringTemplate = strings.ReplaceAll(stringTemplate, inlineFunctionCall, "{{ STAB \""+stubHash+"\" }}")
				}
				sqlCreateTempl, err := template.New(originalName).Funcs(funcMap).Parse(stringTemplate)
				if err != nil {
					return nil, err
				}

				var configByteBuffer bytes.Buffer
				var modelProfile configs.ModelProfile
				err = sqlCreateTempl.ExecuteTemplate(&configByteBuffer, "profile.yaml", nil)
				if err == nil {
					err = yaml.Unmarshal(configByteBuffer.Bytes(), &modelProfile)
					if err != nil {
						return nil, err
					}
				}

				var sqlByteBuffer bytes.Buffer
				err = sqlCreateTempl.Execute(io.Writer(&sqlByteBuffer), nil)
				if err != nil {
					return nil, err
				}

				if modelProfile.Connection == "" {
					modelProfile.Connection = profiles.GetModelProfile(stageName, refName).Connection
				}

				if modelProfile.Materialization == "" {
					modelProfile.Materialization = profiles.GetModelProfile(stageName, refName).Materialization
				}

				data := &internalmodels.ModelConfig{
					ModelName:     refName,
					GoName:        goModelName,
					Stage:         stage.Name,
					NameUpperCase: fmt.Sprintf("%s_%s", strings.ToUpper(stageName), strings.ToUpper(strings.ReplaceAll(originalName, ".sql", ""))),
					SqlByteBuffer: sqlByteBuffer,
					Config:        config,
					Profile:       profiles,
					Upstreams:     uniqueRefs,
					ModelProfile:  &modelProfile,
					ModelType:     internalmodels.DATABASE,
				}
				modelsConfigs = append(modelsConfigs, data)
			}
		}
	}

	return modelsConfigs, nil
}
