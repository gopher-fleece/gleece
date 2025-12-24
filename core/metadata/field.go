package metadata

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/gopher-fleece/gleece/v2/common"
	"github.com/gopher-fleece/gleece/v2/core/annotations"
	"github.com/gopher-fleece/gleece/v2/definitions"
	"github.com/gopher-fleece/gleece/v2/gast"
)

type FieldMeta struct {
	SymNodeMeta
	Type       TypeUsageMeta
	IsEmbedded bool
}

func (f FieldMeta) Reduce(_ ReductionContext) (definitions.FieldMetadata, error) {
	fieldNode, ok := f.Node.(*ast.Field)
	if !ok {
		return definitions.FieldMetadata{}, fmt.Errorf("field '%s' has a non-field node type", f.Name)
	}

	var tag string
	if fieldNode != nil && fieldNode.Tag != nil {
		tag = strings.Trim(fieldNode.Tag.Value, "`")
	}

	return definitions.FieldMetadata{
		Name:        f.Name,
		Type:        f.Type.Root.SimpleTypeString(),
		Description: annotations.GetDescription(f.Annotations),
		Tag:         tag,
		IsEmbedded:  f.IsEmbedded,
		Deprecation: common.Ptr(GetDeprecationOpts(f.Annotations)),
	}, nil
}

func (m TypeUsageMeta) IsUniverseType() bool {
	return gast.IsUniverseType(m.Name)
}

func (m TypeUsageMeta) IsByAddress() bool {
	_, isStar := m.Node.(*ast.StarExpr)
	return isStar
}
