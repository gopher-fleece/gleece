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

type GleeceAnnotation = string

const (
	GleeceAnnotationTag             GleeceAnnotation = "Tag"
	GleeceAnnotationQuery           GleeceAnnotation = "Query"
	GleeceAnnotationPath            GleeceAnnotation = "Path"
	GleeceAnnotationBody            GleeceAnnotation = "Body"
	GleeceAnnotationHeader          GleeceAnnotation = "Header"
	GleeceAnnotationFormField       GleeceAnnotation = "FormField"
	GleeceAnnotationDeprecated      GleeceAnnotation = "Deprecated"
	GleeceAnnotationHidden          GleeceAnnotation = "Hidden"
	GleeceAnnotationSecurity        GleeceAnnotation = "Security"
	GleeceAnnotationRoute           GleeceAnnotation = "Route"
	GleeceAnnotationResponse        GleeceAnnotation = "Response"
	GleeceAnnotationDescription     GleeceAnnotation = "Description"
	GleeceAnnotationMethod          GleeceAnnotation = "Method"
	GleeceAnnotationErrorResponse   GleeceAnnotation = "ErrorResponse"
	GleeceAnnotationTemplateContext GleeceAnnotation = "TemplateContext"
	// AttributeAdvancedSecurity GleeceAnnotation = "AdvancedSecurity"
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

type NonAttributeComment struct {
	Index int
	Value string
}

type AnnotationHolder struct {
	attributes           []Attribute
	nonAttributeComments []NonAttributeComment
}

type CommentSource string

// controller, route, schema, property

const (
	CommentSourceController CommentSource = "controller"
	CommentSourceRoute      CommentSource = "route"
	CommentSourceSchema     CommentSource = "schema"
	CommentSourceProperty   CommentSource = "property"
)

// Captures: 1. TEXT (after @), 2. TEXT (inside parentheses), 3. JSON5 Object, 4. Remaining TEXT
var parsingRegex *regexp.Regexp = regexp.MustCompile(`^// @(\w+)(?:(?:\(([\w-_/\\{} ]+))(?:\s*,\s*(\{.*\}))?\))?(?:\s+(.+))?$`)

func NewAnnotationHolder(comments []string, commentSource CommentSource) (AnnotationHolder, error) {
	holder := AnnotationHolder{
		nonAttributeComments: make([]NonAttributeComment, 0),
	}

	for lineIndex, comment := range comments {
		attr, isAnAttribute, err := parseComment(parsingRegex, strings.TrimSpace(comment))
		if err != nil {
			return holder, err
		}

		if isAnAttribute {
			// Check that this is a valid attribute with valid properties
			if err := IsValidAnnotation(attr, commentSource); err != nil {
				return holder, err
			}
			holder.attributes = append(holder.attributes, attr)
		} else {
			holder.nonAttributeComments = append(
				holder.nonAttributeComments,
				NonAttributeComment{
					Index: lineIndex,
					Value: strings.Trim(strings.TrimPrefix(comment, "//"), " "),
				},
			)
		}
	}

	if err := IsValidAnnotationCollection(holder.attributes, commentSource); err != nil {
		return holder, err
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

func (holder AnnotationHolder) GetFirst(attribute GleeceAnnotation) *Attribute {
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

func (holder AnnotationHolder) GetDescription() string {
	descriptionAttr := holder.GetFirst(GleeceAnnotationDescription)
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
	for _, comment := range holder.nonAttributeComments {
		if comment.Index > lastFreeCommentIndex+1 {
			break
		}
		freeComments = append(freeComments, comment.Value)
		lastFreeCommentIndex++
	}

	takeUntil := len(freeComments)
	// Trim all empty comments at the end
	for i := len(freeComments); i > 0; i-- {
		if freeComments[i-1] != "" {
			break
		}
		takeUntil--
	}
	return strings.Join(freeComments[:takeUntil], "\n")
}
