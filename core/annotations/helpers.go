package annotations

import (
	"fmt"
	"reflect"
)

func getSliceProperty[TPropertyType any](value *any, targetType reflect.Type) (*TPropertyType, error) {
	// Ensure the value is also a slice
	if reflect.TypeOf(*value).Kind() != reflect.Slice {
		return nil, fmt.Errorf("value %v cannot be converted to type %s", value, targetType.String())
	}

	sourceSlice := reflect.ValueOf(*value)
	targetElemType := targetType.Elem()

	// Create a new slice of the target type
	convertedSlice := reflect.MakeSlice(targetType, sourceSlice.Len(), sourceSlice.Len())

	// Iterate through the source slice and convert each element
	for i := 0; i < sourceSlice.Len(); i++ {
		sourceElem := sourceSlice.Index(i).Interface()
		sourceElemValue := reflect.ValueOf(sourceElem)

		// Check if the source element can be converted to the target element type
		if !sourceElemValue.Type().ConvertibleTo(targetElemType) {
			return nil, fmt.Errorf("element %v at index %d cannot be converted to type %s", sourceElem, i, targetElemType.String())
		}

		// Convert the source element and set it in the target slice
		convertedElem := sourceElemValue.Convert(targetElemType)
		convertedSlice.Index(i).Set(convertedElem)
	}

	// Return the converted slice as the desired type
	converted := convertedSlice.Interface().(TPropertyType)
	return &converted, nil
}

func GetCastProperty[TPropertyType any](attrib *Attribute, property string) (*TPropertyType, error) {
	value := attrib.GetProperty(property)
	if value == nil {
		return nil, nil
	}

	targetType := reflect.TypeOf((*TPropertyType)(nil)).Elem()
	if targetType.Kind() == reflect.Slice {
		castValue, err := getSliceProperty[TPropertyType](value, targetType)
		if err != nil {
			return castValue, fmt.Errorf(
				"failed to cast attribute property '%v' to %v - %v",
				property,
				targetType.String(),
				err,
			)
		}
		return castValue, nil
	}

	castParam, castOk := (*value).(TPropertyType)
	if castOk {
		return &castParam, nil
	}

	return nil, fmt.Errorf("property '%s' exists but cannot be cast to %s", property, targetType.Name())
}
