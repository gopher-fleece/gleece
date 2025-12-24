package metadata_test

import (
	"github.com/gopher-fleece/gleece/v2/core/metadata"
	"github.com/gopher-fleece/gleece/v2/core/metadata/typeref"
	"github.com/gopher-fleece/gleece/v2/gast"
	"github.com/gopher-fleece/gleece/v2/graphs"
	"github.com/gopher-fleece/gleece/v2/test/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Unit Tests - ParamTypeRef", func() {

	var _ = Describe("ParamTypeRef", func() {
		Context("Kind", func() {
			It("Returns TypeRefKindParam", func() {
				param := &typeref.ParamTypeRef{Name: "T", Index: 0}
				Expect(param.Kind()).To(Equal(metadata.TypeRefKindParam))
			})
		})

		Context("CanonicalString", func() {
			It("Shows index when Index >= 0", func() {
				param := &typeref.ParamTypeRef{Name: "T", Index: 2}
				Expect(param.CanonicalString()).To(Equal("P#2"))
			})

			It("Shows name when Index is negative", func() {
				param := &typeref.ParamTypeRef{Name: "T", Index: -1}
				Expect(param.CanonicalString()).To(Equal("P{T}"))
			})
		})

		Context("SimpleTypeString", func() {
			It("Delegates to canonical representation", func() {
				param := &typeref.ParamTypeRef{Name: "T", Index: 3}
				Expect(param.SimpleTypeString()).To(Equal(param.CanonicalString()))
			})
		})

		Context("ToSymKey and CacheLookupKey", func() {
			var fVersion *gast.FileVersion

			BeforeEach(func() {
				fVersion = utils.MakeFileVersion("file", "")
			})

			It("Errors when fileVersion is nil", func() {
				param := &typeref.ParamTypeRef{Name: "T", Index: 1}
				_, err := param.ToSymKey(nil)
				Expect(err).To(HaveOccurred())
				_, err2 := param.CacheLookupKey(nil)
				Expect(err2).To(HaveOccurred())
			})

			It("Returns a Param SymbolKey when fileVersion is provided", func() {
				param := &typeref.ParamTypeRef{Name: "MyT", Index: 5}
				key, err := param.ToSymKey(fVersion)
				Expect(err).ToNot(HaveOccurred())

				expected := graphs.NewParamSymbolKey(fVersion, "MyT", 5)
				Expect(key.Equals(expected)).To(BeTrue())

				// CacheLookupKey delegates to ToSymKey
				cacheKey, err := param.CacheLookupKey(fVersion)
				Expect(err).ToNot(HaveOccurred())
				Expect(cacheKey.Equals(expected)).To(BeTrue())
			})
		})

		Context("Flatten", func() {
			It("Returns nil when flattening a type parameter", func() {
				param := &typeref.ParamTypeRef{Name: "Z", Index: -1}
				flat := param.Flatten()
				Expect(flat).To(BeNil())
			})
		})
	})
})
