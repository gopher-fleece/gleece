package swagen

import (
	"strconv"
	"strings"

	"github.com/haimkastner/gleece/external"
)

func HttpStatusCodeToString(httpStatusCode external.HttpStatusCode) string {
	statusCode := uint64(httpStatusCode)
	return strconv.FormatUint(statusCode, 10)
}

// Helper function to parse numeric validation values
func ParseNumber(value string) *float64 {
	if v, err := strconv.ParseFloat(value, 64); err == nil {
		return &v
	}
	return nil
}

// Helper function to parse integer validation values
func ParseInteger(value string) *uint64 {
	if v, err := strconv.ParseUint(value, 10, 64); err == nil {
		return &v
	}
	return nil
}

// Helper function to parse boolean validation values
func ParseBool(value string) *bool {
	if v, err := strconv.ParseBool(value); err == nil {
		return &v
	}
	return nil
}

func IsFieldRequired(validationString string) bool {
	validationRules := strings.Split(validationString, ",")
	for _, rule := range validationRules {
		if rule == "required" {
			return true
		}
	}
	return false
}

// Helper function to determine the item type of an array
func GetArrayItemType(fieldType string) string {
	// Implement logic to extract the item type from the array type
	// For example, if fieldType is "[]string", return "string"
	return strings.TrimPrefix(fieldType, "[]")
}
