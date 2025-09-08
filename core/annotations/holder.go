package annotations

import (
	"regexp"
	"strings"

	"github.com/gopher-fleece/gleece/common"
	"github.com/gopher-fleece/gleece/gast"
	"github.com/titanous/json5"
)

type AnnotationHolder struct {
	// The file name from which the comments were retrieved.
	//
	// Note that this does not attempt to consider cases where AST was 'messed with' or corrupted
	// as well as //line comments
	fileName string

	// The source from which the comments held by the struct were retrieved.
	// This affects downstream validation
	source CommentSource

	attributes           []Attribute
	nonAttributeComments []NonAttributeComment

	docRange common.ResolvedRange
}

// Captures: 1. TEXT (after @), 2. TEXT (inside parentheses), 3. JSON5 Object, 4. Remaining TEXT
var parsingRegex *regexp.Regexp = regexp.MustCompile(`^// @(\w+)(?:(?:\(([\w-_/\\{} ]+))(?:\s*,\s*(\{.*\}))?\))?(?:\s+(.+))?$`)

// NewAnnotationHolder creates a holder from CommentNode entries (with positions).
func NewAnnotationHolder(commentBlock gast.CommentBlock, source CommentSource) (AnnotationHolder, error) {
	holder := AnnotationHolder{
		fileName:             commentBlock.FileName,
		source:               source,
		nonAttributeComments: make([]NonAttributeComment, 0),
		docRange:             commentBlock.Range,
	}

	for _, comment := range commentBlock.Comments {
		attr, isAnAttribute, err := parseCommentNode(parsingRegex, comment)
		if err != nil {
			return holder, err
		}

		if isAnAttribute {
			// TODO - REMOVE VALIDATION HERE, MOVE TO VALIDATOR LOGIC
			// Validation step left as-is
			/*
				if err := IsValidAnnotation(attr, commentSource); err != nil {
					return holder, err
				}*/
			holder.attributes = append(holder.attributes, attr)
		} else {
			holder.nonAttributeComments = append(
				holder.nonAttributeComments,
				NonAttributeComment{
					Comment: comment,
					Index:   comment.Index,
					Value:   strings.Trim(strings.TrimPrefix(comment.Text, "//"), " "),
				},
			)
		}
	}

	/*
		if err := IsValidAnnotationCollection(holder.attributes, commentSource); err != nil {
			return holder, err
		}
	*/

	return holder, nil
}

// NewAnnotationHolderFromData creates an annotation holder from pre-parsed attributes and comments.
// Note that no validations are done here - this is mostly to help with tests
func NewAnnotationHolderFromData(
	attributes []Attribute,
	nonAttributeComments []NonAttributeComment,
) AnnotationHolder {
	return AnnotationHolder{
		attributes:           attributes,
		nonAttributeComments: nonAttributeComments,
	}
}

func (holder AnnotationHolder) FileName() string {
	return holder.fileName
}

func (holder AnnotationHolder) Range() common.ResolvedRange {
	return holder.docRange
}

// Attributes returns a shallow copy of the holder's attributes
func (holder AnnotationHolder) Attributes() []Attribute {
	return append([]Attribute(nil), holder.attributes...)
}

// NonAttributeComments returns a shallow copy of the holder's non-attributes
func (holder AnnotationHolder) NonAttributeComments() []NonAttributeComment {
	return append([]NonAttributeComment(nil), holder.nonAttributeComments...)
}

func (holder AnnotationHolder) AttributeCounts() map[string]int {
	result := make(map[string]int)
	for _, attr := range holder.attributes {

		_, exists := result[attr.Name]
		if exists {
			result[attr.Name]++
		} else {
			result[attr.Name] = 1
		}
	}
	return result
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

func parseCommentNode(parsingRegex *regexp.Regexp, comment gast.CommentNode) (Attribute, bool, error) {
	text := strings.TrimSpace(comment.Text)
	matches := parsingRegex.FindStringSubmatch(text)

	if len(matches) == 0 {
		return Attribute{}, false, nil
	}

	attributeName := matches[1]
	primaryValue := matches[2]
	jsonConfig := matches[3]
	description := matches[4]

	var props map[string]any
	if len(jsonConfig) > 0 {
		if err := json5.Unmarshal([]byte(jsonConfig), &props); err != nil {
			return Attribute{}, true, err
		}
	}

	return Attribute{
		Name:        attributeName,
		Value:       primaryValue,
		Properties:  props,
		Description: description,
		Comment:     comment,
	}, true, nil
}
