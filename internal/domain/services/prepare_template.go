package services

import (
	"strings"
	"text/template"

	"github.com/go-teal/teal/internal/domain/utils"
	"github.com/go-teal/teal/pkg/configs"
)

func prepareModelTemplate(modelFileByte []byte, refName string, modelsProjetDir string, profiles *configs.ProjectProfile) (*template.Template, *utils.UpstreamDependencies, error) {
	modelFileString := string(modelFileByte)
	inlineFunctions := utils.ExtractStringsBetweenBraces(modelFileByte)
	stubs := make(map[string]string, len(inlineFunctions))
	funcsMap, uniqueRefs := GetStaticFunctions(refName, modelsProjetDir, stubs, profiles)
	for _, inlineFunctionCall := range inlineFunctions {
		stubHash := utils.GetMD5Hash(inlineFunctionCall)
		stubs[stubHash] = inlineFunctionCall
		modelFileString = strings.ReplaceAll(modelFileString, inlineFunctionCall, "{{ STAB \""+stubHash+"\" }}")
	}
	modelFileFinalTemplate, err := template.New("sql model").Funcs(funcsMap).Parse(modelFileString)
	return modelFileFinalTemplate, uniqueRefs, err
}
