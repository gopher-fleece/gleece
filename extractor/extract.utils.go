package extractor

import (
	"go/ast"
	"strings"
)

func FindAndExtract(input []string, search string) string {
	matches := FindAndExtractOccurrences(input, search, 1)
	if len(matches) > 0 {
		return matches[0]
	}
	return ""
}

func FindAndExtractOccurrences(input []string, search string, maxMatchCount uint) []string {
	matches := []string{}

	for _, rawStr := range input {
		str := strings.TrimPrefix(strings.TrimPrefix(rawStr, "// "), "//")
		// Check if the string starts with the search term
		if strings.HasPrefix(strings.TrimSpace(str), search+" ") {
			// Remove the search term from the string and split the rest by spaces
			restText := strings.TrimPrefix(str, search)
			matches = append(matches, strings.TrimSpace(restText))
			if maxMatchCount > 0 && len(matches)+1 >= int(maxMatchCount) {
				break
			}
		}
	}
	return matches
}

// MapDocListToStrings converts a list of comment nodes (ast.Comment) to a string array
func MapDocListToStrings(docList []*ast.Comment) []string {
	var result []string
	for _, comment := range docList {
		result = append(result, comment.Text)
	}
	return result
}
