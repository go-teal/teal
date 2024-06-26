package utils

import (
	"bytes"
)

func ExtractStringsBetweenBraces(input []byte) []string {
	var results []string
	start := 0
	for {
		// Find the start of the first double curly brace
		startIdx := bytes.Index(input[start:], []byte("{{{"))
		if startIdx == -1 {
			break
		}
		start = start + startIdx

		// Find the end of the substring by locating the next double curly brace
		endIdx := bytes.Index(input[start:], []byte("}}}"))
		if endIdx == -1 {
			break
		}

		// Extract the substring
		substring := input[start : start+endIdx+3]
		results = append(results, string(substring))

		// Move the start index to after the found double curly brace
		start = start + endIdx + 3
	}
	return results
}
