package graphs_test

import (
	"fmt"

	"github.com/gopher-fleece/gleece/v2/common"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Unit Tests - SymbolGraph", func() {
	Context("ToPrimitiveType", func() {
		It("Returns true for a valid primitive type", func() {
			t, ok := common.ToPrimitiveType("int64")
			Expect(ok).To(BeTrue())
			Expect(t).To(Equal(common.PrimitiveTypeInt64))
		})

		It("Returns true for an alias type like byte", func() {
			t, ok := common.ToPrimitiveType("byte")
			Expect(ok).To(BeTrue())
			Expect(t).To(Equal(common.PrimitiveTypeByte))
		})

		It("Returns false for an unknown type", func() {
			t, ok := common.ToPrimitiveType("notatype")
			Expect(ok).To(BeFalse())
			Expect(t).To(Equal(common.PrimitiveType("")))
		})
	})

	Context("SpecialType.IsUniverse", func() {
		It("Returns true for error", func() {
			Expect(common.SpecialTypeError.IsUniverse()).To(BeTrue())
		})

		It("Returns true for interface{}", func() {
			Expect(common.SpecialTypeEmptyInterface.IsUniverse()).To(BeTrue())
		})

		It("Returns true for any", func() {
			Expect(common.SpecialTypeAny.IsUniverse()).To(BeTrue())
		})

		It("Returns false for non-universe type", func() {
			Expect(common.SpecialTypeTime.IsUniverse()).To(BeFalse())
		})
	})

	Context("ToSpecialType", func() {
		It("Returns true for error", func() {
			t, ok := common.ToSpecialType("error")
			Expect(ok).To(BeTrue())
			Expect(t).To(Equal(common.SpecialTypeError))
		})

		It("Returns true for interface{}", func() {
			t, ok := common.ToSpecialType("interface{}")
			Expect(ok).To(BeTrue())
			Expect(t).To(Equal(common.SpecialTypeEmptyInterface))
		})

		It("Returns true for any", func() {
			t, ok := common.ToSpecialType("any")
			Expect(ok).To(BeTrue())
			Expect(t).To(Equal(common.SpecialTypeAny))
		})

		It("Returns true for context.Context", func() {
			t, ok := common.ToSpecialType("context.Context")
			Expect(ok).To(BeTrue())
			Expect(t).To(Equal(common.SpecialTypeContext))
		})

		It("Returns true for time.Time", func() {
			t, ok := common.ToSpecialType("time.Time")
			Expect(ok).To(BeTrue())
			Expect(t).To(Equal(common.SpecialTypeTime))
		})

		It("Returns true for unsafe.Pointer", func() {
			t, ok := common.ToSpecialType("unsafe.Pointer")
			Expect(ok).To(BeTrue())
			Expect(t).To(Equal(common.SpecialTypeUnsafePointer))
		})

		It("Returns false for an unknown special type", func() {
			t, ok := common.ToSpecialType("does.NotExist")
			Expect(ok).To(BeFalse())
			Expect(t).To(Equal(common.SpecialType("")))
		})
	})

	Context("ToSymbolKind", func() {
		validKinds := map[string]common.SymKind{
			"Unknown":    common.SymKindUnknown,
			"Package":    common.SymKindPackage,
			"Struct":     common.SymKindStruct,
			"Controller": common.SymKindController,
			"Interface":  common.SymKindInterface,
			"Alias":      common.SymKindAlias,
			"Enum":       common.SymKindEnum,
			"EnumValue":  common.SymKindEnumValue,
			"Function":   common.SymKindFunction,
			"Receiver":   common.SymKindReceiver,
			"Field":      common.SymKindField,
			"Parameter":  common.SymKindParameter,
			"Variable":   common.SymKindVariable,
			"Constant":   common.SymKindConstant,
			"RetType":    common.SymKindReturnType,
			"Builtin":    common.SymKindBuiltin,
			"Special":    common.SymKindSpecialBuiltin,
		}

		It("converts valid strings to SymKind without error", func() {
			for str, expected := range validKinds {
				result, err := common.ToSymbolKind(str)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(Equal(expected), fmt.Sprintf("string %q", str))
			}
		})

		It("returns an error for invalid strings", func() {
			invalidInputs := []string{"foo", "Bar", "", "123"}

			for _, input := range invalidInputs {
				result, err := common.ToSymbolKind(input)
				Expect(err).To(HaveOccurred(), fmt.Sprintf("input: %q", input))
				Expect(result).To(Equal(common.SymKind("")))
			}
		})
	})
})
