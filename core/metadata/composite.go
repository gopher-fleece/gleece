package metadata

import "github.com/gopher-fleece/gleece/graphs"

// CompositeMeta holds minimal metadata for composite-type nodes.
// Stored in SymbolNode.Data when creating composite nodes.
type CompositeMeta struct {
	Canonical string
	Operands  []graphs.SymbolKey
}
