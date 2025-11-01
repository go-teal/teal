package services

import (
	"fmt"
	"regexp"
	"strings"

	pongo2 "github.com/flosch/pongo2/v6"
	"github.com/go-teal/teal/internal/domain/utils"
	"github.com/go-teal/teal/pkg/configs"
)

func GetStaticFunctions(
	refName string,
	modelsProjetDir string,
	dynamicStubs map[string]string,
	profiles *configs.ProjectProfile,
) (pongo2.Context, *utils.UpstreamDependencies) {
	var uniqueRefs *utils.UpstreamDependencies = &utils.UpstreamDependencies{}
	profilesMap := profiles.ToMap()

	// DYNAMIC_STAB handles all pongo2 template syntax
	dynamicStabFunc := func(stubHash string) string {
		originalContent := dynamicStubs[stubHash]

		// Check if it contains Ref(...) - extract model name and track dependency
		if strings.Contains(originalContent, "Ref(") {
			// Extract the argument: Ref("stage.model") or Ref('stage.model')
			refPattern := regexp.MustCompile(`Ref\s*\(\s*["']([^"']+)["']\s*\)`)
			matches := refPattern.FindStringSubmatch(originalContent)
			if len(matches) > 1 {
				ref := matches[1] // e.g., "staging.model_name"

				// Validate model exists
				isExists, err := utils.CheckModelExists(modelsProjetDir, ref, "sql")
				_, isRawProfileExist := profilesMap[ref]
				isExists = isExists || isRawProfileExist
				if !isExists {
					fmt.Println(err)
					panic(fmt.Sprintf("Model %s not found", ref))
				}

				// Track dependency (avoid duplicates)
				alreadyTracked := false
				for _, r := range *uniqueRefs {
					if r == ref {
						alreadyTracked = true
						break
					}
				}
				if !alreadyTracked {
					*uniqueRefs = append(*uniqueRefs, ref)
				}

				// Check if we need temp name for dataframed models
				currentProfile, okCurrentProfile := profilesMap[refName]
				refProfile, okRefProfile := profilesMap[ref]
				if okCurrentProfile && okRefProfile {
					if currentProfile.PersistInputs && refProfile.IsDataFramed {
						return profilesMap[ref].GetTempName()
					}
				}

				// Return the model name for SQL
				return ref
			}
		}

		// Check if it's this() - return current model name
		if strings.Contains(originalContent, "this()") {
			return refName
		}

		// For everything else (TaskID, ENV, control structures, etc.),
		// return the original template syntax for runtime evaluation
		return originalContent
	}

	return pongo2.Context{
		"DYNAMIC_STAB": dynamicStabFunc,
	}, uniqueRefs
}
