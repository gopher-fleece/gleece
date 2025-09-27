package symboldg

import "github.com/gopher-fleece/gleece/graphs"

type SymbolEdgeKind string

const (
	// Structural/ownership relationships
	EdgeKindContains SymbolEdgeKind = "cnt"   // Parent contains child (e.g., controller → route)
	EdgeKindField    SymbolEdgeKind = "fld"   // Struct → field (if fields are nodes)
	EdgeKindReceiver SymbolEdgeKind = "recv"  // Struct → receiver function (method)
	EdgeKindEmbed    SymbolEdgeKind = "embed" // Struct → embedded type (anonymous field)

	// Type and usage relationships
	EdgeKindReference SymbolEdgeKind = "ref"     // General type reference (param → type)
	EdgeKindType      SymbolEdgeKind = "ty"      // Specific "has type" relationship
	EdgeKindTypeParam SymbolEdgeKind = "typaram" // A "has generic type parameter" relationship
	EdgeKindAlias     SymbolEdgeKind = "alias"   // Symbol → aliased symbol

	// Code behavior
	EdgeKindCall SymbolEdgeKind = "call" // Func → called func

	// Semantic constructs
	EdgeKindParam    SymbolEdgeKind = "param" // Func/Route → parameter node
	EdgeKindRetVal   SymbolEdgeKind = "ret"   // Func/Route → return value
	EdgeKindValue    SymbolEdgeKind = "val"   // Enum → individual enum value
	EdgeKindDocument SymbolEdgeKind = "doc"   // Symbol → doc/annotation node

	// Misc
	EdgeKindInit SymbolEdgeKind = "init" // Struct → init function (e.g., for defaults)
)

type SymbolEdge struct {
	From     graphs.SymbolKey
	To       graphs.SymbolKey
	Kind     SymbolEdgeKind
	Metadata map[string]string
}

type SymbolEdgeDescriptor struct {
	// The edge data itself
	Edge SymbolEdge

	// The ordinal representing the order/index in which the edge was inserted
	Ordinal uint32
}
