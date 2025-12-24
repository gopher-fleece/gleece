package metadata_test

import (
	"fmt"

	"github.com/gopher-fleece/gleece/v2/core/metadata"
	"github.com/gopher-fleece/gleece/v2/core/metadata/typeref"
	"github.com/gopher-fleece/gleece/v2/gast"
	"github.com/gopher-fleece/gleece/v2/graphs"
	"github.com/gopher-fleece/gleece/v2/test/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Unit Tests - FuncTypeRef", func() {
	var fVersion *gast.FileVersion

	BeforeEach(func() {
		fVersion = utils.MakeFileVersion("file", "")
	})

	Context("Kind", func() {
		It("Returns Func kind", func() {
			funcRef := &typeref.FuncTypeRef{}
			Expect(funcRef.Kind()).To(Equal(metadata.TypeRefKindFunc))
		})
	})

	Context("string representations", func() {
		It("Produces canonical string using params/results canonical strings", func() {
			paramRef := &utils.FakeTypeRef{
				RefKind:      metadata.TypeRefKindNamed,
				CanonicalStr: "pkg.A",
				SimpleStr:    "A",
			}
			resultRef := &utils.FakeTypeRef{
				RefKind:      metadata.TypeRefKindNamed,
				CanonicalStr: "pkg.B",
				SimpleStr:    "B",
			}

			funcRef := &typeref.FuncTypeRef{
				Params:  []metadata.TypeRef{paramRef},
				Results: []metadata.TypeRef{resultRef},
			}

			Expect(funcRef.CanonicalString()).To(Equal("func(pkg.A)(pkg.B)"))
			Expect(funcRef.SimpleTypeString()).To(Equal("func(A)(B)"))
		})

		It("Handles multiple params/results and empty lists", func() {
			p1 := &utils.FakeTypeRef{CanonicalStr: "X1", SimpleStr: "X1"}
			p2 := &utils.FakeTypeRef{CanonicalStr: "X2", SimpleStr: "X2"}
			r1 := &utils.FakeTypeRef{CanonicalStr: "R1", SimpleStr: "R1"}

			funcRef := &typeref.FuncTypeRef{
				Params:  []metadata.TypeRef{p1, p2},
				Results: []metadata.TypeRef{r1},
			}

			Expect(funcRef.CanonicalString()).To(Equal("func(X1,X2)(R1)"))
			Expect(funcRef.SimpleTypeString()).To(Equal("func(X1,X2)(R1)"))
		})
	})

	Context("ToSymKey / CacheLookupKey", func() {
		It("Returns composite symkey combining param and result symkeys", func() {
			paramSymKey := graphs.NewUniverseSymbolKey("int")
			resultSymKey := graphs.NewUniverseSymbolKey("string")

			paramRef := &utils.FakeTypeRef{SymKey: paramSymKey}
			resultRef := &utils.FakeTypeRef{SymKey: resultSymKey}

			funcRef := &typeref.FuncTypeRef{
				Params:  []metadata.TypeRef{paramRef},
				Results: []metadata.TypeRef{resultRef},
			}

			expectedKey := graphs.NewCompositeTypeKey(graphs.CompositeKindFunc, fVersion, []graphs.SymbolKey{
				paramSymKey, resultSymKey,
			})

			gotKey, err := funcRef.ToSymKey(fVersion)
			Expect(err).ToNot(HaveOccurred())
			Expect(gotKey).To(Equal(expectedKey))

			// CacheLookupKey delegates to ToSymKey
			cacheKey, err := funcRef.CacheLookupKey(fVersion)
			Expect(err).ToNot(HaveOccurred())
			Expect(cacheKey).To(Equal(expectedKey))
		})

		It("Propagates error when a param's ToSymKey fails", func() {
			paramErr := fmt.Errorf("param failure")
			badParam := &utils.FakeTypeRef{ToSymKeyErr: paramErr}

			funcRef := &typeref.FuncTypeRef{
				Params:  []metadata.TypeRef{badParam},
				Results: nil,
			}

			_, err := funcRef.ToSymKey(fVersion)
			Expect(err).To(MatchError(ContainSubstring("param failure")))
		})
	})

	Context("Flatten", func() {
		It("Returns flattened list of params then results", func() {
			paramRef := &utils.FakeTypeRef{FlattenResponse: []metadata.TypeRef{}}
			resultRef := &utils.FakeTypeRef{FlattenResponse: []metadata.TypeRef{}}

			// We expect flatten to return concatenation of param.Flatten() and result.Flatten().
			// Provide non-empty marker slices on each FakeTypeRef.
			paramRef.FlattenResponse = []metadata.TypeRef{paramRef}
			resultRef.FlattenResponse = []metadata.TypeRef{resultRef}

			funcRef := &typeref.FuncTypeRef{
				Params:  []metadata.TypeRef{paramRef},
				Results: []metadata.TypeRef{resultRef},
			}

			got := funcRef.Flatten()
			Expect(got).To(HaveLen(2))
			Expect(got[0]).To(BeIdenticalTo(paramRef))
			Expect(got[1]).To(BeIdenticalTo(resultRef))
		})
	})
})
