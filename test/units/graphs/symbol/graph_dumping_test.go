package symbol_test

import (
	"github.com/gopher-fleece/gleece/common"
	"github.com/gopher-fleece/gleece/core/metadata"
	"github.com/gopher-fleece/gleece/gast"
	"github.com/gopher-fleece/gleece/graphs"
	"github.com/gopher-fleece/gleece/graphs/symboldg"
	"github.com/gopher-fleece/gleece/test/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Unit Tests - SymbolGraph", func() {
	var graph symboldg.SymbolGraph
	var fVersion *gast.FileVersion

	BeforeEach(func() {
		graph = symboldg.NewSymbolGraph()
		fVersion = utils.MakeFileVersion("file", "")
	})

	Context("String", func() {
		It("Outputs an empty graph with 'headers' when empty", func() {
			text := graph.String()
			Expect(text).To(Equal("=== SymbolGraph Dump ===\nTotal nodes: 0\n\n=== End SymbolGraph ===\n"))
		})

		It("Outputs a correct graph when nodes exist but have no dependencies", func() {
			_, err := graph.AddConst(symboldg.CreateConstNode{
				Data: metadata.ConstMeta{
					SymNodeMeta: metadata.SymNodeMeta{
						Name:     "SomeConst",
						Node:     utils.MakeIdent("SomeConst"),
						FVersion: fVersion,
					},
					Value: "Some Value",
					Type: metadata.TypeUsageMeta{
						SymNodeMeta: metadata.SymNodeMeta{Name: "string", FVersion: fVersion},
					},
				},
			})
			Expect(err).To(BeNil())

			text := graph.String()
			expectedPattern := `^=== SymbolGraph Dump ===\n` +
				`Total nodes: 1\n\n` +
				`\[Constant\] SomeConst\n` +
				`    • file\n` +
				`    • \d+\n` + // matches timestamp
				`    • hash-file-\n\n` +
				`  Dependencies: \(none\)\n` +
				`  Dependents: \(none\)\n\n` +
				`=== End SymbolGraph ===\n$`

			Expect(text).To(MatchRegexp(expectedPattern))
		})

		It("Outputs a correct graph when nodes exist and have a dependent node", func() {
			constNode, err := graph.AddConst(symboldg.CreateConstNode{
				Data: metadata.ConstMeta{
					SymNodeMeta: metadata.SymNodeMeta{
						Name:     "SomeConst",
						Node:     utils.MakeIdent("SomeConst"),
						FVersion: fVersion,
					},
					Value: "Some Value",
					Type: metadata.TypeUsageMeta{
						SymNodeMeta: metadata.SymNodeMeta{Name: "string", FVersion: fVersion},
					},
				},
			})
			Expect(err).To(BeNil())

			strNode := graph.AddPrimitive(common.PrimitiveTypeString)
			graph.AddEdge(constNode.Id, strNode.Id, symboldg.EdgeKindType, nil)

			text := graph.String()

			// Graph string output does not guarantee ordering so we have to test separately here
			const nodeBlockConstant = `(?s)\[Constant\] SomeConst\n` +
				`    • file\n` +
				`    • \d+\n` +
				`    • hash-file-\n\n` +
				`  Dependencies:\n` +
				`    • \[Builtin\] string \(ty\)\n` +
				`  Dependents: \(none\)`

			const nodeBlockBuiltin = `(?s)\[Builtin\] string\n` +
				`  Dependencies: \(none\)\n` +
				`  Dependents:\n` +
				`    • \[Constant\] SomeConst\n` +
				`    • file\n` +
				`    • \d+\n` +
				`    • hash-file-`

			// Assert that both node blocks exist somewhere in the dump
			Expect(text).To(MatchRegexp(nodeBlockConstant))
			Expect(text).To(MatchRegexp(nodeBlockBuiltin))

		})

		It("Outputs a correct graph when nodes exist and have a dependent edge without node", func() {
			constNode, err := graph.AddConst(symboldg.CreateConstNode{
				Data: metadata.ConstMeta{
					SymNodeMeta: metadata.SymNodeMeta{
						Name:     "SomeConst",
						Node:     utils.MakeIdent("SomeConst"),
						FVersion: fVersion,
					},
					Value: "Some Value",
					Type: metadata.TypeUsageMeta{
						SymNodeMeta: metadata.SymNodeMeta{Name: "string", FVersion: fVersion},
					},
				},
			})
			Expect(err).To(BeNil())

			// Add an edge without actually adding a node
			graph.AddEdge(constNode.Id, graphs.NewUniverseSymbolKey("string"), symboldg.EdgeKindType, nil)

			text := graph.String()

			const expectedRx = `(?m)^=== SymbolGraph Dump ===\n` +
				`Total nodes: 1\n\n` +
				`\[Constant\] SomeConst\n` +
				`    • file\n` +
				`    • \d+\n` +
				`    • hash-file-\n\n` +
				`  Dependencies:\n` +
				`    • \[string\] \(ty\)\n` +
				`  Dependents: \(none\)\n\n` +
				`=== End SymbolGraph ===$`

			Expect(text).To(MatchRegexp(expectedRx))
		})
	})

	Context("ToDot", func() {
		It("Outputs correct empty graph with default style when empty", func() {
			text := graph.ToDot(nil)

			// Graph renders legend only if not empty
			Expect(text).To(Equal("digraph SymbolGraph {\n  rankdir=TB;\n}\n"))
		})

		It("Outputs nodes and their edges", func() {
			anyNode := graph.AddSpecial(common.SpecialTypeAny)
			strNode := graph.AddPrimitive(common.PrimitiveTypeString)
			graph.AddEdge(anyNode.Id, strNode.Id, symboldg.EdgeKindType, nil)

			text := graph.ToDot(nil)

			Expect(text).To(MatchRegexp(`N\d \[label=\"any@\.| \(Special\)\"`))
			Expect(text).To(MatchRegexp(`N\d \[label=\"string@\.| \(Builtin\)\"`))
			Expect(text).To(MatchRegexp(`N\d -> N\d \[label=\"Type\"`))
		})
	})
})
