package symboldg

type SymbolEdgeKind string

const (
	EdgeKindContains   SymbolEdgeKind = "cnt"
	EdgeKindReferences SymbolEdgeKind = "ref"
	EdgeKindCalls      SymbolEdgeKind = "call"
	EdgeKindEmbeds     SymbolEdgeKind = "embed"
)

type SymbolEdge struct {
	To       SymbolKey
	Kind     SymbolEdgeKind
	Metadata map[string]string
}
