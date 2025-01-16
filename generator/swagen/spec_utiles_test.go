package swagen

import (
	"github.com/haimkastner/gleece/definitions"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Spec Utilities", func() {

	Describe("HttpStatusCodeToString", func() {
		It("should convert HttpStatusCode to string", func() {
			code := definitions.HttpStatusCode(200)
			Expect(HttpStatusCodeToString(code)).To(Equal("200"))
		})
	})

	Describe("ParseNumber", func() {
		It("should parse a valid number string", func() {
			Expect(*ParseNumber("123.45")).To((BeEquivalentTo(123.45)))
		})

		It("should return nil for an invalid number string", func() {
			Expect(ParseNumber("abc")).To(BeNil())
		})
	})

	Describe("ParseInteger", func() {
		It("should parse a valid integer string", func() {
			Expect(*ParseInteger("123")).To((BeEquivalentTo(123)))
		})

		It("should return nil for an invalid integer string", func() {
			Expect(ParseInteger("abc")).To(BeNil())
		})
	})

	Describe("ParseBool", func() {
		It("should parse a valid boolean string", func() {
			Expect(*ParseBool("true")).To((BeTrue()))
			Expect(*ParseBool("false")).To((BeFalse()))
		})

		It("should return nil for an invalid boolean string", func() {
			Expect(ParseBool("notabool")).To(BeNil())
		})
	})

	Describe("IsFieldRequired", func() {
		It("should return true if 'required' is present in the validation string", func() {
			Expect(IsFieldRequired("required")).To(BeTrue())
			Expect(IsFieldRequired("min=1,required")).To(BeTrue())
		})

		It("should return false if 'required' is not present in the validation string", func() {
			Expect(IsFieldRequired("min=1,max=10")).To(BeFalse())
		})
	})

	Describe("GetArrayItemType", func() {
		It("should return the item type of an array", func() {
			Expect(GetArrayItemType("[]string")).To(Equal("string"))
			Expect(GetArrayItemType("[]int")).To(Equal("int"))
		})

		It("should return the sub array item type of an array", func() {
			Expect(GetArrayItemType("[][]string")).To(Equal("[]string"))
			Expect(GetArrayItemType("[][][]abc")).To(Equal("[][]abc"))
		})
	})
})
