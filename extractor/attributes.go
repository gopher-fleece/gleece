package extractor

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/titanous/json5"
)

const (
	PropertyName            = "name"
	PropertySecurityScopes  = "scopes"
	PropertyValidatorString = "validate"
)

const (
	AttributeTag              = "Tag"
	AttributeQuery            = "Query"
	AttributeDeprecated       = "Deprecated"
	AttributeHidden           = "Hidden"
	AttributeSecurity         = "Security"
	AttributeAdvancedSecurity = "AdvancedSecurity"
	AttributeRoute            = "Route"
	AttributeResponse         = "Response"
	AttributeDescription      = "Description"
	AttributeMethod           = "Method"
	AttributeErrorResponse    = "ErrorResponse"
)

type Attribute struct {
	Name        string
	Value       string
	Properties  map[string]any
	Description string
}

func (attr Attribute) HasProperty(name string) bool {
	return attr.GetProperty(name) != nil
}

func (attr Attribute) GetProperty(name string) *any {
	value, exists := attr.Properties[name]
	if exists {
		return &value
	}
	return nil
}

type AttributesHolder struct {
	attributes           []Attribute
	nonAttributeComments map[int]string
}

func NewAttributeHolder(comments []string) (AttributesHolder, error) {
	// Captures: 1. TEXT (after @), 2. TEXT (inside parentheses), 3. JSON5 Object, 4. Remaining TEXT
	parsingRegex := regexp.MustCompile(`^// @(\w+)(?:(?:\(([\w-_/\\{} ]+))(?:\s*,\s*(\{.*\}))?\))?(?:\s+(.+))?$`)

	holder := AttributesHolder{
		nonAttributeComments: make(map[int]string),
	}

	for lineIndex, comment := range comments {
		attr, isAnAttribute, err := parseComment(parsingRegex, comment)
		if err != nil {
			return holder, err
		}

		if isAnAttribute {
			holder.attributes = append(holder.attributes, attr)
		} else {
			holder.nonAttributeComments[lineIndex] = strings.Trim(strings.TrimPrefix(comment, "//"), " ")
		}
	}

	return holder, nil
}

func parseComment(parsingRegex *regexp.Regexp, comment string) (Attribute, bool, error) {
	matches := parsingRegex.FindStringSubmatch(comment)

	if len(matches) == 0 {
		return Attribute{}, false, nil
	}

	// Extract matched groups
	attributeName := matches[1] // The TEXT after @ (e.g., Query)
	primaryValue := matches[2]  // The TEXT inside parentheses (e.g., someValue)
	jsonConfig := matches[3]    // The JSON5 object (e.g., {someProp: v1})
	description := matches[4]   // The remaining TEXT (e.g., some description)

	var props map[string]any
	if len(jsonConfig) > 0 {
		err := json5.Unmarshal([]byte(jsonConfig), &props)
		if err != nil {
			return Attribute{}, true, err
		}
	}

	// Return the parsed parts
	return Attribute{
		Name:        attributeName,
		Value:       primaryValue,
		Properties:  props,
		Description: description,
	}, true, nil
}

func (holder AttributesHolder) GetFirst(attribute string) *Attribute {
	for _, attrib := range holder.attributes {
		if attrib.Name == attribute {
			return &attrib
		}
	}

	return nil
}

func (holder AttributesHolder) GetFirstValueOrEmpty(attribute string) string {
	attrib := holder.GetFirst(attribute)
	if attrib == nil {
		return ""
	}

	return attrib.Value
}

func (holder AttributesHolder) GetFirstDescriptionOrEmpty(attribute string) string {
	attrib := holder.GetFirst(attribute)
	if attrib == nil {
		return ""
	}

	return attrib.Description
}

func (holder AttributesHolder) GetAll(attribute string) []*Attribute {
	attributes := []*Attribute{}
	for _, attrib := range holder.attributes {
		if attrib.Name == attribute {
			attributes = append(attributes, &attrib)
		}
	}

	return attributes
}

func (holder AttributesHolder) Has(attribute string) bool {
	return holder.GetFirst(attribute) != nil
}

func (holder AttributesHolder) FindFirstByValue(value string) *Attribute {
	for _, attrib := range holder.attributes {
		if attrib.Value == value {
			return &attrib
		}
	}
	return nil
}

func (holder AttributesHolder) FindFirstByProperty(key string, value string) *Attribute {
	for _, attrib := range holder.attributes {
		if attrib.Properties[key] == value {
			return &attrib
		}
	}
	return nil
}

func (holder AttributesHolder) FindByValueOrProperty(key string, value string) *Attribute {
	for _, attrib := range holder.attributes {
		if attrib.Value == value || attrib.Properties[key] == value {
			return &attrib
		}
	}
	return nil
}

func (holder AttributesHolder) GetFirstPropertyValueOrEmpty(property string) string {
	prop := holder.GetFirst(property)
	if prop != nil {
		return prop.Value
	}
	return ""
}

func (holder AttributesHolder) GetDescription() string {
	descriptionAttr := holder.GetFirst(AttributeDescription)
	if descriptionAttr != nil {
		// If there's a description attribute, even an empty one, use that
		return descriptionAttr.Description
	}

	freeComments := []string{}

	lastFreeCommentIndex := -1
	// Gather all non-attribute comments starting from position 0;
	// Once there's a break in continuity, break.
	// Example:
	// We've a non-attribute comment on line #0, #1 and #3.
	// We start at -1 so 0 is included. On the next iteration, #1 is included as well.
	// Then, line #2 is ignored as it doesn't have a comment and #3 breaks as index is at #1.
	for index, comment := range holder.nonAttributeComments {
		if index > lastFreeCommentIndex+1 {
			break
		}
		freeComments = append(freeComments, comment)
		lastFreeCommentIndex++
	}

	return strings.Join(freeComments, "\n")
}

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
		return getSliceProperty[TPropertyType](value, targetType)
	}

	castParam, castOk := (*value).(TPropertyType)
	if castOk {
		return &castParam, nil
	}

	return nil, nil
}
