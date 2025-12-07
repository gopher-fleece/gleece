package swagtool

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strconv"

	"github.com/gopher-fleece/runtime"
)

func HttpStatusCodeToString(httpStatusCode runtime.HttpStatusCode) string {
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
func ParseInteger(value string) *int64 {
	if v, err := strconv.ParseInt(value, 10, 64); err == nil {
		return &v
	}
	return nil
}

// Helper function to parse integer validation values
func ParseUInteger(value string) *uint64 {
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

func AreJSONsIdentical(json1 []byte, json2 []byte) (bool, error) {
	var obj1, obj2 map[string]interface{}

	err := json.Unmarshal(json1, &obj1)
	if err != nil {
		return false, fmt.Errorf("invalid JSON 1: %v", err)
	}

	err = json.Unmarshal(json2, &obj2)
	if err != nil {
		return false, fmt.Errorf("invalid JSON 2: %v", err)
	}

	return reflect.DeepEqual(obj1, obj2), nil
}

func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func ForceOrderedJSON(input []byte) ([]byte, error) {
	var orderedData any
	if err := json.Unmarshal(input, &orderedData); err != nil {
		return nil, fmt.Errorf("error unmarshaling JSON for ordering: %v", err)
	}

	// Sort enum values in components.schemas
	sortEnumValues(orderedData)

	orderedJSON, err := json.MarshalIndent(orderedData, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("error marshaling ordered JSON: %v", err)
	}

	return orderedJSON, nil
}

// sortEnumValues finds and sorts enum arrays in components.schemas
func sortEnumValues(data any) {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return
	}

	// Navigate to components.schemas
	components, ok := dataMap["components"].(map[string]interface{})
	if !ok {
		return
	}

	schemas, ok := components["schemas"].(map[string]interface{})
	if !ok {
		return
	}

	// Iterate through each schema
	for _, schemaValue := range schemas {
		sortEnumInSchema(schemaValue)
	}
}

// sortEnumInSchema recursively finds and sorts enum arrays in a schema
func sortEnumInSchema(schema any) {
	schemaMap, ok := schema.(map[string]interface{})
	if !ok {
		return
	}

	// Check if this schema has an enum field
	if enumArray, ok := schemaMap["enum"].([]interface{}); ok && len(enumArray) > 0 {
		schemaMap["enum"] = sortEnumArray(enumArray)
	}

	// Recursively process nested schemas (properties, items, allOf, oneOf, anyOf, etc.)
	if properties, ok := schemaMap["properties"].(map[string]interface{}); ok {
		for _, prop := range properties {
			sortEnumInSchema(prop)
		}
	}

	if items, ok := schemaMap["items"]; ok {
		sortEnumInSchema(items)
	}

	// Handle allOf, oneOf, anyOf
	for _, key := range []string{"allOf", "oneOf", "anyOf"} {
		if arrayValue, ok := schemaMap[key].([]interface{}); ok {
			for _, item := range arrayValue {
				sortEnumInSchema(item)
			}
		}
	}

	// Handle additionalProperties
	if additionalProps, ok := schemaMap["additionalProperties"]; ok {
		if additionalPropsMap, isMap := additionalProps.(map[string]interface{}); isMap {
			sortEnumInSchema(additionalPropsMap)
		}
	}
}

// sortEnumArray sorts an enum array based on JSON representation
func sortEnumArray(arr []interface{}) []interface{} {
	if len(arr) <= 1 {
		return arr
	}

	// Create a copy to avoid modifying the original during sort
	sorted := make([]interface{}, len(arr))
	copy(sorted, arr)

	// Simple bubble sort for small enum arrays
	for i := 0; i < len(sorted)-1; i++ {
		for j := 0; j < len(sorted)-i-1; j++ {
			if compareEnumValues(sorted[j], sorted[j+1]) > 0 {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}

	return sorted
}

// compareEnumValues compares two enum values
func compareEnumValues(a, b interface{}) int {
	// Convert to JSON strings for comparison
	aJSON, aErr := json.Marshal(a)
	bJSON, bErr := json.Marshal(b)

	if aErr != nil || bErr != nil {
		return 0
	}

	aStr := string(aJSON)
	bStr := string(bJSON)

	if aStr < bStr {
		return -1
	} else if aStr > bStr {
		return 1
	}
	return 0
}
