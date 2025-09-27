package metadata

import (
	"go/ast"
	"strings"

	"github.com/gopher-fleece/gleece/common"
	"github.com/gopher-fleece/gleece/core/annotations"
	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/gast"
)

type FieldMeta struct {
	SymNodeMeta
	Type       TypeUsageMeta
	IsEmbedded bool
}

func (f FieldMeta) Reduce() definitions.FieldMetadata {
	fieldNode, ok := f.Node.(*ast.Field)
	if !ok {
		// Reduce has a pretty nice signature so pretty reluctant to hole it with an added error
		panic("field %s has a non-field node type")
	}

	var tag string
	if fieldNode != nil && fieldNode.Tag != nil {
		tag = strings.Trim(fieldNode.Tag.Value, "`")
	}

	decoratedType := f.Type.GetArrayLayersString() + f.Type.Name
	return definitions.FieldMetadata{
		Name:        f.Name,
		Type:        decoratedType,
		Description: annotations.GetDescription(f.Annotations),
		Tag:         tag,
		IsEmbedded:  f.IsEmbedded,
		Deprecation: common.Ptr(GetDeprecationOpts(f.Annotations)),
	}
}

func (m TypeUsageMeta) IsUniverseType() bool {
	return gast.IsUniverseType(m.Name)
}

func (m TypeUsageMeta) IsByAddress() bool {
	_, isStar := m.Node.(*ast.StarExpr)
	return isStar
}
