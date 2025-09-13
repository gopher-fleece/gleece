package annotations

import (
	"regexp"
	"strings"
	"unicode/utf8"

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

func (holder AnnotationHolder) Source() CommentSource {
	return holder.source
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
	// use raw comment.Text so the regex byte offsets line up with Position
	text := comment.Text

	matchIndices := parsingRegex.FindStringSubmatchIndex(text)
	if matchIndices == nil {
		return Attribute{}, false, nil
	}

	name, _ := getGroupString(comment.Text, matchIndices, 1)                 // Attr. name
	value, _ := getGroupString(comment.Text, matchIndices, 2)                // Attr. Value (inside parentheses)
	jsonConfig, jsonPresent := getGroupString(comment.Text, matchIndices, 3) // Optional JSON5 props object
	description, _ := getGroupString(comment.Text, matchIndices, 4)          // Optional description

	var parsedProps map[string]any
	if jsonPresent && jsonConfig != "" {
		if err := json5.Unmarshal([]byte(jsonConfig), &parsedProps); err != nil {
			return Attribute{}, true, err
		}
	}

	return Attribute{
		Name:            name,
		Value:           value,
		Properties:      parsedProps,
		PropertiesRange: getPropertiesRange(comment, matchIndices),
		Description:     description,
		Comment:         comment,
	}, true, nil
}

func getPropertiesRange(comment gast.CommentNode, matchIndices []int) common.ResolvedRange {
	// if group 3 exists, compute its absolute ResolvedRange

	propsRange := common.ResolvedRange{}
	if startByte, endByte, ok := getGroupOffsets(matchIndices, 3); ok {
		startLine, startCol := byteOffsetToLineCol(comment.Text, startByte, comment.Position.StartLine, comment.Position.StartCol)
		endLine, endCol := byteOffsetToLineCol(comment.Text, endByte, comment.Position.StartLine, comment.Position.StartCol)

		propsRange = common.ResolvedRange{
			StartLine: startLine,
			StartCol:  startCol,
			EndLine:   endLine,
			EndCol:    endCol,
		}
	}

	return propsRange

}

// helper: return start,end and ok for group (1-based group index)
func getGroupOffsets(matchIndices []int, group int) (startByte, endByte int, ok bool) {
	pairIndex := 2 * group
	if pairIndex+1 >= len(matchIndices) {
		return 0, 0, false
	}
	startByte = matchIndices[pairIndex]
	endByte = matchIndices[pairIndex+1]
	if startByte < 0 || endByte < 0 {
		return 0, 0, false
	}
	return startByte, endByte, true
}

// helper: return string for a group if present
func getGroupString(commentText string, matchIndices []int, group int) (string, bool) {
	startByte, endByte, ok := getGroupOffsets(matchIndices, group)
	if !ok {
		return "", false
	}
	return commentText[startByte:endByte], true
}

// byteOffsetToLineCol converts a byte offset into (line, col).
// - Counts runes for columns.
// - Properly treats '\n', '\r', and '\r\n' as single line breaks.
// - startLine/startCol are the absolute coordinates of the first rune of s.
func byteOffsetToLineCol(s string, byteOffset int, startLine, startCol int) (line, col int) {
	line = startLine
	col = startCol

	i := 0
	// iterate until we reach the requested byte offset
	for i < byteOffset && i < len(s) {
		// fast-path ASCII newline checks first to handle CRLF simply
		if s[i] == '\r' {
			// CR or CRLF
			if i+1 < len(s) && s[i+1] == '\n' {
				// treat CRLF as a single newline
				i += 2
			} else {
				// lone CR
				i++
			}
			line++
			col = 0
			continue
		}
		if s[i] == '\n' {
			// lone LF
			i++
			line++
			col = 0
			continue
		}

		// decode next rune (handles multibyte UTF-8)
		runeValue, runeSize := utf8.DecodeRuneInString(s[i:])
		// defensive: if invalid rune, advance by 1 byte
		if runeValue == utf8.RuneError && runeSize == 1 {
			i++
			col++
			continue
		}
		i += runeSize
		col++
	}

	return
}
