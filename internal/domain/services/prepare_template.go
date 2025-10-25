package services

import (
	"regexp"
	"strings"

	pongo2 "github.com/flosch/pongo2/v6"
	"github.com/go-teal/teal/internal/domain/utils"
	"github.com/go-teal/teal/pkg/configs"
)

type PreparedTemplate struct {
	processedSQL  string
	uniqueRefs    *utils.UpstreamDependencies
	profileYAML   string
}

func (pt *PreparedTemplate) Execute(data interface{}) (string, error) {
	// Template is already executed during prepareModelTemplate, just return the processed string
	return pt.processedSQL, nil
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

		// Extract and normalize indentation in the block content
		// Don't trim the whole string first - work with lines directly
		blockContent := matches[1]
		lines := strings.Split(blockContent, "\n")

		// Remove empty lines at start and end
		for len(lines) > 0 && strings.TrimSpace(lines[0]) == "" {
			lines = lines[1:]
		}
		for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
			lines = lines[:len(lines)-1]
		}

		if len(lines) == 0 {
			return strings.TrimSpace(cleanContent), ""
		}

		// Find minimum indentation from non-empty lines
		minIndent := -1
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" {
				continue
			}
			indent := len(line) - len(strings.TrimLeft(line, " \t"))
			if minIndent == -1 || indent < minIndent {
				minIndent = indent
			}
		}

		// Remove common indentation from all lines
		if minIndent > 0 {
			for i, line := range lines {
				if len(line) >= minIndent && strings.TrimSpace(line) != "" {
					lines[i] = line[minIndent:]
				} else if strings.TrimSpace(line) == "" {
					lines[i] = ""
				}
			}
		}

		// Return cleaned content and the normalized block content
		return strings.TrimSpace(cleanContent), strings.Join(lines, "\n")
	}

	return content, ""
}

func prepareModelTemplate(modelFileByte []byte, refName string, modelsProjetDir string, profiles *configs.ProjectProfile) (*PreparedTemplate, *utils.UpstreamDependencies, error) {
	modelFileString := string(modelFileByte)

	// Extract and remove profile.yaml define block before processing
	modelFileString, profileYAML := extractDefineBlock(modelFileString, "profile.yaml")

	// Preserve runtime template blocks by stubbing them out before pongo2 processing
	// Replace {% if ... %} and {% endif %} with placeholders ((?s) enables DOTALL mode)
	runtimeBlockPattern := regexp.MustCompile(`(?s)(\{%\s*if\s+.*?%\}.*?\{%\s*endif\s*%\})`)
	runtimeBlocks := runtimeBlockPattern.FindAllString(modelFileString, -1)
	runtimeBlockStubs := make(map[string]string)
	for _, block := range runtimeBlocks {
		stubHash := utils.GetMD5Hash(block)
		runtimeBlockStubs[stubHash] = block
		modelFileString = strings.ReplaceAll(modelFileString, block, "___RUNTIME_BLOCK_"+stubHash+"___")
	}

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

	output, err := modelFileFinalTemplate.Execute(funcsContext)
	if err != nil {
		return nil, uniqueRefs, err
	}

	// Restore runtime blocks after pongo2 processing
	for stubHash, block := range runtimeBlockStubs {
		output = strings.ReplaceAll(output, "___RUNTIME_BLOCK_"+stubHash+"___", block)
	}

	// Just return the processed string with runtime blocks intact
	return &PreparedTemplate{
		processedSQL: output,
		uniqueRefs:   uniqueRefs,
		profileYAML:  profileYAML,
	}, uniqueRefs, nil
}
