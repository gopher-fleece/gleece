package symbol_test

import (
	"github.com/gopher-fleece/gleece/v2/common"
	"github.com/gopher-fleece/gleece/v2/graphs/symboldg"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Unit Tests - SymbolGraph", func() {
	var graph symboldg.SymbolGraph

	BeforeEach(func() {
		graph = symboldg.NewSymbolGraph()
	})

	Context("FindByKind", func() {

		It("Returns an empty slice when nothing was found", func() {
			results := graph.FindByKind(common.SymKindSpecialBuiltin)
			Expect(results).To(HaveLen(0))
		})

		It("Returns an slice with one node when only one node matches", func() {
			// A relevant node
			anyNode := graph.AddSpecial(common.SpecialTypeAny)

			// Irrelevant nodes
			graph.AddPrimitive(common.PrimitiveTypeBool)
			graph.AddPrimitive(common.PrimitiveTypeString)

			results := graph.FindByKind(common.SymKindSpecialBuiltin)

			Expect(results).To(HaveLen(1))
			Expect(results).To(ContainElements(anyNode))
		})

		It("Correctly finds elements by the symbol kind", func() {
			// Add a couple of relevant nodes
			anyNode := graph.AddSpecial(common.SpecialTypeAny)
			errNode := graph.AddSpecial(common.SpecialTypeError)

			// Add a couple unrelated nodes that should be ignored
			graph.AddPrimitive(common.PrimitiveTypeBool)
			graph.AddPrimitive(common.PrimitiveTypeString)

			results := graph.FindByKind(common.SymKindSpecialBuiltin)

			Expect(results).To(HaveLen(2))
			Expect(results).To(ContainElements(anyNode, errNode))
		})

	})

	Context("IsPrimitivePresent", func() {
		It("Recognizes that a previously added primitive is present", func() {
			graph := symboldg.NewSymbolGraph()
			graph.AddPrimitive(common.PrimitiveTypeBool)
			Expect(graph.IsPrimitivePresent(common.PrimitiveTypeBool)).To(BeTrue())
		})

		It("Returns false for primitives that have not been added", func() {
			graph := symboldg.NewSymbolGraph()
			Expect(graph.IsPrimitivePresent(common.PrimitiveTypeInt)).To(BeFalse())
		})
	})

	Context("IsSpecialPresent", func() {
		It("Recognizes a previously added special type", func() {
			graph := symboldg.NewSymbolGraph()
			graph.AddSpecial(common.SpecialTypeError)
			Expect(graph.IsSpecialPresent(common.SpecialTypeError)).To(BeTrue())
		})

		It("Returns false for special types not added", func() {
			graph := symboldg.NewSymbolGraph()
			Expect(graph.IsSpecialPresent(common.SpecialTypeTime)).To(BeFalse())
		})
	})
})
