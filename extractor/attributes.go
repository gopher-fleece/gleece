package extractor

import (
	"regexp"

	"github.com/titanous/json5"
)

const (
	PropertyName = "name"
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
	Properties  map[string]string
	Description string
}

func (attr Attribute) HasProperty(name string) bool {
	return attr.GetProperty(name) != nil
}

func (attr Attribute) GetProperty(name string) *string {
	value, exists := attr.Properties[name]
	if exists {
		return &value
	}
	return nil
}

type AttributesHolder struct {
	attributes []Attribute
}

func NewAttributeHolder(comments []string) (AttributesHolder, error) {
	// Regular expression to capture the different parts of the input.
	// Captures: 1. TEXT (after @), 2. TEXT (inside parentheses), 3. JSON5 Object, 4. Remaining TEXT
	parsingRegex := regexp.MustCompile(`^// @(\w+)(?:(?:\(([\w-_/\\{}]+))(?:\s*,\s*(\{.*\}))?\))?(?:\s+(.+))?$`)
	// TODO tomorrow: Regex doesn't handle all cases...

	holder := AttributesHolder{}
	for _, comment := range comments {
		attr, isAnAttribute, err := parseComment(parsingRegex, comment)
		if isAnAttribute {
			if err != nil {
				return holder, err
			}
			holder.attributes = append(holder.attributes, attr)
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

	var props map[string]string
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
