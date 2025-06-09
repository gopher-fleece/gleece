package symboldg

import (
	"github.com/gopher-fleece/gleece/common"
	"github.com/gopher-fleece/gleece/extractor/annotations"
	"github.com/gopher-fleece/gleece/gast"
)

type SymbolNode struct {
	Id          SymbolKey
	Kind        common.SymKind
	Version     *gast.FileVersion
	Data        any // Actual metadata: RouteMetadata, TypeMetadata, etc.
	Annotations *annotations.AnnotationHolder
}
