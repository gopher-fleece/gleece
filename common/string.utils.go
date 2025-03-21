package common

import (
	"regexp"
	"strings"
)

var multiSlashRegex = *regexp.MustCompile(`/+`)

// UnwrapArrayTypeString Unwraps type strings like "[]string" to "string".
// Works iteratively ('Recursive' is name is for clarity) to unwrap nested arrays as well.
// For example, "[][][]string" will be unwrapped to "string".
func UnwrapArrayTypeString(value string) string {
	resultValue := value
	for {
		newValue := strings.TrimPrefix(resultValue, "[]")
		if len(resultValue) == len(newValue) {
			break
		}
		resultValue = newValue
	}
	return resultValue
}

func RemoveDuplicateSlash(value string) string {
	return multiSlashRegex.ReplaceAllString(value, "/")
}
