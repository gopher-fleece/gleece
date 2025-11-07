package symboldg

import (
	"slices"

	"github.com/gopher-fleece/gleece/common"
)

type TraversalResultSorting int

const (
	TraversalSortingNone TraversalResultSorting = iota
	TraversalSortingOrdinalAsc
	TraversalSortingOrdinalDesc
)

type TraversalFilter struct {
	NodeKinds  []common.SymKind       // Only include nodes of this kind (optional)
	EdgeKinds  []SymbolEdgeKind       // Only follow edges of this kind (optional)
	FilterFunc func(*SymbolNode) bool // Optional user-defined predicate
}

type TraversalBehavior struct {
	Filtering TraversalFilter
	Sorting   TraversalResultSorting
}

func shouldIncludeEdge(edge SymbolEdge, behavior *TraversalBehavior) bool {
	if behavior == nil || behavior.Filtering.EdgeKinds == nil {
		return true
	}

	return slices.Contains(behavior.Filtering.EdgeKinds, edge.Kind)
}

func shouldIncludeNode(node *SymbolNode, behavior *TraversalBehavior) bool {
	if behavior == nil {
		return true
	}

	if behavior.Filtering.NodeKinds != nil && !slices.Contains(behavior.Filtering.NodeKinds, node.Kind) {
		return false
	}

	if behavior.Filtering.FilterFunc != nil && !behavior.Filtering.FilterFunc(node) {
		return false
	}
	
	return true
}
