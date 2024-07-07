package services

import (
	"fmt"
	"strings"
	templ "text/template"

	"github.com/go-teal/teal/internal/domain/utils"
	"github.com/go-teal/teal/pkg/configs"
)

func GetStaticFunctions(
	refName string,
	modelsProjetDir string,
	stubs map[string]string,
	profiles *configs.ProjectProfile,
) (templ.FuncMap, *utils.UpstreamDependencies) {
	var uniqueRefs *utils.UpstreamDependencies = &utils.UpstreamDependencies{}
	profilesMap := profiles.ToMap()
	return templ.FuncMap{
		"Ref": func(ref string) string {
			isExists, err := utils.CheckModelExists(modelsProjetDir, ref, "sql")
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
		},
		"Source": func(ref string) string {
			for _, r := range *uniqueRefs {
				if r == ref {
					return ref
				}
			}
			*uniqueRefs = append(*uniqueRefs, ref)
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
	}, uniqueRefs
}
