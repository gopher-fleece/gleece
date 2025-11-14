package metadata_test

import (
	"strings"

	"github.com/gopher-fleece/gleece/core/metadata"
	"github.com/gopher-fleece/gleece/core/metadata/typeref"
	"github.com/gopher-fleece/gleece/graphs"
	"github.com/gopher-fleece/gleece/test/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Unit Tests - InlineStructTypeRef", func() {
	Describe("Kind", func() {
		It("Returns the InlineStruct kind", func() {
			inline := &typeref.InlineStructTypeRef{}
			Expect(inline.Kind()).To(Equal(metadata.TypeRefKindInlineStruct))
		})
	})

	Describe("string representations", func() {
		var (
			fieldTypeA *utils.FakeTypeRef
			fieldTypeB *utils.FakeTypeRef
		)

		BeforeEach(func() {
			fieldTypeA = &utils.FakeTypeRef{
				RefKind:      metadata.TypeRefKindNamed,
				CanonicalStr: "pkg.A",
				SimpleStr:    "A",
			}
			fieldTypeB = &utils.FakeTypeRef{
				RefKind:      metadata.TypeRefKindNamed,
				CanonicalStr: "pkg.B",
				SimpleStr:    "B",
			}
		})

		It("Builds canonical string including field names and canonical element strings", func() {
			inline := &typeref.InlineStructTypeRef{
				Fields: []metadata.FieldMeta{
					{
						SymNodeMeta: metadata.SymNodeMeta{
							Name: "First",
						},
						Type: metadata.TypeUsageMeta{Root: fieldTypeA}},
					{
						SymNodeMeta: metadata.SymNodeMeta{
							Name: "",
						},
						Type: metadata.TypeUsageMeta{Root: fieldTypeB},
					},
				},
			}

			canon := inline.CanonicalString()
			Expect(canon).To(ContainSubstring("inline{"))
			Expect(canon).To(ContainSubstring("First:pkg.A"))
			Expect(canon).To(ContainSubstring("pkg.B"))
			// ensure canonical and simple are different in this case
			simple := inline.SimpleTypeString()
			Expect(simple).To(ContainSubstring("First:A"))
			Expect(simple).To(ContainSubstring("B"))
			Expect(simple).ToNot(Equal(canon))
		})

		It("Returns minimal form for empty field list", func() {
			inline := &typeref.InlineStructTypeRef{}
			Expect(inline.CanonicalString()).To(Equal("inline{}"))
			Expect(inline.SimpleTypeString()).To(Equal("inline{}"))
		})

		It("Appends representative key suffix when RepKey is present", func() {
			repKey := graphs.NewNonUniverseBuiltInSymbolKey("somepkg.SomeAnon")
			inline := &typeref.InlineStructTypeRef{
				Fields: []metadata.FieldMeta{
					{
						SymNodeMeta: metadata.SymNodeMeta{
							Name: "X",
						},
						Type: metadata.TypeUsageMeta{Root: fieldTypeA},
					},
				},
				RepKey: repKey,
			}

			canon := inline.CanonicalString()
			Expect(canon).To(ContainSubstring("inline{"))
			Expect(canon).To(ContainSubstring("X:pkg.A"))
			// should have the '|' suffix appended for RepKey
			Expect(strings.Contains(canon, "|")).To(BeTrue(),
				"Canonical should include '|' when RepKey is present; got: %s", canon)
		})
	})

	Describe("ToSymKey / CacheLookupKey", func() {
		It("Errors when receiver is nil", func() {
			var inline *typeref.InlineStructTypeRef = nil
			_, err := inline.ToSymKey(nil)
			Expect(err).To(MatchError(ContainSubstring("nil InlineStructTypeRef")))
		})

		It("Errors when RepKey is not set", func() {
			inline := &typeref.InlineStructTypeRef{
				Fields: []metadata.FieldMeta{
					{
						SymNodeMeta: metadata.SymNodeMeta{
							Name: "f",
						},
						Type: metadata.TypeUsageMeta{Root: &utils.FakeTypeRef{}},
					},
				},
				RepKey: graphs.SymbolKey{}, // zero
			}
			_, err := inline.ToSymKey(nil)
			Expect(err).To(MatchError(ContainSubstring("inline struct missing RepKey")))
		})

		It("Returns the representative key when present", func() {
			repKey := graphs.NewNonUniverseBuiltInSymbolKey("mypkg.MyAnon")
			inline := &typeref.InlineStructTypeRef{
				Fields: []metadata.FieldMeta{
					{
						SymNodeMeta: metadata.SymNodeMeta{
							Name: "f",
						},
						Type: metadata.TypeUsageMeta{Root: &utils.FakeTypeRef{}},
					},
				},
				RepKey: repKey,
			}

			gotKey, err := inline.ToSymKey(nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(gotKey).To(Equal(repKey))

			// CacheLookupKey delegates to ToSymKey
			cacheKey, err := inline.CacheLookupKey(nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(cacheKey).To(Equal(repKey))
		})
	})

	Describe("Flatten", func() {
		It("Returns the TypeRef roots of each field in order", func() {
			fType1 := &utils.FakeTypeRef{
				RefKind:      metadata.TypeRefKindNamed,
				CanonicalStr: "pkg.One",
				SimpleStr:    "One",
			}
			fType2 := &utils.FakeTypeRef{
				RefKind:      metadata.TypeRefKindNamed,
				CanonicalStr: "pkg.Two",
				SimpleStr:    "Two",
			}

			inline := &typeref.InlineStructTypeRef{
				Fields: []metadata.FieldMeta{
					{
						SymNodeMeta: metadata.SymNodeMeta{
							Name: "A",
						}, Type: metadata.TypeUsageMeta{Root: fType1}},
					{
						SymNodeMeta: metadata.SymNodeMeta{
							Name: "B",
						}, Type: metadata.TypeUsageMeta{Root: fType2},
					},
				},
			}

			flattened := inline.Flatten()
			Expect(flattened).To(HaveLen(2))
			Expect(flattened[0]).To(Equal(metadata.TypeRef(fType1)))
			Expect(flattened[1]).To(Equal(metadata.TypeRef(fType2)))
		})

		It("Returns empty slice when no fields present", func() {
			inline := &typeref.InlineStructTypeRef{}
			flattened := inline.Flatten()
			// As flatten returns nil for default/leaf types, InlineStruct with no fields yields nil
			Expect(len(flattened)).To(BeZero())
		})
	})
})
