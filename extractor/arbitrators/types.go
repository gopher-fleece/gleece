package arbitrators

import (
	"go/ast"

	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/extractor/annotations"
)

type TypeMetadataWithAst struct {
	definitions.TypeMetadata
	// The AST expression for the type itself
	TypeExpr    ast.Expr
	Annotations *annotations.AnnotationHolder
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

	// The AST expression for the parameter itself
	ParamExpr ast.Expr
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

	// The AST expression for the return value itself
	Expr ast.Expr
}

func (s FuncReturnValueWithAst) Reduce() definitions.FuncReturnValue {
	return definitions.FuncReturnValue{
		OrderedIdent:       s.OrderedIdent,
		TypeMetadata:       s.TypeMetadata,
		UniqueImportSerial: s.UniqueImportSerial,
	}
}
