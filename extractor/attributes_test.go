package extractor_test

import (
	"testing"

	"github.com/gopher-fleece/gleece/extractor"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Attributes Holder", func() {

	Context("Given a single comment", func() {

		When("Attribute is simple", func() {

			It("Constructs without error", func() {
				comments := []string{"// @Description Abcd"}
				_, err := extractor.NewAttributeHolder(comments)
				Expect(err).To(BeNil())
			})

			It("Correctly detects attribute exists", func() {
				comments := []string{"// @Description Abcd"}
				holder, _ := extractor.NewAttributeHolder(comments)
				Expect(holder.Has(extractor.AttributeDescription)).To(BeTrue())
			})

			It("Correctly gets the attribute", func() {
				comments := []string{"// @Description Abcd"}
				holder, _ := extractor.NewAttributeHolder(comments)
				attrib := holder.GetFirst(extractor.AttributeDescription)
				Expect(attrib).ToNot(BeNil())
			})

			It("Returns correct value from the GetDescription method", func() {
				comments := []string{"// @Description Abcd"}
				holder, _ := extractor.NewAttributeHolder(comments)
				Expect(holder.GetDescription()).To(Equal("Abcd"))
			})
		})

		When("Attribute is complex", func() {
			It("Constructs without error", func() {
				comments := []string{`// @Security(securitySchemaName, { scopes: ["read:users", "write:users"] }) Abcd`}
				_, err := extractor.NewAttributeHolder(comments)
				Expect(err).To(BeNil())
			})

			It("Correctly detects attribute exists", func() {
				comments := []string{`// @Security(securitySchemaName, { scopes: ["read:users", "write:users"] }) Abcd`}
				holder, _ := extractor.NewAttributeHolder(comments)
				Expect(holder.Has(extractor.AttributeSecurity)).To(BeTrue())
			})

			It("Correctly gets the attribute", func() {
				comments := []string{`// @Security(securitySchemaName, { scopes: ["read:users", "write:users"] }) Abcd`}
				holder, _ := extractor.NewAttributeHolder(comments)
				attrib := holder.GetFirst(extractor.AttributeSecurity)
				Expect(attrib).ToNot(BeNil())
			})

			It("Attribute has correct basic values", func() {
				comments := []string{`// @Security(securitySchemaName, { scopes: ["read:users", "write:users"] }) Abcd`}
				holder, _ := extractor.NewAttributeHolder(comments)
				attrib := holder.GetFirst(extractor.AttributeSecurity)

				Expect(attrib.Name).To(Equal(extractor.AttributeSecurity))
				Expect(attrib.Value).To(Equal("securitySchemaName"))
				Expect(attrib.Description).To(Equal("Abcd"))
			})

			It("Attribute has correct properties", func() {
				comments := []string{`// @Security(securitySchemaName, { scopes: ["read:users", "write:users"] }) Abcd`}
				holder, _ := extractor.NewAttributeHolder(comments)
				attrib := holder.GetFirst(extractor.AttributeSecurity)

				Expect(len(attrib.Properties)).To(Equal(1))
				Expect(attrib.HasProperty(extractor.PropertySecurityScopes)).To(BeTrue())
				value, err := extractor.GetCastProperty[[]string](attrib, extractor.PropertySecurityScopes)
				Expect(err).To(BeNil())
				Expect(*value).To(HaveExactElements("read:users", "write:users"))
			})
		})
	})

	Context("Given multiple comments", func() {
		stdComments := []string{
			`// @Description Create a new user`,
			`// @Method(POST) This text is not part of the OpenAPI spec`,
			`// @Route(/user/{user_name}/{user_id}/{user_id_2}) Same here`,
			`// @Query(email, { validate: "required,email" }) The user's email`,
			`// @Path(id, { name: "user_id", validate:"gt=1" }) The user's ID`,
			`// @Path(id2, { name: "user_id_2", validate:"gt=10" }) The user's ID 2`,
			`// @Path(name, { name: "user_name" }) The user's name`,
			`// @Body(domicile) The user's domicile`,
			`// @Header(origin, { name: "x-origin" }) The request origin`,
			`// @Header(trace) The trace info`,
			`// @Response(200) The ID of the newly created user`,
			`// @ErrorResponse(500) The error when process failed`,
			`// @Security(schema1, { scopes: ["read:users", "write:users"] })`,
			`// @Security(schema2, { scopes: ["read:devices", "write:devices"] })`,
		}

		It("Constructs without error", func() {
			_, err := extractor.NewAttributeHolder(stdComments)
			Expect(err).To(BeNil())
		})

		It("Correctly detects all attributes exist", func() {
			holder, _ := extractor.NewAttributeHolder(stdComments)

			Expect(holder.Has(extractor.AttributeDescription)).To(BeTrue())
			Expect(holder.Has(extractor.AttributeMethod)).To(BeTrue())
			Expect(holder.Has(extractor.AttributeRoute)).To(BeTrue())
			Expect(holder.Has(extractor.AttributeQuery)).To(BeTrue())
			Expect(holder.Has(extractor.AttributePath)).To(BeTrue())
			Expect(holder.Has(extractor.AttributeBody)).To(BeTrue())
			Expect(holder.Has(extractor.AttributeHeader)).To(BeTrue())
			Expect(holder.Has(extractor.AttributeResponse)).To(BeTrue())
			Expect(holder.Has(extractor.AttributeErrorResponse)).To(BeTrue())
			Expect(holder.Has(extractor.AttributeSecurity)).To(BeTrue())
		})

		It("Correctly gets all attributes of the same type", func() {
			holder, _ := extractor.NewAttributeHolder(stdComments)
			attributes := holder.GetAll(extractor.AttributePath)
			Expect(attributes).To(HaveLen(3))
		})

		It("Attributes of the same type are ordered and have correct values", func() {
			holder, _ := extractor.NewAttributeHolder(stdComments)
			attributes := holder.GetAll(extractor.AttributePath)

			Expect(attributes[0].Value).To(Equal("id"))
			Expect(attributes[1].Value).To(Equal("id2"))
			Expect(attributes[2].Value).To(Equal("name"))

			value1, err1 := extractor.GetCastProperty[string](attributes[0], extractor.PropertyName)
			value2, err2 := extractor.GetCastProperty[string](attributes[1], extractor.PropertyName)
			value3, err3 := extractor.GetCastProperty[string](attributes[2], extractor.PropertyName)

			Expect(err1).To(BeNil())
			Expect(err2).To(BeNil())
			Expect(err3).To(BeNil())

			Expect(*value1).To(Equal("user_id"))
			Expect(*value2).To(Equal("user_id_2"))
			Expect(*value3).To(Equal("user_name"))

			value1, err1 = extractor.GetCastProperty[string](attributes[0], extractor.PropertyValidatorString)
			value2, err2 = extractor.GetCastProperty[string](attributes[1], extractor.PropertyValidatorString)
			value3, err3 = extractor.GetCastProperty[string](attributes[2], extractor.PropertyValidatorString)

			Expect(err1).To(BeNil())
			Expect(err2).To(BeNil())
			Expect(err3).To(BeNil())

			Expect(*value1).To(Equal("gt=1"))
			Expect(*value2).To(Equal("gt=10"))
			Expect(value3).To(BeNil())
		})
	})
})

func TestAttributesHolder(t *testing.T) {
	logger.SetLogLevel(logger.LogLevelNone)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Attributes Holder")
}
