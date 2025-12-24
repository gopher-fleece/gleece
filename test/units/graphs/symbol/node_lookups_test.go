package symbol_test

import (
	"github.com/gopher-fleece/gleece/common"
	"github.com/gopher-fleece/gleece/graphs/symboldg"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Unit Tests - SymbolGraph", func() {
	var graph symboldg.SymbolGraph

	BeforeEach(func() {
		graph = symboldg.NewSymbolGraph()
	})

	Context("FindByKind", func() {
		It("Correctly finds elements by the symbol kind", func() {
			// Add a couple of relevant nodes
			anyNode := graph.AddSpecial(symboldg.SpecialTypeAny)
			errNode := graph.AddSpecial(symboldg.SpecialTypeError)

			// Add a couple unrelated nodes that should be ignored

			graph.AddPrimitive(symboldg.PrimitiveTypeBool)
			graph.AddPrimitive(symboldg.PrimitiveTypeString)

			results := graph.FindByKind(common.SymKindSpecialBuiltin)

			Expect(results).To(HaveLen(2))
			Expect(results).To(ContainElements(anyNode, errNode))
		})
	})

	Context("IsPrimitivePresent", func() {
		It("Recognizes that a previously added primitive is present", func() {
			graph := symboldg.NewSymbolGraph()
			graph.AddPrimitive(symboldg.PrimitiveTypeBool)
			Expect(graph.IsPrimitivePresent(symboldg.PrimitiveTypeBool)).To(BeTrue())
		})

		It("Returns false for primitives that have not been added", func() {
			graph := symboldg.NewSymbolGraph()
			Expect(graph.IsPrimitivePresent(symboldg.PrimitiveTypeInt)).To(BeFalse())
		})
	})

	Context("IsSpecialPresent", func() {
		It("Recognizes a previously added special type", func() {
			graph := symboldg.NewSymbolGraph()
			graph.AddSpecial(symboldg.SpecialTypeError)
			Expect(graph.IsSpecialPresent(symboldg.SpecialTypeError)).To(BeTrue())
		})

		It("Returns false for special types not added", func() {
			graph := symboldg.NewSymbolGraph()
			Expect(graph.IsSpecialPresent(symboldg.SpecialTypeTime)).To(BeFalse())
		})
	})
})
