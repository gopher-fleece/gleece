package arbitrators

import (
	"go/ast"

	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/extractor/annotations"
	"github.com/gopher-fleece/gleece/gast"
)

type TypeMetadataWithAst struct {
	definitions.TypeMetadata
	// The AST node for the type itself
	TypeExpr    ast.Node
	Annotations *annotations.AnnotationHolder
	FVersion    *gast.FileVersion
}

func (s TypeMetadataWithAst) Equals(other TypeMetadataWithAst) bool {
	if s.TypeExpr != other.TypeExpr {
		return false
	}

	// Is this... correct?
	if s.Annotations != other.Annotations {
		return false
	}

	if !s.FVersion.Equals(other.FVersion) {
		return false
	}

	return s.TypeMetadata.Equals(other.TypeMetadata)
}

func (s TypeMetadataWithAst) Reduce() definitions.TypeMetadata {
	return s.TypeMetadata
}

type ParamMetaWithAst struct {
	TypeMetadataWithAst
	definitions.OrderedIdent
	Name      string
	IsContext bool
}

func (s ParamMetaWithAst) Reduce() definitions.ParamMeta {
	return definitions.ParamMeta{
		OrderedIdent: s.OrderedIdent,
		Name:         s.Name,
		IsContext:    s.IsContext,
		TypeMeta:     s.TypeMetadata,
	}
}

// This struct describes a function parameter's metadata with Gleece's additions.
type FuncParamWithAst struct {
	ParamMetaWithAst
	PassedIn           definitions.ParamPassedIn
	NameInSchema       string
	Description        string
	UniqueImportSerial uint64
	Validator          string
	Deprecation        *definitions.DeprecationOptions

	// The AST node for the parameter itself
	ParamExpr ast.Node
}

func (s FuncParamWithAst) Reduce() definitions.FuncParam {
	return definitions.FuncParam{
		ParamMeta:          s.ParamMetaWithAst.Reduce(),
		PassedIn:           s.PassedIn,
		NameInSchema:       s.NameInSchema,
		Description:        s.Description,
		UniqueImportSerial: s.UniqueImportSerial,
		Validator:          s.Validator,
		Deprecation:        s.Deprecation,
	}
}

type FuncReturnValueWithAst struct {
	definitions.OrderedIdent
	TypeMetadataWithAst
	UniqueImportSerial uint64

	// The AST node for the return value itself
	RetValExpr ast.Node
}

func (s FuncReturnValueWithAst) Reduce() definitions.FuncReturnValue {
	return definitions.FuncReturnValue{
		OrderedIdent:       s.OrderedIdent,
		TypeMetadata:       s.TypeMetadata,
		UniqueImportSerial: s.UniqueImportSerial,
	}
}
