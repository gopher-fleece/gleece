package utils

import (
	"fmt"
	"slices"

	"github.com/gopher-fleece/gleece/v2/common"
	"github.com/gopher-fleece/gleece/v2/common/linq"
	"github.com/gopher-fleece/gleece/v2/core/metadata"
	"github.com/gopher-fleece/gleece/v2/core/metadata/typeref"
	"github.com/gopher-fleece/gleece/v2/graphs/symboldg"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type ControllerInfo struct {
	Node *symboldg.SymbolNode
	Data metadata.ControllerMeta
}

type ReceiverInfo struct {
	Node *symboldg.SymbolNode
	Data metadata.ReceiverMeta
}

type FuncParamInfo struct {
	Node *symboldg.SymbolNode
	Data metadata.FieldMeta
}

type FuncRetValInfo struct {
	Node *symboldg.SymbolNode
	Data metadata.FieldMeta
}

type ApiEndpointInfo struct {
	Controller ControllerInfo
	Receiver   ReceiverInfo
	Params     []FuncParamInfo
	RetVals    []FuncRetValInfo
}

type TypeParamInstantiationInfo struct {
	Node        *symboldg.SymbolNode
	UsedInIndex int
}

// MustFindController finds a single controller node and asserts it's present.
func MustFindController(g symboldg.SymbolGraphBuilder, name string) (*symboldg.SymbolNode, metadata.ControllerMeta) {
	controllers := g.FindByKind(common.SymKindController)

	for _, controllerNode := range controllers {
		ctrl, ok := controllerNode.Data.(metadata.ControllerMeta)
		// Check the cast even for unrelated entities - a wrong value here means
		// something has gone horribly wrong.
		Expect(ok).To(BeTrue(), "A controller node had an unexpected Data type")
		if ctrl.Struct.Name == name {
			return controllerNode, ctrl
		}
	}

	Fail(fmt.Sprintf("Could not locate controller '%s'", name))
	return nil, metadata.ControllerMeta{} // Appease the compiler
}

func MustFindControllerReceiver(
	g symboldg.SymbolGraphBuilder,
	controllerNode *symboldg.SymbolNode,
	name string,
) (*symboldg.SymbolNode, metadata.ReceiverMeta) {
	recvEdge := MustFindOutgoingEdgeToName(
		g,
		controllerNode,
		[]symboldg.SymbolEdgeKind{symboldg.EdgeKindReceiver},
		name,
	)

	recvNode := g.Get(recvEdge.Edge.To)
	Expect(recvNode).ToNot(BeNil())

	recvMeta, isRecv := recvNode.Data.(*metadata.ReceiverMeta)
	Expect(isRecv).To(BeTrue())

	return recvNode, *recvMeta
}

// MustFindOutgoingEdgeToName finds an outgoing edge from `from` whose To.Name matches `toName`.
// kindFilter can be nil to match any kind.
func MustFindOutgoingEdgeToName(
	g symboldg.SymbolGraphBuilder,
	from *symboldg.SymbolNode,
	kindsFilter []symboldg.SymbolEdgeKind,
	toName string,
) symboldg.SymbolEdgeDescriptor {
	edges := common.MapValues(g.GetEdges(from.Id, kindsFilter))

	found := linq.First(edges, func(e symboldg.SymbolEdgeDescriptor) bool {
		return e.Edge.To.Name == toName
	})

	Expect(found).ToNot(BeNil(), "couldn't find outgoing edge to %s", toName)
	return *found
}

// CollectParamsAndRetVals returns Param and RetVal edges for a Receiver node.
//
// Note that this function will FAIL the test if anything other than Param or RetVal edges are encountered
func CollectAssertParamsAndRetVals(
	g symboldg.SymbolGraphBuilder,
	receiverNode *symboldg.SymbolNode,
) ([]*symboldg.SymbolNode, []*symboldg.SymbolNode) {
	paramNodes := []*symboldg.SymbolNode{}
	retValNodes := []*symboldg.SymbolNode{}

	edges := common.MapValues(g.GetEdges(
		receiverNode.Id,
		[]symboldg.SymbolEdgeKind{symboldg.EdgeKindParam, symboldg.EdgeKindRetVal},
	))

	for _, e := range edges {
		node := g.Get(e.Edge.To)
		Expect(node).ToNot(BeNil(), fmt.Sprintf("Could not obtain node with key '%v'", e.Edge.To))
		switch e.Edge.Kind {
		case symboldg.EdgeKindParam:
			paramNodes = append(paramNodes, node)
		case symboldg.EdgeKindRetVal:
			retValNodes = append(retValNodes, node)
		}
	}

	return paramNodes, retValNodes
}

func GetSingularChildNode(
	g symboldg.SymbolGraphBuilder,
	node *symboldg.SymbolNode,
	targetEdgeKind symboldg.SymbolEdgeKind,
) *symboldg.SymbolNode {
	relevantEdges := common.MapValues(g.GetEdges(node.Id, []symboldg.SymbolEdgeKind{targetEdgeKind}))
	Expect(relevantEdges).To(HaveLen(1), fmt.Sprintf("Node '%s' has more than one '%v' edges", targetEdgeKind, node.Id.Name))

	target := g.Get(relevantEdges[0].Edge.To)
	Expect(target).ToNot(BeNil())

	return target
}

func GetSingularChildTypeNode(g symboldg.SymbolGraphBuilder, node *symboldg.SymbolNode) *symboldg.SymbolNode {
	return GetSingularChildNode(g, node, symboldg.EdgeKindType)
}

// MustGetTypeNodeForEdge returns the node the edge points to (convenience).
func MustGetTypeNodeForEdge(g symboldg.SymbolGraphBuilder, edge symboldg.SymbolEdgeDescriptor) *symboldg.SymbolNode {
	node := g.Get(edge.Edge.To)
	Expect(node).ToNot(BeNil())
	return node
}

// MustStructMeta converts node.Data to StructMeta and asserts it.
func MustStructMeta(node *symboldg.SymbolNode) metadata.StructMeta {
	sm, ok := node.Data.(metadata.StructMeta)
	Expect(ok).To(BeTrue(), "expected node to contain StructMeta")
	return sm
}

// MustFieldMeta converts node.Data to FieldMeta and asserts it.
func MustFieldMeta(node *symboldg.SymbolNode) metadata.FieldMeta {
	fm, ok := node.Data.(metadata.FieldMeta)
	Expect(ok).To(BeTrue(), "expected node to contain FieldMeta")
	return fm
}

// MustAliasMeta converts node.Data to AliasMeta and asserts it.
func MustAliasMeta(node *symboldg.SymbolNode) metadata.AliasMeta {
	am, ok := node.Data.(metadata.AliasMeta)
	Expect(ok).To(BeTrue(), "expected node to contain AliasMeta")
	return am
}

// AssertFieldIsMap asserts a field exists with given name and that its type is a Map with key/value canonical strings.
func AssertFieldIsMap(structMeta metadata.StructMeta, fieldName, wantKey, wantValue string) {
	var f metadata.FieldMeta
	found := false
	for _, fld := range structMeta.Fields {
		if fld.Name == fieldName {
			f = fld
			found = true
			break
		}
	}
	Expect(found).To(BeTrue(), "field %s not found on struct", fieldName)
	root := f.Type.Root
	Expect(root.Kind()).To(Equal(metadata.TypeRefKindMap), "expected field %s to be a map type", fieldName)
	mapRef, ok := root.(*typeref.MapTypeRef)
	Expect(ok).To(BeTrue(), "map type assertion failed for field %s", fieldName)
	Expect(mapRef.Key.CanonicalString()).To(Equal(wantKey))
	Expect(mapRef.Value.CanonicalString()).To(Equal(wantValue))
}

func AssertGetField(
	g symboldg.SymbolGraphBuilder,
	structNode *symboldg.SymbolNode,
	fieldName string,
) (*symboldg.SymbolNode, metadata.FieldMeta) {
	Expect(structNode).ToNot(BeNil())
	MustStructMeta(structNode)

	edges := common.MapValues(g.GetEdges(structNode.Id, []symboldg.SymbolEdgeKind{symboldg.EdgeKindField}))

	relevantEdge := linq.First(edges, func(edge symboldg.SymbolEdgeDescriptor) bool {
		return edge.Edge.To.Name == fieldName
	})
	Expect(relevantEdge).ToNot(BeNil())

	fieldNode := g.Get(relevantEdge.Edge.To)
	Expect(fieldNode).ToNot(BeNil())
	fieldMeta := MustFieldMeta(fieldNode)

	return fieldNode, fieldMeta
}

func GetApiEndpointHierarchy(
	g symboldg.SymbolGraphBuilder,
	controllerName, receiverName string,
	paramNames []string,
) ApiEndpointInfo {
	controllerNode, controllerMeta := MustFindController(g, controllerName)
	receiverNode, receiverMeta := MustFindControllerReceiver(g, controllerNode, receiverName)
	paramNodes, retValNodes := CollectAssertParamsAndRetVals(g, receiverNode)

	var fParams []FuncParamInfo
	if len(paramNames) > 0 {
		fParams = linq.Map(paramNodes, func(pNode *symboldg.SymbolNode) FuncParamInfo {
			fMeta, isFMeta := pNode.Data.(metadata.FieldMeta)
			Expect(isFMeta).To(BeTrue())
			return FuncParamInfo{Node: pNode, Data: fMeta}
		})

		fParams = linq.Filter(fParams, func(fpi FuncParamInfo) bool {
			return slices.Contains(paramNames, fpi.Data.Name)
		})
	}

	// If we've missing parameters, length will differ and we fail
	Expect(fParams).To(HaveLen(len(paramNames)))

	fRetVals := linq.Map(retValNodes, func(pNode *symboldg.SymbolNode) FuncRetValInfo {
		fMeta, isFMeta := pNode.Data.(metadata.FieldMeta)
		Expect(isFMeta).To(BeTrue())
		return FuncRetValInfo{Node: pNode, Data: fMeta}
	})

	return ApiEndpointInfo{
		Controller: ControllerInfo{Node: controllerNode, Data: controllerMeta},
		Receiver:   ReceiverInfo{Node: receiverNode, Data: receiverMeta},
		Params:     fParams,
		RetVals:    fRetVals,
	}
}

func FollowThroughCompositeToTypeParams(
	g symboldg.SymbolGraphBuilder,
	fromNode *symboldg.SymbolNode,
) []*symboldg.SymbolNode {
	compositeNodes := g.Children(
		fromNode,
		&symboldg.TraversalBehavior{Filtering: symboldg.TraversalFilter{
			EdgeKinds: []symboldg.SymbolEdgeKind{symboldg.EdgeKindType},
		}},
	)
	Expect(compositeNodes).To(HaveLen(1))

	compNode := compositeNodes[0]

	_, isCompData := compNode.Data.(*metadata.CompositeMeta)
	Expect(isCompData).To(BeTrue())

	return g.Children(
		compositeNodes[0],
		&symboldg.TraversalBehavior{Filtering: symboldg.TraversalFilter{
			EdgeKinds: []symboldg.SymbolEdgeKind{symboldg.EdgeKindTypeParameter},
		}},
	)
}

func AssertFollowEdgesToNode(
	g symboldg.SymbolGraphBuilder,
	startNode *symboldg.SymbolNode,
	edgeTypeToFollow symboldg.SymbolEdgeKind,
	targetNodeFilter func(node *symboldg.SymbolNode) bool,
) *symboldg.SymbolNode {
	edges := g.GetEdges(startNode.Id, []symboldg.SymbolEdgeKind{edgeTypeToFollow})
	for _, edgeDesc := range common.MapValues(edges) {
		node := g.Get(edgeDesc.Edge.To)
		Expect(node).ToNot(BeNil())

		if targetNodeFilter(node) {
			return node
		}
	}

	Fail("AssertFollowEdgesToNode concluded with no result")
	return nil
}
