package graph_test

import (
	"testing"

	"github.com/gopher-fleece/gleece/common"
	"github.com/gopher-fleece/gleece/graphs/symboldg"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
	"github.com/gopher-fleece/gleece/test/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var orderTraversalBehavior = common.Ptr(symboldg.TraversalBehavior{Sorting: symboldg.TraversalSortingOrdinalAsc})

var _ = Describe("Graph Controller", func() {
	It("Creates a valid Symbol Graph", func() {
		pipe := utils.GetPipelineOrFail()
		err := pipe.GenerateGraph()
		Expect(err).To(BeNil())

		controllers := pipe.Graph().FindByKind(common.SymKindController)
		Expect(controllers).To(HaveLen(1))

		// Notice we're passing a traversal behavior with sorting here, otherwise output may be out-of-order
		receivers := pipe.Graph().Children(controllers[0], orderTraversalBehavior)
		Expect(receivers).To(HaveLen(2))

		receiver1Children := pipe.Graph().Children(receivers[0], orderTraversalBehavior)
		Expect(receiver1Children).To(HaveLen(4))

		// First parameter - routeParam
		Expect(receiver1Children[0].Id.Name).To(Equal("routeParam"))
		Expect(receiver1Children[0].Id.IsUniverse).To(BeFalse())
		Expect(receiver1Children[0].Id.IsBuiltIn).To(BeFalse())
		Expect(receiver1Children[0].Kind).To(Equal(common.SymKindField))

		routeParamChildren := pipe.Graph().Children(receiver1Children[0], orderTraversalBehavior)
		Expect(routeParamChildren).To(HaveLen(1))
		Expect(routeParamChildren[0].Id.Name).To(Equal("string"))
		Expect(routeParamChildren[0].Id.IsUniverse).To(BeTrue())
		Expect(routeParamChildren[0].Id.IsBuiltIn).To(BeTrue())
		Expect(routeParamChildren[0].Kind).To(Equal(common.SymKindBuiltin))

		// Second parameter - queryParam
		Expect(receiver1Children[1].Id.Name).To(Equal("queryParam"))
		Expect(receiver1Children[1].Id.IsUniverse).To(BeFalse())
		Expect(receiver1Children[1].Id.IsBuiltIn).To(BeFalse())
		Expect(receiver1Children[1].Kind).To(Equal(common.SymKindField))

		queryParamChildren := pipe.Graph().Children(receiver1Children[1], orderTraversalBehavior)
		Expect(queryParamChildren).To(HaveLen(1))
		Expect(queryParamChildren[0].Id.Name).To(Equal("int"))
		Expect(queryParamChildren[0].Id.IsUniverse).To(BeTrue())
		Expect(queryParamChildren[0].Id.IsBuiltIn).To(BeTrue())
		Expect(queryParamChildren[0].Kind).To(Equal(common.SymKindBuiltin))

		// Third parameter - headerParam
		Expect(receiver1Children[2].Id.Name).To(Equal("headerParam"))
		Expect(receiver1Children[2].Id.IsUniverse).To(BeFalse())
		Expect(receiver1Children[2].Id.IsBuiltIn).To(BeFalse())
		Expect(receiver1Children[2].Kind).To(Equal(common.SymKindField))

		headerParamChildren := pipe.Graph().Children(receiver1Children[2], orderTraversalBehavior)
		Expect(headerParamChildren).To(HaveLen(1))
		Expect(headerParamChildren[0].Id.Name).To(Equal("float32"))
		Expect(headerParamChildren[0].Id.IsUniverse).To(BeTrue())
		Expect(headerParamChildren[0].Id.IsBuiltIn).To(BeTrue())
		Expect(headerParamChildren[0].Kind).To(Equal(common.SymKindBuiltin))

		// Single 'error' return value
		Expect(receiver1Children[3].Id.Name).To(Equal(""))
		Expect(receiver1Children[3].Id.IsUniverse).To(BeFalse())
		Expect(receiver1Children[3].Id.IsBuiltIn).To(BeFalse())
		Expect(receiver1Children[3].Kind).To(Equal(common.SymKindField))

		retValChildren := pipe.Graph().Children(receiver1Children[3], orderTraversalBehavior)
		Expect(retValChildren).To(HaveLen(1))
		Expect(retValChildren[0].Id.Name).To(Equal("error"))
		Expect(retValChildren[0].Id.IsUniverse).To(BeTrue())
		Expect(retValChildren[0].Id.IsBuiltIn).To(BeTrue())
		Expect(retValChildren[0].Kind).To(Equal(common.SymKindSpecialBuiltin))
	})
})

func TestGraphController(t *testing.T) {
	logger.SetLogLevel(logger.LogLevelNone)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Graph Controller")
}
