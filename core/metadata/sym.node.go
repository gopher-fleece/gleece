package metadata

import (
	"go/ast"

	"github.com/gopher-fleece/gleece/common"
	"github.com/gopher-fleece/gleece/core/annotations"
	"github.com/gopher-fleece/gleece/gast"
)

type SymNodeMeta struct {
	Name        string
	Node        ast.Node
	SymbolKind  common.SymKind
	PkgPath     string
	Annotations *annotations.AnnotationHolder
	// The node's resolved range in the file.
	// Note this information may or may not be available
	Range    common.ResolvedRange
	FVersion *gast.FileVersion
}
