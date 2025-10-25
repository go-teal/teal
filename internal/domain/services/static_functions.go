package services

import (
	"fmt"
	"strings"

	pongo2 "github.com/flosch/pongo2/v6"
	"github.com/go-teal/teal/internal/domain/utils"
	"github.com/go-teal/teal/pkg/configs"
)

func GetStaticFunctions(
	refName string,
	modelsProjetDir string,
	stubs map[string]string,
	profiles *configs.ProjectProfile,
) (pongo2.Context, *utils.UpstreamDependencies) {
	var uniqueRefs *utils.UpstreamDependencies = &utils.UpstreamDependencies{}
	profilesMap := profiles.ToMap()

	refFunc := func(ref string) string {
		isExists, err := utils.CheckModelExists(modelsProjetDir, ref, "sql")
		_, isRawProfileExist := profilesMap[ref]
		isExists = isExists || isRawProfileExist
		if !isExists {
			fmt.Println(err)
			panic(fmt.Sprintf("Model %s not found", ref))
		}
		for _, r := range *uniqueRefs {
			if r == ref {
				return ref
			}
		}
		*uniqueRefs = append(*uniqueRefs, ref)
		currentProfile, okCurrentProfile := profilesMap[refName]
		refProfile, okRefProfile := profilesMap[ref]
		if okCurrentProfile && okRefProfile {
			if currentProfile.PersistInputs && refProfile.IsDataFramed {
				return profilesMap[ref].GetTempName()
			}
		}

		return ref
	}

	sourceFunc := func(ref string) string {
		for _, r := range *uniqueRefs {
			if r == ref {
				return ref
			}
		}
		*uniqueRefs = append(*uniqueRefs, ref)
		return ref
	}

	envFunc := func(arg1 string, arg2 string) string {
		return fmt.Sprintf("{{ ENV(\"%s\", \"%s\") }}", arg1, arg2)
	}

	thisFunc := func() string {
		return refName
	}

	stabFunc := func(stubHash string) string {
		result := strings.ReplaceAll(stubs[stubHash], "{{{", "")
		result = strings.ReplaceAll(result, "}}}", "")
		return "{{" + result + "}}"
	}

	return pongo2.Context{
		"Ref":    refFunc,
		"Source": sourceFunc,
		"ENV":    envFunc,
		"this":   thisFunc,
		"STAB":   stabFunc,
	}, uniqueRefs
}
