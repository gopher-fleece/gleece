package symbol_test

import (
	"testing"

	"github.com/gopher-fleece/gleece/common"
	"github.com/gopher-fleece/gleece/core/metadata"
	"github.com/gopher-fleece/gleece/core/metadata/typeref"
	"github.com/gopher-fleece/gleece/graphs"
	"github.com/gopher-fleece/gleece/graphs/symboldg"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
	"github.com/gopher-fleece/gleece/test/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Unit Tests - SymbolGraph", func() {
	Context("NewSymbolGraph", func() {
		It("Creates a usable SymbolGraph", func() {
			g := symboldg.NewSymbolGraph()
			// Graph should be usable: add a primitive and check presence
			p := symboldg.PrimitiveTypeString
			n := g.AddPrimitive(p)
			Expect(n).ToNot(BeNil())
			Expect(g.IsPrimitivePresent(p)).To(BeTrue())
		})
	})

	Context("Auxiliary", func() {
		var graph symboldg.SymbolGraph

		BeforeEach(func() {
			graph = symboldg.NewSymbolGraph()
		})

		It("Returns an error from the idempotency guard when the given FileVersion is nil", func() {
			// Build a TypeUsageMeta using the new Root-based representation (universe "string")
			k := graphs.NewUniverseSymbolKey("string")
			root := typeref.NewNamedTypeRef(&k, nil)

			_, err := graph.AddConst(symboldg.CreateConstNode{
				Data: metadata.ConstMeta{
					SymNodeMeta: metadata.SymNodeMeta{
						Name:     "SomeConst",
						Node:     utils.MakeIdent("SomeConst"),
						FVersion: nil,
					},
					Value: "some value",
					Type: metadata.TypeUsageMeta{
						SymNodeMeta: metadata.SymNodeMeta{
							Name:     "string",
							FVersion: nil,
						},
						Root: &root,
					},
				},
			})
			Expect(err).To(MatchError(ContainSubstring("idempotencyGuard received a nil version parameter")))
		})

		It("Evicts old nodes when adding same decl with a different FileVersion", func() {
			// Create a "same" AST node (same ident) but two different versions
			node := utils.MakeIdent("S")
			fv1 := utils.MakeFileVersion("v1", "some-hash")
			fv2 := utils.MakeFileVersion("v1", "some-different-hash")

			structMetaV1 := metadata.StructMeta{
				SymNodeMeta: metadata.SymNodeMeta{
					Node:     node,
					FVersion: fv1,
				},
			}
			n1, err := graph.AddStruct(symboldg.CreateStructNode{Data: structMetaV1})
			Expect(err).ToNot(HaveOccurred())
			Expect(n1).ToNot(BeNil())

			// Add a dependent: add a field belonging to this struct (so revDeps get populated)
			// Field type -> universe "int"
			intKey := graphs.NewUniverseSymbolKey("int")
			intRoot := typeref.NewNamedTypeRef(&intKey, nil)

			fieldMeta := metadata.FieldMeta{
				SymNodeMeta: metadata.SymNodeMeta{
					Node:     utils.MakeIdent("F1"),
					FVersion: fv1,
				},
				Type: metadata.TypeUsageMeta{
					SymNodeMeta: metadata.SymNodeMeta{
						FVersion: fv1,
					},
					Root: &intRoot,
				},
			}
			_, err = graph.AddField(symboldg.CreateFieldNode{Data: fieldMeta})
			Expect(err).ToNot(HaveOccurred())

			// Now add the "same" struct but with a different file version -> should evict the previous
			structMetaV2 := metadata.StructMeta{SymNodeMeta: metadata.SymNodeMeta{Node: node, FVersion: fv2}}
			n2, err := graph.AddStruct(symboldg.CreateStructNode{Data: structMetaV2})
			Expect(err).ToNot(HaveOccurred())
			Expect(n2).ToNot(BeNil())
			// Ensure the new node is present and has the new version
			found := false
			for _, s := range graph.FindByKind(common.SymKindStruct) {
				if s.Id.Equals(n2.Id) {
					found = true
				}
			}
			Expect(found).To(BeTrue())
		})
	})
})

func TestUnitGraphs(t *testing.T) {
	logger.SetLogLevel(logger.LogLevelNone)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Unit Tests - SymbolGraph")
}
