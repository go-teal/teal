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

	// Find ALL pongo2 template syntax: {{ ... }} and {% ... %}
	// We'll replace them with {{ DYNAMIC_STAB("hash") }} so we can control what gets evaluated
	dynamicStubs := make(map[string]string)

	// Pattern to match {{ ... }} (variable/expression blocks)
	variablePattern := regexp.MustCompile(`\{\{[^}]*\}\}`)
	variableBlocks := variablePattern.FindAllString(modelFileString, -1)
	for _, block := range variableBlocks {
		stubHash := utils.GetMD5Hash(block)
		dynamicStubs[stubHash] = block
		modelFileString = strings.ReplaceAll(modelFileString, block, "{{ DYNAMIC_STAB(\""+stubHash+"\") }}")
	}

	// Pattern to match {% ... %} (control structure tags ONLY, not the content between)
	// This matches individual tags like {% if ... %}, {% endif %}, {% for ... %}, {% endfor %}, etc.
	controlPattern := regexp.MustCompile(`\{%[^%]*%\}`)
	controlBlocks := controlPattern.FindAllString(modelFileString, -1)
	for _, block := range controlBlocks {
		stubHash := utils.GetMD5Hash(block)
		dynamicStubs[stubHash] = block
		modelFileString = strings.ReplaceAll(modelFileString, block, "{{ DYNAMIC_STAB(\""+stubHash+"\") }}")
	}

	// Create function context with DYNAMIC_STAB function
	funcsContext, uniqueRefs := GetStaticFunctions(refName, modelsProjetDir, dynamicStubs, profiles)

	modelFileFinalTemplate, err := pongo2.FromString(modelFileString)
	if err != nil {
		return nil, uniqueRefs, err
	}

	output, err := modelFileFinalTemplate.Execute(funcsContext)
	if err != nil {
		return nil, uniqueRefs, err
	}

	// Return the processed string with runtime template syntax intact
	return &PreparedTemplate{
		processedSQL: output,
		uniqueRefs:   uniqueRefs,
		profileYAML:  profileYAML,
	}, uniqueRefs, nil
}
