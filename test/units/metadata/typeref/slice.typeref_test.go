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

var _ = Describe("Unit Tests - SliceTypeRef", func() {
	var fVersion *gast.FileVersion

	BeforeEach(func() {
		fVersion = utils.MakeFileVersion("file", "")
	})

	Context("Kind", func() {
		It("Returns slice kind", func() {
			sliceRef := &typeref.SliceTypeRef{Elem: &utils.FakeTypeRef{}}
			Expect(sliceRef.Kind()).To(Equal(metadata.TypeRefKindSlice))
		})
	})

	Context("String representations", func() {
		It("Prefixes canonical and simple strings with '[]'", func() {
			elemRef := &utils.FakeTypeRef{
				CanonicalStr: "pkg.MyType",
				SimpleStr:    "MyType",
			}
			sliceRef := &typeref.SliceTypeRef{Elem: elemRef}

			Expect(sliceRef.CanonicalString()).To(Equal("[]pkg.MyType"))
			Expect(sliceRef.SimpleTypeString()).To(Equal("[]MyType"))
		})
	})

	Context("ToSymKey and CacheLookupKey", func() {
		It("Builds composite slice key from element symkey", func() {
			elemKey := graphs.NewNonUniverseBuiltInSymbolKey("pkg.Elem")
			elemRef := &utils.FakeTypeRef{SymKey: elemKey}
			sliceRef := &typeref.SliceTypeRef{Elem: elemRef}

			gotKey, err := sliceRef.ToSymKey(fVersion)
			Expect(err).ToNot(HaveOccurred())

			expectedKey := graphs.NewCompositeTypeKey(graphs.CompositeKindSlice, fVersion, []graphs.SymbolKey{elemKey})
			Expect(gotKey).To(Equal(expectedKey))
		})

		It("Propagates error when element ToSymKey fails", func() {
			elemErr := errors.New("elem-to-key-failed")
			badElem := &utils.FakeTypeRef{ToSymKeyErr: elemErr}
			sliceRef := &typeref.SliceTypeRef{Elem: badElem}

			_, err := sliceRef.ToSymKey(fVersion)
			Expect(err).To(MatchError(ContainSubstring("elem-to-key-failed")))
		})

		It("CacheLookupKey delegates to element CacheLookupKey", func() {
			elemKey := graphs.NewNonUniverseBuiltInSymbolKey("pkg.Elem")
			elemRef := &utils.FakeTypeRef{SymKey: elemKey}
			sliceRef := &typeref.SliceTypeRef{Elem: elemRef}

			cacheKey, err := sliceRef.CacheLookupKey(fVersion)
			Expect(err).ToNot(HaveOccurred())
			Expect(cacheKey).To(Equal(elemKey))
		})
	})

	Context("Flatten", func() {
		It("Returns a single-element slice containing the element TypeRef", func() {
			elemRef := &utils.FakeTypeRef{CanonicalStr: "pkg.Elem"}
			sliceRef := &typeref.SliceTypeRef{Elem: elemRef}

			flat := sliceRef.Flatten()
			Expect(flat).To(HaveLen(1))
			Expect(flat[0]).To(Equal(metadata.TypeRef(elemRef)))
		})
	})
})
