package swagtool

import (
	"github.com/gopher-fleece/runtime"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Swagtools - Spec Utilities", func() {

	Describe("HttpStatusCodeToString", func() {
		It("should convert HttpStatusCode to string", func() {
			code := runtime.HttpStatusCode(200)
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

	Describe("ParseUInteger", func() {
		It("should parse a valid integer string", func() {
			Expect(*ParseUInteger("123")).To((BeEquivalentTo(123)))
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
})
