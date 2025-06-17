package visitors

import (
	"go/ast"

	"github.com/gopher-fleece/gleece/common"
	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/extractor/annotations"
	"github.com/gopher-fleece/gleece/extractor/arbitrators"
)

type StructMetadataWithAst struct {
	Name        string
	Node        ast.Node
	PkgPath     string
	Fields      []FieldMetadataWithAst
	Annotations *annotations.AnnotationHolder
}

type EnumMetadataWithAst struct {
	Name        string
	Node        ast.Node
	PkgPath     string
	Values      []string
	Type        arbitrators.TypeMetadataWithAst
	Annotations *annotations.AnnotationHolder
}

type FieldMetadataWithAst struct {
	Name        string
	Node        ast.Node
	Type        arbitrators.TypeMetadataWithAst
	Tag         string
	IsEmbedded  bool
	Annotations *annotations.AnnotationHolder
}

func (s StructMetadataWithAst) Reduce() definitions.StructMetadata {
	reducedFields := []definitions.FieldMetadata{}
	for _, f := range s.Fields {
		reducedFields = append(reducedFields, f.Reduce())
	}

	return definitions.StructMetadata{
		Name:        s.Name,
		PkgPath:     s.PkgPath,
		Description: s.Annotations.GetDescription(),
		Fields:      reducedFields,
		Deprecation: getDeprecationOpts(s.Annotations),
	}
}

func (s EnumMetadataWithAst) Reduce() definitions.EnumMetadata {
	return definitions.EnumMetadata{
		Name:        s.Name,
		PkgPath:     s.PkgPath,
		Description: s.Annotations.GetDescription(),
		Values:      s.Values,
		Type:        s.Type.Name,
		Deprecation: getDeprecationOpts(s.Annotations),
	}
}

func (s FieldMetadataWithAst) Reduce() definitions.FieldMetadata {
	return definitions.FieldMetadata{
		Name:        s.Name,
		Description: s.Annotations.GetDescription(),
		Tag:         s.Tag,
		IsEmbedded:  s.IsEmbedded,
		Deprecation: common.Ptr(getDeprecationOpts(s.Annotations)),
	}
}
