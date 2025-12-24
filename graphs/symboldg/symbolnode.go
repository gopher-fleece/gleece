package symboldg

import (
	"github.com/gopher-fleece/gleece/v2/common"
	"github.com/gopher-fleece/gleece/v2/core/annotations"
	"github.com/gopher-fleece/gleece/v2/gast"
	"github.com/gopher-fleece/gleece/v2/graphs"
)

type SymbolNode struct {
	Id          graphs.SymbolKey
	Kind        common.SymKind
	Version     *gast.FileVersion
	Data        any // Actual metadata: RouteMetadata, TypeMetadata, etc.
	Annotations *annotations.AnnotationHolder
}

type SymbolNodeWithOrdinal struct {
	Node    *SymbolNode
	Ordinal uint32
}
