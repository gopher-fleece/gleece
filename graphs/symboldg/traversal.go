package symboldg

import "github.com/gopher-fleece/gleece/common"

type TraversalResultSorting int

const (
	TraversalSortingNone TraversalResultSorting = iota
	TraversalSortingOrdinalAsc
	TraversalSortingOrdinalDesc
)

type TraversalFilter struct {
	NodeKind   *common.SymKind        // Only include nodes of this kind (optional)
	EdgeKind   *SymbolEdgeKind        // Only follow edges of this kind (optional)
	FilterFunc func(*SymbolNode) bool // Optional user-defined predicate
}

type TraversalBehavior struct {
	Filtering TraversalFilter
	Sorting   TraversalResultSorting
}

func shouldIncludeEdge(edge SymbolEdge, behavior *TraversalBehavior) bool {
	return behavior == nil || behavior.Filtering.EdgeKind == nil || edge.Kind == *behavior.Filtering.EdgeKind
}

func shouldIncludeNode(node *SymbolNode, behavior *TraversalBehavior) bool {
	if behavior == nil {
		return true
	}

	if behavior.Filtering.NodeKind != nil && node.Kind != *behavior.Filtering.NodeKind {
		return false
	}
	if behavior.Filtering.FilterFunc != nil && !behavior.Filtering.FilterFunc(node) {
		return false
	}
	return true
}
