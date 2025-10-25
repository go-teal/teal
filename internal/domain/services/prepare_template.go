package services

import (
	"strings"

	pongo2 "github.com/flosch/pongo2/v6"
	"github.com/go-teal/teal/internal/domain/utils"
	"github.com/go-teal/teal/pkg/configs"
)

type PreparedTemplate struct {
	template   *pongo2.Template
	context    pongo2.Context
	uniqueRefs *utils.UpstreamDependencies
}

func (pt *PreparedTemplate) Execute(data interface{}) (string, error) {
	return pt.template.Execute(pt.context)
}

func prepareModelTemplate(modelFileByte []byte, refName string, modelsProjetDir string, profiles *configs.ProjectProfile) (*PreparedTemplate, *utils.UpstreamDependencies, error) {
	modelFileString := string(modelFileByte)
	inlineFunctions := utils.ExtractStringsBetweenBraces(modelFileByte)
	stubs := make(map[string]string, len(inlineFunctions))
	funcsContext, uniqueRefs := GetStaticFunctions(refName, modelsProjetDir, stubs, profiles)
	for _, inlineFunctionCall := range inlineFunctions {
		stubHash := utils.GetMD5Hash(inlineFunctionCall)
		stubs[stubHash] = inlineFunctionCall
		modelFileString = strings.ReplaceAll(modelFileString, inlineFunctionCall, "{{ STAB(\""+stubHash+"\") }}")
	}
	modelFileFinalTemplate, err := pongo2.FromString(modelFileString)
	if err != nil {
		return nil, uniqueRefs, err
	}

	return &PreparedTemplate{
		template:   modelFileFinalTemplate,
		context:    funcsContext,
		uniqueRefs: uniqueRefs,
	}, uniqueRefs, nil
}
