package dot_test

import (
	"strings"
	"testing"

	"github.com/gopher-fleece/gleece/common"
	"github.com/gopher-fleece/gleece/graphs"
	"github.com/gopher-fleece/gleece/graphs/dot"
	"github.com/gopher-fleece/gleece/graphs/symboldg"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Unit Tests - Dot", func() {

	Describe("DotBuilder", func() {
		var builder *dot.DotBuilder

		BeforeEach(func() {
			builder = dot.NewDotBuilder(nil) // use default theme
		})

		It("Initializes with graph header and default rankdir", func() {
			output := builder.Finish()
			Expect(output).To(ContainSubstring("digraph SymbolGraph {"))
			Expect(output).To(ContainSubstring("rankdir=TB")) // Default here is TB (top-to-bottom)
		})

		Describe("NewDotBuilder with theme", func() {
			It("Uses provided theme if non-nil", func() {
				customTheme := &dot.DotTheme{
					Direction: dot.RankDirLR,
				}
				db := dot.NewDotBuilder(customTheme)
				Expect(db).NotTo(BeNil())
				Expect(strings.Contains(db.Finish(), "rankdir=LR")).To(BeTrue())
			})

			It("Uses default theme if nil provided", func() {
				db := dot.NewDotBuilder(nil)
				Expect(db).NotTo(BeNil())
				Expect(strings.Contains(db.Finish(), "rankdir=TB")).To(BeTrue())
			})
		})

		Describe("AddNode", func() {
			It("Adds a node with proper label and style", func() {
				sk := graphs.SymbolKey{Name: "MyNode"}
				builder.AddNode(sk, common.SymKindStruct, "NodeLabel")

				out := builder.Finish()
				Expect(out).To(ContainSubstring("N0 [label=\"NodeLabel (Struct)\""))
				Expect(out).To(ContainSubstring("shape="))     // shape set
				Expect(out).To(ContainSubstring("fillcolor=")) // fill color set
			})

			It("Uses style from theme for known kind", func() {
				db := dot.NewDotBuilder(nil)
				key := graphs.SymbolKey{Name: "node1"}
				db.AddNode(key, common.SymKindStruct, "Label")
				output := db.Finish()
				Expect(output).To(ContainSubstring("label=\"Label (Struct)\""))
				Expect(output).To(ContainSubstring("shape=box")) // from DefaultDotTheme for SymKindStruct
			})

			It("Falls back to default style if kind unknown", func() {
				db := dot.NewDotBuilder(nil)
				key := graphs.SymbolKey{Name: "node2"}
				unknownKind := common.SymKind("unknown_kind")
				db.AddNode(key, unknownKind, "Label2")
				output := db.Finish()
				Expect(output).To(ContainSubstring("label=\"Label2 (unknown_kind)\""))
				Expect(output).To(ContainSubstring("shape=ellipse"))
				Expect(output).To(ContainSubstring("fillcolor=\"gray90\""))
			})

			It("Fills empty style fields with defaults", func() {
				theme := dot.DotTheme{
					NodeStyles: map[common.SymKind]dot.DotStyle{
						common.SymKindStruct: {Color: "", Shape: "", FontColor: ""},
					},
				}
				db := dot.NewDotBuilder(&theme)
				key := graphs.SymbolKey{Name: "node3"}
				db.AddNode(key, common.SymKindStruct, "Label3")
				output := db.Finish()
				Expect(output).To(ContainSubstring("shape=ellipse"))
				Expect(output).To(ContainSubstring("fillcolor=\"gray90\""))
				Expect(output).To(ContainSubstring("fontcolor=\"black\""))
			})
		})

		Describe("AddEdge", func() {
			It("Adds an edge between two existing nodes", func() {
				from := graphs.SymbolKey{Name: "FromNode"}
				to := graphs.SymbolKey{Name: "ToNode"}

				builder.AddNode(from, common.SymKindStruct, "From")
				builder.AddNode(to, common.SymKindInterface, "To")

				builder.AddEdge(from, to, "calls", nil)

				out := builder.Finish()
				Expect(out).To(ContainSubstring("N0 -> N1"))
				Expect(out).To(ContainSubstring("[label=\"calls\""))
			})

			It("Adds edge with style and label from theme", func() {
				theme := dot.DefaultDotTheme
				db := dot.NewDotBuilder(&theme)
				from := graphs.SymbolKey{Name: "from"}
				to := graphs.SymbolKey{Name: "to"}
				db.AddNode(from, common.SymKindStruct, "From")
				db.AddNode(to, common.SymKindStruct, "To")

				db.AddEdge(from, to, "ty", nil) // edge label and style present in theme
				output := db.Finish()
				Expect(output).To(ContainSubstring("label=\"Type\""))
				Expect(output).To(ContainSubstring("color=\"black\""))
				Expect(output).To(ContainSubstring("style=\"solid\""))
				Expect(output).To(ContainSubstring("arrowhead=\"vee\""))
			})

			It("Adds edge with default style when kind unknown", func() {
				db := dot.NewDotBuilder(nil)
				from := graphs.SymbolKey{Name: "from2"}
				db.AddNode(from, common.SymKindStruct, "From2")

				unknownTo := graphs.SymbolKey{Name: "unknown_to"}
				db.AddEdge(from, unknownTo, "unknown_kind", nil)
				output := db.Finish()

				// Should add error node and one edge to error node
				Expect(strings.Count(output, "N_ERROR")).To(Equal(2))
				Expect(strings.Count(output, "-> N_ERROR")).To(Equal(1))
			})

			It("Appends a given edge suffix string if provided", func() {
				from := graphs.SymbolKey{Name: "FromNode"}
				to := graphs.SymbolKey{Name: "ToNode"}

				builder.AddNode(from, common.SymKindStruct, "From")
				builder.AddNode(to, common.SymKindInterface, "To")

				builder.AddEdge(from, to, string(symboldg.EdgeKindType), common.Ptr(" (Suffix Test)"))

				out := builder.Finish()

				Expect(out).To(ContainSubstring("N0 -> N1 [label=\"Type (Suffix Test)\""))
			})
		})

		Describe("addErrorNodeOnce", func() {
			It("Adds error node with default styles when fields empty", func() {
				emptyStyle := dot.DotStyle{}
				theme := &dot.DotTheme{
					Direction:      dot.RankDirTB,
					ErrorNodeStyle: emptyStyle,
				}
				db := dot.NewDotBuilder(theme)

				// trigger addErrorNodeOnce by adding edge to unknown node
				from := graphs.SymbolKey{Name: "from"}
				to := graphs.SymbolKey{Name: "to_unknown"}

				db.AddNode(from, common.SymKindStruct, "From")
				db.AddEdge(from, to, "edgekind", nil)

				output := db.Finish()
				Expect(output).To(ContainSubstring("fillcolor=\"red\""))
				Expect(output).To(ContainSubstring("shape=octagon"))
				Expect(output).To(ContainSubstring("fontcolor=\"white\""))
			})

			It("Does not add error node twice", func() {
				db := dot.NewDotBuilder(nil)

				from := graphs.SymbolKey{Name: "from"}
				to := graphs.SymbolKey{Name: "unknown1"}

				// We're testing wether the N_ERROR node is being re-used across different edges.
				// If it's re-used, we expect it to occur (1 + number of error edges) in the output.
				// If it's not, it'll appear twice for each edge.

				db.AddNode(from, common.SymKindStruct, "From")
				db.AddEdge(from, to, "edgekind", nil)
				before := db.Finish()

				// Add another edge to unknown node, should not duplicate error node
				to2 := graphs.SymbolKey{Name: "unknown2"}
				db.AddEdge(from, to2, "edgekind", nil)
				after := db.Finish()

				Expect(strings.Count(before, "N_ERROR")).To(Equal(2))
				Expect(strings.Count(after, "N_ERROR")).To(Equal(3))
			})
		})

		Describe("RenderLegend", func() {
			It("Does not render legend for an empty graph", func() {
				builder.RenderLegend()

				out := builder.Finish()
				Expect(out).ToNot(ContainSubstring("subgraph cluster_legend"))
				Expect(out).ToNot(ContainSubstring("label = \"Legend\""))
			})

			It("Renders legend with expected content", func() {
				builder.AddNode(graphs.NewUniverseSymbolKey("string"), common.SymKindBuiltin, "Test")
				builder.RenderLegend()

				out := builder.Finish()
				Expect(out).To(ContainSubstring("subgraph cluster_legend"))
				Expect(out).To(ContainSubstring("label = \"Legend\""))
				// Spot check for at least one node kind from the theme
				Expect(out).To(ContainSubstring("style=filled"))
			})

			It("Renders legend with default fallback color and shape", func() {
				theme := dot.DotTheme{
					NodeStyles: map[common.SymKind]dot.DotStyle{
						common.SymKindBuiltin: {Color: "", Shape: ""},
					},
				}
				db := dot.NewDotBuilder(&theme)
				// Add a node to trigger legend render
				db.AddNode(graphs.NewUniverseSymbolKey("string"), common.SymKindBuiltin, "Test")
				db.RenderLegend()
				output := db.Finish()

				Expect(output).To(ContainSubstring("fillcolor=\"gray90\""))
				Expect(output).To(ContainSubstring("shape=ellipse"))
			})
		})
	})
})

func TestUnitsGraphsDot(t *testing.T) {
	logger.SetLogLevel(logger.LogLevelNone)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Unit Tests - Dot")
}
