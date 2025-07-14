package symboldg

import "github.com/gopher-fleece/gleece/graphs"

type SymbolEdgeKind string

const (
	EdgeKindContains   SymbolEdgeKind = "cnt"
	EdgeKindReferences SymbolEdgeKind = "ref"
	EdgeKindCalls      SymbolEdgeKind = "call"
	EdgeKindEmbeds     SymbolEdgeKind = "embed"
)

type SymbolEdge struct {
	From     graphs.SymbolKey
	To       graphs.SymbolKey
	Kind     SymbolEdgeKind
	Metadata map[string]string
}
