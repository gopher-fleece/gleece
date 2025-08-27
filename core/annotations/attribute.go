package annotations

import "github.com/gopher-fleece/gleece/gast"

type NonAttributeComment struct {
	Index   int
	Value   string
	Comment gast.CommentNode
}

type Attribute struct {
	Name        string
	Value       string
	Properties  map[string]any
	Description string

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
