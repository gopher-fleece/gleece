package metadata

import (
	"go/ast"

	"github.com/gopher-fleece/gleece/v2/common"
	"github.com/gopher-fleece/gleece/v2/core/annotations"
	"github.com/gopher-fleece/gleece/v2/gast"
)

// A symbol's general properties
//
// This struct serves as the HIR's backbone and the basis for nearly all symbols; They all have a name, an associated AST node/s, a kind and so forth.
type SymNodeMeta struct {
	// The symbol's name.
	// May be empty for anonymous fields like return types.
	Name string
	// The AST node associated with this symbol. Common examples are Fields, Structs, Funcs and so forth.
	Node       ast.Node
	SymbolKind common.SymKind
	// The package path in which this symbol resides
	PkgPath string
	// Any Gleece annotations decorating this symbol
	Annotations *annotations.AnnotationHolder
	// The node's resolved range in the file.
	Range common.ResolvedRange

	// The FileVersion for the file in which this symbol is located.
	//
	// Currently, this field is only partially used (though fully populated).
	// Its purpose is mostly to allow tracking file changes to perform graph prunes and incremental rebuilds.
	FVersion *gast.FileVersion
}
