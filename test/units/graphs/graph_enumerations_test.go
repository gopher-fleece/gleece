package graphs_test

import (
	"github.com/gopher-fleece/gleece/graphs/symboldg"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Unit Tests - SymbolGraph", func() {
	Context("ToPrimitiveType", func() {
		It("Returns true for a valid primitive type", func() {
			t, ok := symboldg.ToPrimitiveType("int64")
			Expect(ok).To(BeTrue())
			Expect(t).To(Equal(symboldg.PrimitiveTypeInt64))
		})

		It("Returns true for an alias type like byte", func() {
			t, ok := symboldg.ToPrimitiveType("byte")
			Expect(ok).To(BeTrue())
			Expect(t).To(Equal(symboldg.PrimitiveTypeByte))
		})

		It("Returns false for an unknown type", func() {
			t, ok := symboldg.ToPrimitiveType("notatype")
			Expect(ok).To(BeFalse())
			Expect(t).To(Equal(symboldg.PrimitiveType("")))
		})
	})

	Context("SpecialType.IsUniverse", func() {
		It("Returns true for error", func() {
			Expect(symboldg.SpecialTypeError.IsUniverse()).To(BeTrue())
		})

		It("Returns true for interface{}", func() {
			Expect(symboldg.SpecialTypeEmptyInterface.IsUniverse()).To(BeTrue())
		})

		It("Returns true for any", func() {
			Expect(symboldg.SpecialTypeAny.IsUniverse()).To(BeTrue())
		})

		It("Returns false for non-universe type", func() {
			Expect(symboldg.SpecialTypeTime.IsUniverse()).To(BeFalse())
		})
	})

	Context("ToSpecialType", func() {
		It("Returns true for error", func() {
			t, ok := symboldg.ToSpecialType("error")
			Expect(ok).To(BeTrue())
			Expect(t).To(Equal(symboldg.SpecialTypeError))
		})

		It("Returns true for interface{}", func() {
			t, ok := symboldg.ToSpecialType("interface{}")
			Expect(ok).To(BeTrue())
			Expect(t).To(Equal(symboldg.SpecialTypeEmptyInterface))
		})

		It("Returns true for any", func() {
			t, ok := symboldg.ToSpecialType("any")
			Expect(ok).To(BeTrue())
			Expect(t).To(Equal(symboldg.SpecialTypeAny))
		})

		It("Returns true for context.Context", func() {
			t, ok := symboldg.ToSpecialType("context.Context")
			Expect(ok).To(BeTrue())
			Expect(t).To(Equal(symboldg.SpecialTypeContext))
		})

		It("Returns true for time.Time", func() {
			t, ok := symboldg.ToSpecialType("time.Time")
			Expect(ok).To(BeTrue())
			Expect(t).To(Equal(symboldg.SpecialTypeTime))
		})

		It("Returns true for unsafe.Pointer", func() {
			t, ok := symboldg.ToSpecialType("unsafe.Pointer")
			Expect(ok).To(BeTrue())
			Expect(t).To(Equal(symboldg.SpecialTypeUnsafePointer))
		})

		It("Returns false for an unknown special type", func() {
			t, ok := symboldg.ToSpecialType("does.NotExist")
			Expect(ok).To(BeFalse())
			Expect(t).To(Equal(symboldg.SpecialType("")))
		})
	})

})
