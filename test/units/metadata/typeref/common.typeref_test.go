package metadata_test

import (
	"github.com/gopher-fleece/gleece/core/metadata"
	"github.com/gopher-fleece/gleece/core/metadata/typeref"
	"github.com/gopher-fleece/gleece/graphs"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// These cases are to test the 'private' TypeRef commons logic.
// We're using concrete implementations of TypeRef here as the entrypoints to the package.
var _ = Describe("Unit Tests - TypeRef Commons", func() {
	Context("Flatten", func() {
		It("Returns Elem for Ptr, Slice and Array types", func() {
			base := mustNamedPtr("Base", "")
			ptr := &typeref.PtrTypeRef{Elem: base}
			slice := &typeref.SliceTypeRef{Elem: base}
			array := &typeref.ArrayTypeRef{Elem: base}

			Expect(ptr.Flatten()).To(HaveLen(1))
			Expect(ptr.Flatten()[0].CanonicalString()).To(Equal(base.CanonicalString()))

			Expect(slice.Flatten()).To(HaveLen(1))
			Expect(slice.Flatten()[0].CanonicalString()).To(Equal(base.CanonicalString()))

			Expect(array.Flatten()).To(HaveLen(1))
			Expect(array.Flatten()[0].CanonicalString()).To(Equal(base.CanonicalString()))
		})

		It("Returns Key and Value for Map types", func() {
			k := mustNamedPtr("K", "")
			v := mustNamedPtr("V", "")
			m := &typeref.MapTypeRef{Key: k, Value: v}

			out := m.Flatten()
			Expect(out).To(HaveLen(2))
			Expect(out[0].CanonicalString()).To(Equal(k.CanonicalString()))
			Expect(out[1].CanonicalString()).To(Equal(v.CanonicalString()))
		})

		It("Returns Params and Results for Func types", func() {
			p1 := mustNamedPtr("P1", "")
			p2 := mustNamedPtr("P2", "")
			r1 := mustNamedPtr("R1", "")
			f := &typeref.FuncTypeRef{Params: []metadata.TypeRef{p1, p2}, Results: []metadata.TypeRef{r1}}

			out := f.Flatten()
			Expect(out).To(HaveLen(3))
			Expect(out[0].CanonicalString()).To(Equal(p1.CanonicalString()))
			Expect(out[1].CanonicalString()).To(Equal(p2.CanonicalString()))
			Expect(out[2].CanonicalString()).To(Equal(r1.CanonicalString()))
		})

		It("Returns TypeArgs for Named types", func() {
			a1 := mustNamedPtr("A1", "")
			a2 := mustNamedPtr("A2", "")
			key := graphs.SymbolKey{Name: "MyType", FileId: "f1"}
			n := typeref.NewNamedTypeRef(&key, []metadata.TypeRef{a1, a2})
			out := (&n).Flatten()
			Expect(out).To(HaveLen(2))
			Expect(out[0].CanonicalString()).To(Equal(a1.CanonicalString()))
			Expect(out[1].CanonicalString()).To(Equal(a2.CanonicalString()))
		})

		It("Returns field root types for InlineStruct types", func() {
			f1 := mustNamedPtr("F1", "")
			fields := []metadata.FieldMeta{
				{Type: metadata.TypeUsageMeta{Root: f1}},
				{Type: metadata.TypeUsageMeta{Root: mustNamedPtr("Anon", "")}},
			}
			in := &typeref.InlineStructTypeRef{Fields: fields}

			out := in.Flatten()
			Expect(out).To(HaveLen(2))
			Expect(out[0].CanonicalString()).To(Equal(f1.CanonicalString()))
			Expect(out[1].CanonicalString()).To(Equal(fields[1].Type.Root.CanonicalString()))
		})
	})

	Context("Canonical Sym Key through public APIs", func() {
		It("NamedTypeRef CanonicalString uses universe/builtin name only", func() {
			k := graphs.SymbolKey{Name: "int", IsBuiltIn: true}
			n := typeref.NewNamedTypeRef(&k, nil)
			Expect((&n).CanonicalString()).To(Equal("int"))
		})

		It("NamedTypeRef CanonicalString prefers FileId over FilePath", func() {
			k := graphs.SymbolKey{Name: "T", FileId: "file-123", FilePath: "some/path"}
			n := typeref.NewNamedTypeRef(&k, nil)
			Expect((&n).CanonicalString()).To(Equal("T|file-123"))
		})

		It("NamedTypeRef CanonicalString falls back to FilePath when no FileId", func() {
			k := graphs.SymbolKey{Name: "T", FilePath: "some/path"}
			n := typeref.NewNamedTypeRef(&k, nil)
			Expect((&n).CanonicalString()).To(Equal("T|some/path"))
		})

		It("InlineStruct CanonicalString appends repkey when present", func() {
			f1 := mustNamedPtr("S1", "")
			fields := []metadata.FieldMeta{{Type: metadata.TypeUsageMeta{Root: f1}}}
			rep := graphs.SymbolKey{Name: "anon", FileId: "rep-1"}
			in := &typeref.InlineStructTypeRef{Fields: fields, RepKey: rep}
			cs := in.CanonicalString()
			Expect(cs).To(ContainSubstring("inline{"))
			// repkey should be appended after a pipe
			Expect(cs).To(ContainSubstring("|" + rep.Name + "|"))
		})
	})
})

// ---------------- helpers ----------------
func mustNamedPtr(name, fileId string) *typeref.NamedTypeRef {
	k := graphs.SymbolKey{Name: name, FileId: fileId}
	r := typeref.NewNamedTypeRef(&k, nil)
	return &r
}
