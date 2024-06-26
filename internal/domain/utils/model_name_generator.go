package utils

import (
	"fmt"
	"strings"
	"unicode"
)

type UpstreamDependencies []string

func (ud UpstreamDependencies) ToModelsNameArray() []string {
	var upsreamModels = make([]string, len(ud))
	for i, d := range ud {
		spleated := strings.Split(d, ".")
		upsreamModels[i], _ = CreateModelName(spleated[0], spleated[1])
	}
	return upsreamModels
}

func capitalizeFirstLetter(s string) string {
	if len(s) == 0 {
		return s
	}

	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

func CreateModelName(stageName string, fileName string) (string, string) {
	name := strings.Replace(fileName, ".sql", "", -1)
	fullName := fmt.Sprintf("%s_%s", stageName, name)

	return ToCamelCase(fullName), fmt.Sprintf("%s.%s", stageName, name)
}

func ToCamelCase(input string) string {
	input = strings.ReplaceAll(input, ".", "_")
	splittedParts := strings.Split(input, "_")
	for i, part := range splittedParts {
		if i == 0 {
			splittedParts[i] = strings.ToLower(part)
			continue
		}
		splittedParts[i] = capitalizeFirstLetter(strings.ToLower(part))
	}
	return strings.Join(splittedParts, "")
}
