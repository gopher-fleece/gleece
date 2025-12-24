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

var _ = Describe("Unit Tests - NamedTypeRef", func() {
	var fVersion *gast.FileVersion

	BeforeEach(func() {
		fVersion = utils.MakeFileVersion("file", "")
	})

	Context("Kind", func() {
		It("Returns named kind", func() {
			named := &typeref.NamedTypeRef{}
			Expect(named.Kind()).To(Equal(metadata.TypeRefKindNamed))
		})
	})

	Context("String representations", func() {
		It("Builds simple string for plain named ref", func() {
			baseKey := graphs.NewNonUniverseBuiltInSymbolKey("pkg.Simple")
			named := &typeref.NamedTypeRef{Key: baseKey}
			Expect(named.SimpleTypeString()).To(Equal(baseKey.Name))
		})

		It("Includes type args in simple and canonical strings", func() {
			baseKey := graphs.NewNonUniverseBuiltInSymbolKey("pkg.Generic")
			arg1 := &utils.FakeTypeRef{CanonicalStr: "pkg.A", SimpleStr: "A", SymKey: graphs.NewNonUniverseBuiltInSymbolKey("pkg.A")}
			arg2 := &utils.FakeTypeRef{CanonicalStr: "pkg.B", SimpleStr: "B", SymKey: graphs.NewNonUniverseBuiltInSymbolKey("pkg.B")}

			named := &typeref.NamedTypeRef{
				Key:      baseKey,
				TypeArgs: []metadata.TypeRef{arg1, arg2},
			}

			simple := named.SimpleTypeString()
			Expect(simple).To(Equal("pkg.Generic[A,B]"))

			canonical := named.CanonicalString()
			// canonicalSymKey may include more structure; ensure base + arg canonical pieces are present
			Expect(canonical).To(ContainSubstring("pkg.Generic"))
			Expect(canonical).To(ContainSubstring("pkg.A"))
			Expect(canonical).To(ContainSubstring("pkg.B"))
		})
	})

	Context("ToSymKey and CacheLookupKey", func() {
		It("Returns base key when no type args", func() {
			baseKey := graphs.NewNonUniverseBuiltInSymbolKey("pkg.Base")
			named := &typeref.NamedTypeRef{Key: baseKey}

			got, err := named.ToSymKey(fVersion)
			Expect(err).ToNot(HaveOccurred())
			Expect(got).To(Equal(baseKey))

			cacheKey, err := named.CacheLookupKey(fVersion)
			Expect(err).ToNot(HaveOccurred())
			Expect(cacheKey).To(Equal(baseKey))
		})

		It("Builds instantiation key when type args present", func() {
			baseKey := graphs.NewNonUniverseBuiltInSymbolKey("pkg.G")
			argKey1 := graphs.NewNonUniverseBuiltInSymbolKey("pkg.Arg1")
			argKey2 := graphs.NewNonUniverseBuiltInSymbolKey("pkg.Arg2")

			argRef1 := &utils.FakeTypeRef{SymKey: argKey1}
			argRef2 := &utils.FakeTypeRef{SymKey: argKey2}

			named := &typeref.NamedTypeRef{
				Key:      baseKey,
				TypeArgs: []metadata.TypeRef{argRef1, argRef2},
			}

			got, err := named.ToSymKey(fVersion)
			Expect(err).ToNot(HaveOccurred())

			expected := graphs.NewInstSymbolKey(baseKey, []graphs.SymbolKey{argKey1, argKey2})
			Expect(got).To(Equal(expected))
		})

		It("Propagates error if an arg ToSymKey fails", func() {
			baseKey := graphs.NewNonUniverseBuiltInSymbolKey("pkg.G")
			argErr := fmt.Errorf("arg-to-key-failed")
			badArg := &utils.FakeTypeRef{ToSymKeyErr: argErr}

			named := &typeref.NamedTypeRef{
				Key:      baseKey,
				TypeArgs: []metadata.TypeRef{badArg},
			}

			_, err := named.ToSymKey(fVersion)
			Expect(err).To(MatchError(ContainSubstring("arg-to-key-failed")))
		})

		It("Errors when no base Key but type args exist", func() {
			argRef := &utils.FakeTypeRef{SymKey: graphs.NewNonUniverseBuiltInSymbolKey("pkg.X")}
			named := &typeref.NamedTypeRef{
				Key:      graphs.SymbolKey{}, // empty
				TypeArgs: []metadata.TypeRef{argRef},
			}

			_, err := named.ToSymKey(fVersion)
			Expect(err).To(MatchError(ContainSubstring("cannot instantiate named type without base Key")))
		})

		It("Errors when Key missing and no type args", func() {
			named := &typeref.NamedTypeRef{
				Key:      graphs.SymbolKey{}, // empty
				TypeArgs: nil,
			}
			_, err := named.ToSymKey(fVersion)
			Expect(err).To(MatchError(ContainSubstring("named type ref missing Key")))
		})
	})

	Context("Flatten", func() {
		It("Returns the TypeArgs slice (or nil when none)", func() {
			argRef := &utils.FakeTypeRef{CanonicalStr: "pkg.X"}
			namedWithArgs := &typeref.NamedTypeRef{Key: graphs.NewNonUniverseBuiltInSymbolKey("pkg.T"), TypeArgs: []metadata.TypeRef{argRef}}
			flat := namedWithArgs.Flatten()
			Expect(flat).To(HaveLen(1))
			Expect(flat[0]).To(Equal(metadata.TypeRef(argRef)))

			namedNoArgs := &typeref.NamedTypeRef{Key: graphs.NewNonUniverseBuiltInSymbolKey("pkg.T")}
			flat2 := namedNoArgs.Flatten()
			Expect(flat2).To(BeNil())
		})
	})
})
