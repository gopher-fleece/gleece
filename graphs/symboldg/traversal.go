package symboldg

import "github.com/gopher-fleece/gleece/common"

type TraversalFilter struct {
	NodeKind   *common.SymKind        // Only include nodes of this kind (optional)
	EdgeKind   *SymbolEdgeKind        // Only follow edges of this kind (optional)
	FilterFunc func(*SymbolNode) bool // Optional user-defined predicate
}

func shouldIncludeEdge(edge SymbolEdge, filter *TraversalFilter) bool {
	return filter == nil || filter.EdgeKind != nil || edge.Kind == *filter.EdgeKind
}

func shouldIncludeNode(node *SymbolNode, filter *TraversalFilter) bool {
	if filter == nil {
		return true
	}

	if filter.NodeKind != nil && node.Kind != *filter.NodeKind {
		return false
	}
	if filter.FilterFunc != nil && !filter.FilterFunc(node) {
		return false
	}
	return true
}
