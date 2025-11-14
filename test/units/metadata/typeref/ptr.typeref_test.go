package metadata_test

import (
	"errors"

	"github.com/gopher-fleece/gleece/core/metadata"
	"github.com/gopher-fleece/gleece/core/metadata/typeref"
	"github.com/gopher-fleece/gleece/gast"
	"github.com/gopher-fleece/gleece/graphs"
	"github.com/gopher-fleece/gleece/test/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Unit Tests - PtrTypeRef", func() {
	var fVersion *gast.FileVersion

	BeforeEach(func() {
		fVersion = utils.MakeFileVersion("file", "")
	})

	Context("Kind", func() {
		It("Returns pointer kind", func() {
			ptrRef := &typeref.PtrTypeRef{Elem: &utils.FakeTypeRef{}}
			Expect(ptrRef.Kind()).To(Equal(metadata.TypeRefKindPtr))
		})
	})

	Context("String representations", func() {
		It("Prefixes canonical string with '*' but omits it for simple string", func() {
			elemRef := &utils.FakeTypeRef{
				CanonicalStr: "pkg.MyType",
				SimpleStr:    "MyType",
			}
			ptrRef := &typeref.PtrTypeRef{Elem: elemRef}

			Expect(ptrRef.CanonicalString()).To(Equal("*pkg.MyType"))
			Expect(ptrRef.SimpleTypeString()).To(Equal("MyType"))
		})
	})

	Context("ToSymKey and CacheLookupKey", func() {
		It("Builds composite ptr key from element symbol key", func() {
			elemKey := graphs.NewNonUniverseBuiltInSymbolKey("pkg.Elem")
			elemRef := &utils.FakeTypeRef{SymKey: elemKey}
			ptrRef := &typeref.PtrTypeRef{Elem: elemRef}

			gotKey, err := ptrRef.ToSymKey(fVersion)
			Expect(err).ToNot(HaveOccurred())

			expectedKey := graphs.NewCompositeTypeKey(graphs.CompositeKindPtr, fVersion, []graphs.SymbolKey{elemKey})
			Expect(gotKey).To(Equal(expectedKey))
		})

		It("Propagates error when element ToSymKey fails", func() {
			elemErr := errors.New("elem-to-key-failed")
			badElem := &utils.FakeTypeRef{ToSymKeyErr: elemErr}
			ptrRef := &typeref.PtrTypeRef{Elem: badElem}

			_, err := ptrRef.ToSymKey(fVersion)
			Expect(err).To(MatchError(ContainSubstring("elem-to-key-failed")))
		})

		It("CacheLookupKey delegates to element ToSymKey (returns element key)", func() {
			elemKey := graphs.NewNonUniverseBuiltInSymbolKey("pkg.Elem")
			elemRef := &utils.FakeTypeRef{SymKey: elemKey}
			ptrRef := &typeref.PtrTypeRef{Elem: elemRef}

			cacheKey, err := ptrRef.CacheLookupKey(fVersion)
			Expect(err).ToNot(HaveOccurred())
			Expect(cacheKey).To(Equal(elemKey))
		})
	})

	Context("Flatten", func() {
		It("Returns a single-element slice containing the element TypeRef", func() {
			elemRef := &utils.FakeTypeRef{CanonicalStr: "pkg.Elem"}
			ptrRef := &typeref.PtrTypeRef{Elem: elemRef}

			flat := ptrRef.Flatten()
			Expect(flat).To(HaveLen(1))
			Expect(flat[0]).To(Equal(metadata.TypeRef(elemRef)))
		})
	})
})
