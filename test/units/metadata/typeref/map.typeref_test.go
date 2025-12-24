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

var _ = Describe("Unit Tests - MapTypeRef", func() {
	var fVersion *gast.FileVersion

	BeforeEach(func() {
		fVersion = utils.MakeFileVersion("file", "")
	})

	Context("Kind", func() {
		It("Returns map kind", func() {
			m := &typeref.MapTypeRef{}
			Expect(m.Kind()).To(Equal(metadata.TypeRefKindMap))
		})
	})

	Context("String representations", func() {
		It("Builds canonical and simple strings from key/value", func() {
			keyRef := &utils.FakeTypeRef{CanonicalStr: "pkg.Key", SimpleStr: "Key"}
			valRef := &utils.FakeTypeRef{CanonicalStr: "pkg.Val", SimpleStr: "Val"}

			m := &typeref.MapTypeRef{Key: keyRef, Value: valRef}
			Expect(m.CanonicalString()).To(Equal("map[pkg.Key]pkg.Val"))
			Expect(m.SimpleTypeString()).To(Equal("map[Key]Val"))
		})
	})

	Context("ToSymKey / CacheLookupKey", func() {
		It("Creates composite key when both operands produce keys", func() {
			keySym := graphs.NewNonUniverseBuiltInSymbolKey("kpkg.K")
			valSym := graphs.NewNonUniverseBuiltInSymbolKey("vpkg.V")
			keyRef := &utils.FakeTypeRef{SymKey: keySym}
			valRef := &utils.FakeTypeRef{SymKey: valSym}

			m := &typeref.MapTypeRef{Key: keyRef, Value: valRef}
			got, err := m.ToSymKey(fVersion)
			Expect(err).ToNot(HaveOccurred())

			expected := graphs.NewCompositeTypeKey(graphs.CompositeKindMap, fVersion, []graphs.SymbolKey{keySym, valSym})
			Expect(got).To(Equal(expected))

			// CacheLookupKey delegates to ToSymKey
			cacheKey, err := m.CacheLookupKey(fVersion)
			Expect(err).ToNot(HaveOccurred())
			Expect(cacheKey).To(Equal(expected))
		})

		It("Returns error when key ToSymKey fails", func() {
			keyErr := fmt.Errorf("key-fail")
			keyRef := &utils.FakeTypeRef{ToSymKeyErr: keyErr}
			valRef := &utils.FakeTypeRef{SymKey: graphs.NewNonUniverseBuiltInSymbolKey("vpkg.V")}

			m := &typeref.MapTypeRef{Key: keyRef, Value: valRef}
			_, err := m.ToSymKey(fVersion)
			Expect(err).To(MatchError(ContainSubstring("key-fail")))
		})

		It("Returns error when value ToSymKey fails", func() {
			keyRef := &utils.FakeTypeRef{SymKey: graphs.NewNonUniverseBuiltInSymbolKey("kpkg.K")}
			valErr := fmt.Errorf("val-fail")
			valRef := &utils.FakeTypeRef{ToSymKeyErr: valErr}

			m := &typeref.MapTypeRef{Key: keyRef, Value: valRef}
			_, err := m.ToSymKey(fVersion)
			Expect(err).To(MatchError(ContainSubstring("val-fail")))
		})
	})

	Context("Flatten", func() {
		It("Returns key and value TypeRefs in order", func() {
			keyRef := &utils.FakeTypeRef{CanonicalStr: "pkg.Key"}
			valRef := &utils.FakeTypeRef{CanonicalStr: "pkg.Val"}

			m := &typeref.MapTypeRef{Key: keyRef, Value: valRef}
			flattened := m.Flatten()
			Expect(flattened).To(HaveLen(2))
			Expect(flattened[0]).To(Equal(metadata.TypeRef(keyRef)))
			Expect(flattened[1]).To(Equal(metadata.TypeRef(valRef)))
		})
	})
})
