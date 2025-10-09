package services

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/go-teal/teal/pkg/configs"
	"gopkg.in/yaml.v2"
)

func CombineProfiles(config *configs.Config, projectProfile *configs.ProjectProfile) {
	fmt.Println("reading model profiles...")
	modelsProjectDir := config.ProjectPath + "/" + MODEL_DIR

	for _, stage := range projectProfile.Models.Stages {
		fmt.Printf("Stage: %s\n", stage.Name)
		
		// Step 1: Get all profiles from projectProfile
		profilesFromYAML := make(map[string]*configs.ModelProfile)
		for _, modelProfile := range stage.Models {
			key := stage.Name + "." + modelProfile.Name
			fmt.Printf("Model from profile.yaml: %s\n", key)
			modelProfile.Stage = stage.Name
			profilesFromYAML[key] = modelProfile
		}

		// Step 2: Find all asset files from MODEL_DIR
		profilesFromSQL := make(map[string]*configs.ModelProfile)
		modelFileNames, err := os.ReadDir(modelsProjectDir + "/" + stage.Name)
		if err != nil {
			fmt.Printf("Error reading directory: %v\n", err)
			panic(err)
		}

		// Step 3: For each file, execute sub template profile.yaml
		for _, modelFileNameEntry := range modelFileNames {
			if modelFileNameEntry.IsDir() {
				continue
			}

			modelFileName := modelFileNameEntry.Name()
			if !strings.HasSuffix(modelFileName, ".sql") {
				continue
			}

			fmt.Printf("Processing file: %s\n", modelFileName)
			modelName := strings.TrimSuffix(modelFileName, ".sql")
			refName := stage.Name + "." + modelName

			// Read and parse the SQL file
			modelFileByte, err := os.ReadFile(modelsProjectDir + "/" + stage.Name + "/" + modelFileName)
			if err != nil {
				panic(err)
			}

			modelFileFinalTemplate, _, err := prepareModelTemplate(modelFileByte, refName, modelsProjectDir, projectProfile)
			if err != nil {
				fmt.Printf("Cannot parse model profile for %s: %v\n", modelFileName, err)
				continue
			}

			// Execute the profile.yaml template if it exists
			var inlineProfileByteBuffer bytes.Buffer
			err = modelFileFinalTemplate.ExecuteTemplate(&inlineProfileByteBuffer, "profile.yaml", nil)
			if err == nil && inlineProfileByteBuffer.Len() > 0 {
				var sqlProfile configs.ModelProfile
				err = yaml.Unmarshal(inlineProfileByteBuffer.Bytes(), &sqlProfile)
				if err != nil {
					fmt.Printf("Cannot unmarshal profile from %s: %v\n", modelFileName, err)
					continue
				}
				
				sqlProfile.Name = modelName
				sqlProfile.Stage = stage.Name
				profilesFromSQL[refName] = &sqlProfile
				fmt.Printf("Found inline profile in: %s\n", refName)
			}
		}

		// Step 4: Merge the two sets of profiles with priority from projectProfile
		mergedProfiles := make(map[string]*configs.ModelProfile)
		
		// First, add all profiles from YAML (they have priority)
		for key, yamlProfile := range profilesFromYAML {
			mergedProfiles[key] = yamlProfile
		}

		// Then, merge or add profiles from SQL files
		for key, sqlProfile := range profilesFromSQL {
			if yamlProfile, exists := mergedProfiles[key]; exists {
				// Merge profiles - YAML has priority
				mergedProfiles[key] = mergeModelProfiles(yamlProfile, sqlProfile)
			} else {
				// No YAML profile exists, use SQL profile and apply defaults
				applyDefaultsToProfile(sqlProfile)
				mergedProfiles[key] = sqlProfile
			}
		}

		// Apply defaults to all profiles and prepare final list
		stage.Models = make([]*configs.ModelProfile, 0, len(mergedProfiles))
		for _, profile := range mergedProfiles {
			applyDefaultsToProfile(profile)
			// Set test connections and stages
			for _, testProfile := range profile.Tests {
				if testProfile.Connection == "" {
					testProfile.Connection = profile.Connection
				}
				testProfile.Stage = profile.Stage
			}
			stage.Models = append(stage.Models, profile)
		}

		fmt.Printf("Stage %s: merged %d profiles\n", stage.Name, len(stage.Models))
	}
}

// mergeModelProfiles merges two profiles with priority given to the primary profile (from YAML)
// Returns a new profile with merged values
func mergeModelProfiles(primary, secondary *configs.ModelProfile) *configs.ModelProfile {
	merged := &configs.ModelProfile{
		Name:  primary.Name,
		Stage: primary.Stage,
	}

	// Merge Description - primary has priority if not empty
	if primary.Description != "" {
		merged.Description = primary.Description
	} else {
		merged.Description = secondary.Description
	}

	// Merge Connection - primary has priority if not empty
	if primary.Connection != "" {
		merged.Connection = primary.Connection
	} else {
		merged.Connection = secondary.Connection
	}

	// Merge Materialization - primary has priority if not empty
	if primary.Materialization != "" {
		merged.Materialization = primary.Materialization
	} else {
		merged.Materialization = secondary.Materialization
	}

	// Merge PrimaryKeyFields - primary has priority if not empty
	if len(primary.PrimaryKeyFields) > 0 {
		merged.PrimaryKeyFields = primary.PrimaryKeyFields
	} else {
		merged.PrimaryKeyFields = secondary.PrimaryKeyFields
	}

	// Merge Indexes - primary has priority if not empty
	if len(primary.Indexes) > 0 {
		merged.Indexes = primary.Indexes
	} else {
		merged.Indexes = secondary.Indexes
	}

	// Merge Tests - primary has priority if not empty
	if len(primary.Tests) > 0 {
		merged.Tests = primary.Tests
	} else {
		merged.Tests = secondary.Tests
	}

	// Merge RawUpstreams - primary has priority if not empty
	if len(primary.RawUpstreams) > 0 {
		merged.RawUpstreams = primary.RawUpstreams
	} else {
		merged.RawUpstreams = secondary.RawUpstreams
	}

	// Merge boolean fields - true takes priority
	merged.IsDataFramed = primary.IsDataFramed || secondary.IsDataFramed
	merged.PersistInputs = primary.PersistInputs || secondary.PersistInputs

	return merged
}

// applyDefaultsToProfile applies default values to empty fields
func applyDefaultsToProfile(profile *configs.ModelProfile) {
	if profile.Connection == "" {
		profile.Connection = "default"
	}
	if profile.Materialization == "" {
		profile.Materialization = configs.MAT_TABLE
	}
}
