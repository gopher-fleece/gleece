package symboldg

import (
	"github.com/gopher-fleece/gleece/common"
	"github.com/gopher-fleece/gleece/core/annotations"
	"github.com/gopher-fleece/gleece/gast"
	"github.com/gopher-fleece/gleece/graphs"
)

type SymbolNode struct {
	Id          graphs.SymbolKey
	Kind        common.SymKind
	Version     *gast.FileVersion
	Data        any // Actual metadata: RouteMetadata, TypeMetadata, etc.
	Annotations *annotations.AnnotationHolder
}
