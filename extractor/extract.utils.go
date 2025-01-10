package extractor

import (
	"fmt"
	"go/ast"
	"regexp"
	"strings"
)

// SearchForParamTerm searches for a term in a list of strings that comes immediately after "// @" with no space, and with space after it.
func SearchForParamTerm(lines []string, searchTerm string) string {
	// Construct a regular expression to match "// @<searchTerm> " (no space between @ and term, space after term)
	searchPattern := fmt.Sprintf(`// @\S+\s+%s\s*`, regexp.QuoteMeta(searchTerm))

	// Iterate through lines to find a match
	for _, line := range lines {
		// Check if line matches the search pattern
		match, _ := regexp.MatchString(searchPattern, line)
		if match {
			return line
		}
	}

	return "" // Return empty string if no match found
}

// GetTextBeforeParenthesis extracts the text before the character '(' in a given string.
func GetTextBeforeParenthesis(input string, splitter string) string {
	// Find the index of the first occurrence of '('
	index := strings.Index(input, splitter)
	if index == -1 {
		// If '(' is not found, return the original string
		return input
	}
	return input[:index]
}

// GetTextAfterParenthesis extracts the text before the character '(' in a given string.
// If '(' is not present, it returns the original string.
func GetTextAfterParenthesis(input string, splitter string) string {
	// Find the index of the first occurrence of '('
	index := strings.Index(input, splitter)
	if index == -1 {
		return ""
	}
	return input[index+len(splitter):]
}

// ExtractParamTerm extracts the term immediately after "// @" and trims surrounding spaces.
func ExtractParamTerm(line string) string {
	// Regular expression to match "// @<term>"
	re := regexp.MustCompile(`// @(\S+)`)

	// Find the match
	matches := re.FindStringSubmatch(line)
	if len(matches) > 1 {
		return GetTextBeforeParenthesis(strings.TrimSpace(matches[1]), "(") // Return the term without leading/trailing spaces
	}
	return "" // Return empty if no match is found
}

// ExtractParenthesesContent extracts the text inside parentheses after the word immediately following "@"
// If no parentheses exist, it returns an empty string.
func ExtractParenthesesContent(line string) string {
	// Regular expression to match "// @<word>(<content>)"
	re := regexp.MustCompile(`// @\S+\(([^)]+)\)`)

	// Find the first match
	match := re.FindStringSubmatch(line)
	if len(match) > 1 {
		return match[1] // Return the content inside the parentheses
	}

	return "" // Return empty string if no match is found
}

func FindAndExtract(input []string, search string) string {
	for _, rawStr := range input {
		str := strings.TrimPrefix(strings.TrimPrefix(rawStr, "// "), "//")
		// Check if the string starts with the search term
		if strings.HasPrefix(strings.TrimSpace(str), search+" ") {
			// Remove the search term from the string and split the rest by spaces
			restText := strings.TrimPrefix(str, search)
			return strings.TrimSpace(restText)
		}
	}
	return ""
}

func FindAndExtractArray(input []string, search string) []string {
	extracted := FindAndExtract(input, search)
	parts := strings.Fields(extracted)
	return parts
}

// MapDocListToStrings converts a list of comment nodes (ast.Comment) to a string array
func MapDocListToStrings(docList []*ast.Comment) []string {
	var result []string
	for _, comment := range docList {
		result = append(result, comment.Text)
	}
	return result
}

func BuildRestMetadata(comments []string) RestMetadata {
	restMetadata := RestMetadata{}
	route := FindAndExtract(comments, "@Route")
	restMetadata.Path = route
	return restMetadata
}

// FilterDecls filters a slice of ast.Decl using a custom type-checking function.
func FilterDecls(decls []ast.Decl, check func(ast.Decl) bool) []ast.Decl {
	var results []ast.Decl
	for _, decl := range decls {
		if check(decl) {
			results = append(results, decl)
		}
	}
	return results
}

// EmbedsBaseStruct checks if a struct embeds the specified base struct.
func EmbedsBaseStruct(structType *ast.StructType, baseStruct string) bool {
	for _, field := range structType.Fields.List {
		// Check for embedded fields
		if len(field.Names) == 0 { // Embedded fields have no names
			if ident, ok := field.Type.(*ast.Ident); ok {
				if ident.Name == baseStruct {
					return true
				}
			}
		}
	}
	return false
}
