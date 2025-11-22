package symboldg

import (
	"github.com/gopher-fleece/gleece/common"
	"github.com/gopher-fleece/gleece/core/metadata"
	"github.com/gopher-fleece/gleece/graphs"
	"github.com/gopher-fleece/gleece/graphs/dot"
)

type TreeWalker func(node *SymbolNode, edges map[string]SymbolEdgeDescriptor) *SymbolNode

type SymbolGraphBuilder interface {
	AddController(request CreateControllerNode) (*SymbolNode, error)
	AddRoute(request CreateRouteNode) (*SymbolNode, error)
	AddRouteParam(request CreateParameterNode) (*SymbolNode, error)
	AddRouteRetVal(request CreateReturnValueNode) (*SymbolNode, error)
	AddStruct(request CreateStructNode) (*SymbolNode, error)
	AddEnum(request CreateEnumNode) (*SymbolNode, error)
	AddField(request CreateFieldNode) (*SymbolNode, error)
	AddConst(request CreateConstNode) (*SymbolNode, error)
	AddAlias(request CreateAliasNode) (*SymbolNode, error)
	AddPrimitive(kind PrimitiveType) *SymbolNode
	AddSpecial(special SpecialType) *SymbolNode
	AddEdge(from, to graphs.SymbolKey, kind SymbolEdgeKind, meta map[string]string)
	RemoveEdge(from, to graphs.SymbolKey, kind *SymbolEdgeKind)
	RemoveNode(key graphs.SymbolKey)

	Structs() []metadata.StructMeta
	Enums() []metadata.EnumMeta

	Exists(key graphs.SymbolKey) bool
	Get(key graphs.SymbolKey) *SymbolNode
	GetEdges(key graphs.SymbolKey, kinds []SymbolEdgeKind) map[string]SymbolEdgeDescriptor
	FindByKind(kind common.SymKind) []*SymbolNode

	IsPrimitivePresent(primitive PrimitiveType) bool
	IsSpecialPresent(special SpecialType) bool

	// Children returns direct outward SymbolNode dependencies from the given node,
	// applying the given traversal behavior if non-nil.
	Children(node *SymbolNode, behavior *TraversalBehavior) []*SymbolNode

	// Parents returns nodes that have edges pointing to the given node,
	// applying the given traversal behavior if non-nil.
	Parents(node *SymbolNode, behavior *TraversalBehavior) []*SymbolNode

	// Descendants returns all transitive children reachable from root,
	// applying the behavior at each step to decide traversal and inclusion.
	Descendants(root *SymbolNode, behavior *TraversalBehavior) []*SymbolNode

	Walk(root *SymbolNode, callback TreeWalker)

	ToDot(theme *dot.DotTheme) string
	String() string
}
