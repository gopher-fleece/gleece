package symboldg

import (
	"go/ast"

	"github.com/gopher-fleece/gleece/core/annotations"
	"github.com/gopher-fleece/gleece/core/metadata"
	"github.com/gopher-fleece/gleece/gast"
	"github.com/gopher-fleece/gleece/graphs"
)

type KeyableNodeMeta struct {
	// Decl is the node's AST decl (*ast.Field, *ast.StructType)
	Decl ast.Node

	// FVersion is the version of the AST file the decl was found on
	FVersion gast.FileVersion
}

func (k KeyableNodeMeta) SymbolKey() graphs.SymbolKey {
	return graphs.NewSymbolKey(k.Decl, &k.FVersion)
}

type CreateControllerNode struct {
	Data        metadata.ControllerMeta
	Annotations *annotations.AnnotationHolder
}

type CreateRouteNode struct {
	Data             *metadata.ReceiverMeta
	Annotations      *annotations.AnnotationHolder
	ParentController KeyableNodeMeta
}

type CreateParameterNode struct {
	Data        metadata.FuncParam
	ParentRoute KeyableNodeMeta
}

type CreateReturnValueNode struct {
	Data        metadata.FuncReturnValue
	ParentRoute KeyableNodeMeta
}

type CreateStructNode struct {
	Data        metadata.StructMeta
	Annotations *annotations.AnnotationHolder
}

type CreateEnumNode struct {
	Data        metadata.EnumMeta
	Annotations *annotations.AnnotationHolder
}

type CreateFieldNode struct {
	Data        metadata.FieldMeta
	Annotations *annotations.AnnotationHolder
}

type CreateConstNode struct {
	Data        metadata.ConstMeta
	Annotations *annotations.AnnotationHolder
}

// CompositeMeta holds minimal metadata for composite-type nodes.
// Stored in SymbolNode.Data when creating composite nodes.
type CompositeMeta struct {
	Canonical string
	Operands  []graphs.SymbolKey
}

// CreateCompositeNode is the request used to add a canonical composite node.
type CreateCompositeNode struct {
	Key       graphs.SymbolKey
	Canonical string
	Operands  []graphs.SymbolKey
}
