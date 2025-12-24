package metadata_test

import (
	"go/ast"

	"github.com/gopher-fleece/gleece/common"
	"github.com/gopher-fleece/gleece/core/metadata"
	"github.com/gopher-fleece/gleece/core/metadata/typeref"
	"github.com/gopher-fleece/gleece/test/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Partial tests for now - a lot of this is covered elsewhere
var _ = Describe("Unit Tests - TypeUsage", func() {
	Context("Resolve", func() {
		When("Usage is a 'Universe' type", func() {

			It("Returns correct metadata for non-pointers", func() {
				usage := metadata.TypeUsageMeta{
					SymNodeMeta: metadata.SymNodeMeta{
						Name:       "string",
						SymbolKind: common.SymKindBuiltin,
					},
					Root: utils.MakeUniverseRoot("string"),
				}

				result, err := usage.Reduce(metadata.ReductionContext{})
				Expect(err).To(BeNil())
				Expect(result.Name).To(Equal("string"))
				Expect(result.Import).To(Equal(common.ImportTypeNone))
				Expect(result.IsUniverseType).To(BeTrue())
				Expect(result.IsByAddress).To(BeFalse())
				Expect(result.SymbolKind).To(Equal(common.SymKindBuiltin))
				Expect(result.AliasMetadata).To(BeNil())
			})

			It("Returns correct metadata for pointers", func() {
				usage := metadata.TypeUsageMeta{
					SymNodeMeta: metadata.SymNodeMeta{
						Name:       "string",
						SymbolKind: common.SymKindBuiltin,
						Node: common.Ptr(ast.StarExpr{
							Star: 15,
							X:    ast.NewIdent("string"),
						}),
					},
					Root: &typeref.PtrTypeRef{
						Elem: utils.MakeUniverseRoot("string"),
					},
				}

				result, err := usage.Reduce(metadata.ReductionContext{})
				Expect(err).To(BeNil())
				Expect(result.Name).To(Equal("string"))
				Expect(result.Import).To(Equal(common.ImportTypeNone))
				Expect(result.IsUniverseType).To(BeTrue())
				Expect(result.IsByAddress).To(BeTrue())
				Expect(result.SymbolKind).To(Equal(common.SymKindBuiltin))
				Expect(result.AliasMetadata).To(BeNil())
			})
		})
	})
})
