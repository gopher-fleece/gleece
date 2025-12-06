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

	Describe("ForceOrderedJSON", func() {
		It("should order JSON keys alphabetically", func() {
			// Unordered JSON input
			input := []byte(`{"zebra":"last","apple":"first","middle":"second"}`)

			result, err := ForceOrderedJSON(input)
			Expect(err).To(BeNil())

			// Expected output with keys ordered alphabetically
			expected := `{
  "apple": "first",
  "middle": "second",
  "zebra": "last"
}`
			Expect(string(result)).To(Equal(expected))
		})

		It("should order nested JSON objects", func() {
			input := []byte(`{"outer":{"zebra":"z","apple":"a"},"first":"value"}`)

			result, err := ForceOrderedJSON(input)
			Expect(err).To(BeNil())

			expected := `{
  "first": "value",
  "outer": {
    "apple": "a",
    "zebra": "z"
  }
}`
			Expect(string(result)).To(Equal(expected))
		})

		It("should handle arrays in JSON", func() {
			input := []byte(`{"items":[3,1,2],"name":"test"}`)

			result, err := ForceOrderedJSON(input)
			Expect(err).To(BeNil())

			// Arrays should maintain order, only keys should be sorted
			expected := `{
  "items": [
    3,
    1,
    2
  ],
  "name": "test"
}`
			Expect(string(result)).To(Equal(expected))
		})

		It("should produce deterministic output for same input", func() {
			input := []byte(`{"z":"last","a":"first","m":"middle"}`)

			result1, err1 := ForceOrderedJSON(input)
			Expect(err1).To(BeNil())

			result2, err2 := ForceOrderedJSON(input)
			Expect(err2).To(BeNil())

			// Both results should be identical
			Expect(string(result1)).To(Equal(string(result2)))
		})

		It("should handle complex OpenAPI-like structures", func() {
			input := []byte(`{
				"paths": {"/users": {"get": {}}},
				"openapi": "3.1.0",
				"info": {"title": "API"}
			}`)

			result, err := ForceOrderedJSON(input)
			Expect(err).To(BeNil())

			expected := `{
  "info": {
    "title": "API"
  },
  "openapi": "3.1.0",
  "paths": {
    "/users": {
      "get": {}
    }
  }
}`
			Expect(string(result)).To(Equal(expected))
		})

		It("should return error for invalid JSON", func() {
			input := []byte(`{invalid json}`)

			result, err := ForceOrderedJSON(input)
			Expect(err).NotTo(BeNil())
			Expect(result).To(BeNil())
			Expect(err.Error()).To(ContainSubstring("error unmarshaling JSON"))
		})

		It("should handle empty JSON object", func() {
			input := []byte(`{}`)

			result, err := ForceOrderedJSON(input)
			Expect(err).To(BeNil())
			Expect(string(result)).To(Equal("{}"))
		})

		It("should handle JSON with different value types", func() {
			input := []byte(`{"string":"text","number":42,"bool":true,"null":null}`)

			result, err := ForceOrderedJSON(input)
			Expect(err).To(BeNil())

			expected := `{
  "bool": true,
  "null": null,
  "number": 42,
  "string": "text"
}`
			Expect(string(result)).To(Equal(expected))
		})
	})
})
