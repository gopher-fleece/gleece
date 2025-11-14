package metadata_test

import (
	"errors"
	"fmt"

	"github.com/gopher-fleece/gleece/core/metadata"
	"github.com/gopher-fleece/gleece/core/metadata/typeref"
	"github.com/gopher-fleece/gleece/graphs"
	"github.com/gopher-fleece/gleece/test/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Unit Tests - ArrayTypeRef", func() {
	var (
		elemKey  = graphs.NewUniverseSymbolKey("elem")
		fVersion = utils.MakeFileVersion("file", "")
	)

	Context("Kind", func() {
		It("Returns the Array TypeRefKind", func() {
			a := &typeref.ArrayTypeRef{Len: nil, Elem: &utils.FakeTypeRef{RefKind: metadata.TypeRefKindArray}}
			Expect(a.Kind()).To(Equal(metadata.TypeRefKindArray))
		})
	})

	Context("CanonicalString", func() {
		It("Formats as slice when Len is nil", func() {
			elem := &utils.FakeTypeRef{CanonicalStr: "MyElem"}
			a := &typeref.ArrayTypeRef{Len: nil, Elem: elem}
			Expect(a.CanonicalString()).To(Equal("[]MyElem"))
		})

		It("Formats as fixed-length array when Len is set", func() {
			n := 5
			elem := &utils.FakeTypeRef{CanonicalStr: "X"}
			a := &typeref.ArrayTypeRef{Len: &n, Elem: elem}
			Expect(a.CanonicalString()).To(Equal("[5]X"))
		})
	})

	Context("SimpleTypeString", func() {
		It("Prefixes element simple string with []", func() {
			elem := &utils.FakeTypeRef{SimpleStr: "T"}
			a := &typeref.ArrayTypeRef{Len: nil, Elem: elem}
			Expect(a.SimpleTypeString()).To(Equal("[]T"))
		})
	})

	Context("CacheLookupKey", func() {
		It("Delegates to element CacheLookupKey", func() {
			f := &utils.FakeTypeRef{SymKey: elemKey}
			a := &typeref.ArrayTypeRef{Len: nil, Elem: f}
			k, err := a.CacheLookupKey(fVersion)
			Expect(err).ToNot(HaveOccurred())
			Expect(k).To(Equal(elemKey))
		})
	})

	Context("ToSymKey", func() {
		It("Builds a composite array key when element ToSymKey succeeds", func() {
			f := &utils.FakeTypeRef{SymKey: elemKey}
			a := &typeref.ArrayTypeRef{Len: nil, Elem: f}

			got, err := a.ToSymKey(fVersion)
			Expect(err).ToNot(HaveOccurred())

			want := graphs.NewCompositeTypeKey(graphs.CompositeKindArray, fVersion, []graphs.SymbolKey{elemKey})
			Expect(got).To(Equal(want))
		})

		It("Propagates element ToSymKey errors", func() {
			expErr := errors.New("boom")
			f := &utils.FakeTypeRef{ToSymKeyErr: expErr}
			a := &typeref.ArrayTypeRef{Len: nil, Elem: f}

			_, err := a.ToSymKey(fVersion)
			Expect(err).To(MatchError(ContainSubstring("boom")))
		})
	})

	Context("Flatten", func() {
		It("Returns a non-nil slice of TypeRefs and canonical strings are usable", func() {
			// provide a fake Flatten response and ensure result is returned and sane
			elem := &utils.FakeTypeRef{
				CanonicalStr:    "E",
				FlattenResponse: []metadata.TypeRef{&utils.FakeTypeRef{CanonicalStr: "leaf"}},
			}
			a := &typeref.ArrayTypeRef{Len: nil, Elem: elem}

			res := a.Flatten()
			Expect(res).ToNot(BeNil())
			Expect(len(res)).To(BeNumerically(">=", 0))

			// ensure calling CanonicalString on each element doesn't panic and returns non-empty
			for i, r := range res {
				cs := r.CanonicalString()
				Expect(cs).ToNot(BeEmpty(), fmt.Sprintf("element %d had empty canonical string", i))
			}
		})
	})
})
