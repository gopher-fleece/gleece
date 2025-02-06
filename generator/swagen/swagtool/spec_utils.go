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
