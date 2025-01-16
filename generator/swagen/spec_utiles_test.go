package swagen_test

import (
	"testing"

	"github.com/haimkastner/gleece/definitions"
	"github.com/haimkastner/gleece/generator/swagen"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var testTitle = "Spec Utilities"

func TestSpecUtilities(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, testTitle)
}

var _ = Describe(testTitle, func() {

	Describe("HttpStatusCodeToString", func() {
		It("should convert HttpStatusCode to string", func() {
			code := definitions.HttpStatusCode(200)
			Expect(swagen.HttpStatusCodeToString(code)).To(Equal("200"))
		})
	})

	Describe("ParseNumber", func() {
		It("should parse a valid number string", func() {
			Expect(*swagen.ParseNumber("123.45")).To((BeEquivalentTo(123.45)))
		})

		It("should return nil for an invalid number string", func() {
			Expect(swagen.ParseNumber("abc")).To(BeNil())
		})
	})

	Describe("ParseInteger", func() {
		It("should parse a valid integer string", func() {
			Expect(*swagen.ParseInteger("123")).To((BeEquivalentTo(123)))
		})

		It("should return nil for an invalid integer string", func() {
			Expect(swagen.ParseInteger("abc")).To(BeNil())
		})
	})

	Describe("ParseBool", func() {
		It("should parse a valid boolean string", func() {
			Expect(*swagen.ParseBool("true")).To((BeTrue()))
			Expect(*swagen.ParseBool("false")).To((BeFalse()))
		})

		It("should return nil for an invalid boolean string", func() {
			Expect(swagen.ParseBool("notabool")).To(BeNil())
		})
	})

	Describe("IsFieldRequired", func() {
		It("should return true if 'required' is present in the validation string", func() {
			Expect(swagen.IsFieldRequired("required")).To(BeTrue())
			Expect(swagen.IsFieldRequired("min=1,required")).To(BeTrue())
		})

		It("should return false if 'required' is not present in the validation string", func() {
			Expect(swagen.IsFieldRequired("min=1,max=10")).To(BeFalse())
		})
	})

	Describe("GetArrayItemType", func() {
		It("should return the item type of an array", func() {
			Expect(swagen.GetArrayItemType("[]string")).To(Equal("string"))
			Expect(swagen.GetArrayItemType("[]int")).To(Equal("int"))
		})

		It("should return the sub array item type of an array", func() {
			Expect(swagen.GetArrayItemType("[][]string")).To(Equal("[]string"))
			Expect(swagen.GetArrayItemType("[][][]abc")).To(Equal("[][]abc"))
		})
	})
})
