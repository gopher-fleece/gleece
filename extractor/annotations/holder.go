package annotations

import (
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
	AttributePath             = "Path"
	AttributeBody             = "Body"
	AttributeHeader           = "Header"
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

type AnnotationHolder struct {
	attributes           []Attribute
	nonAttributeComments map[int]string
}

func NewAnnotationHolder(comments []string) (AnnotationHolder, error) {
	// Captures: 1. TEXT (after @), 2. TEXT (inside parentheses), 3. JSON5 Object, 4. Remaining TEXT
	parsingRegex := regexp.MustCompile(`^// @(\w+)(?:(?:\(([\w-_/\\{} ]+))(?:\s*,\s*(\{.*\}))?\))?(?:\s+(.+))?$`)

	holder := AnnotationHolder{
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

func (holder AnnotationHolder) GetFirst(attribute string) *Attribute {
	for _, attrib := range holder.attributes {
		if attrib.Name == attribute {
			return &attrib
		}
	}

	return nil
}

func (holder AnnotationHolder) GetFirstValueOrEmpty(attribute string) string {
	attrib := holder.GetFirst(attribute)
	if attrib == nil {
		return ""
	}

	return attrib.Value
}

func (holder AnnotationHolder) GetFirstDescriptionOrEmpty(attribute string) string {
	attrib := holder.GetFirst(attribute)
	if attrib == nil {
		return ""
	}

	return attrib.Description
}

func (holder AnnotationHolder) GetAll(attribute string) []*Attribute {
	attributes := []*Attribute{}
	for _, attrib := range holder.attributes {
		if attrib.Name == attribute {
			attributes = append(attributes, &attrib)
		}
	}

	return attributes
}

func (holder AnnotationHolder) Has(attribute string) bool {
	return holder.GetFirst(attribute) != nil
}

func (holder AnnotationHolder) FindFirstByValue(value string) *Attribute {
	for _, attrib := range holder.attributes {
		if attrib.Value == value {
			return &attrib
		}
	}
	return nil
}

func (holder AnnotationHolder) FindFirstByProperty(key string, value string) *Attribute {
	for _, attrib := range holder.attributes {
		if attrib.Properties[key] == value {
			return &attrib
		}
	}
	return nil
}

func (holder AnnotationHolder) FindByValueOrProperty(key string, value string) *Attribute {
	for _, attrib := range holder.attributes {
		if attrib.Value == value || attrib.Properties[key] == value {
			return &attrib
		}
	}
	return nil
}

func (holder AnnotationHolder) GetFirstPropertyValueOrEmpty(property string) string {
	prop := holder.GetFirst(property)
	if prop != nil {
		return prop.Value
	}
	return ""
}

func (holder AnnotationHolder) GetDescription() string {
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