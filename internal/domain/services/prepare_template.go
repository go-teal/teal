package services

import (
	"regexp"
	"strings"

	pongo2 "github.com/flosch/pongo2/v6"
	"github.com/go-teal/teal/internal/domain/utils"
	"github.com/go-teal/teal/pkg/configs"
)

type PreparedTemplate struct {
	template      *pongo2.Template
	context       pongo2.Context
	uniqueRefs    *utils.UpstreamDependencies
	profileYAML   string
}

func (pt *PreparedTemplate) Execute(data interface{}) (string, error) {
	return pt.template.Execute(pt.context)
}

func (pt *PreparedTemplate) GetProfileYAML() string {
	return pt.profileYAML
}

// extractDefineBlock extracts and removes {{ define "name" }}...{{ end }} blocks
func extractDefineBlock(content string, blockName string) (string, string) {
	// Match {{ define "blockName" }} ... {{ end }} with DOTALL flag ((?s) makes . match newlines)
	pattern := `(?s)\{\{\s*define\s+"` + regexp.QuoteMeta(blockName) + `"\s*\}\}(.*?)\{\{\s*end\s*\}\}`
	re := regexp.MustCompile(pattern)

	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		// Remove the define block from content
		cleanContent := re.ReplaceAllString(content, "")
		// Return cleaned content and the block content
		return strings.TrimSpace(cleanContent), strings.TrimSpace(matches[1])
	}

	return content, ""
}

func prepareModelTemplate(modelFileByte []byte, refName string, modelsProjetDir string, profiles *configs.ProjectProfile) (*PreparedTemplate, *utils.UpstreamDependencies, error) {
	modelFileString := string(modelFileByte)

	// Extract and remove profile.yaml define block before processing
	modelFileString, profileYAML := extractDefineBlock(modelFileString, "profile.yaml")

	inlineFunctions := utils.ExtractStringsBetweenBraces([]byte(modelFileString))
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
		template:    modelFileFinalTemplate,
		context:     funcsContext,
		uniqueRefs:  uniqueRefs,
		profileYAML: profileYAML,
	}, uniqueRefs, nil
}
