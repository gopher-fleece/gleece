package annotations

import (
	"strings"
	"unicode/utf8"

	"github.com/gopher-fleece/gleece/v2/common"
	"github.com/gopher-fleece/gleece/v2/gast"
)

type NonAttributeComment struct {
	Index   int
	Value   string
	Comment gast.CommentNode
}

type Attribute struct {
	Name            string
	Value           string
	Properties      map[string]any
	PropertiesRange common.ResolvedRange
	Description     string

	// Position info for the comment that produced this attribute (may be zero)
	Comment gast.CommentNode
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

// GetValueRange returns the resolved range of the attribute's Value within the comment.
// If Value is empty or can't be located, falls back to the comment Range().
func (attr Attribute) GetValueRange() common.ResolvedRange {
	// Find the first occurrence of the value inside the comment text
	idx := -1
	if attr.Comment.Text != "" {
		idx = strings.Index(attr.Comment.Text, attr.Value)
	}

	// If not found, fallback to comment range
	if idx < 0 {
		return attr.Comment.Range()
	}

	// Compute rune-aware column offsets
	// number of runes before the match
	prefixRunes := utf8.RuneCountInString(attr.Comment.Text[:idx])
	valueRunes := utf8.RuneCountInString(attr.Value)

	startCol := attr.Comment.Position.StartCol + prefixRunes
	endCol := startCol + valueRunes

	return common.ResolvedRange{
		StartLine: attr.Comment.Position.StartLine,
		StartCol:  startCol,
		EndLine:   attr.Comment.Position.EndLine,
		EndCol:    endCol,
	}
}
