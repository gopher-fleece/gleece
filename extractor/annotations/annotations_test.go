package annotations_test

import (
	"testing"

	"github.com/gopher-fleece/gleece/extractor/annotations"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Attributes Holder", func() {

	Context("Given a single comment", func() {

		When("Attribute is simple", func() {

			It("Constructs without error", func() {
				comments := []string{"// @Description Abcd"}
				_, err := annotations.NewAnnotationHolder(comments)
				Expect(err).To(BeNil())
			})

			It("Correctly detects attribute exists", func() {
				comments := []string{"// @Description Abcd"}
				holder, _ := annotations.NewAnnotationHolder(comments)
				Expect(holder.Has(annotations.AttributeDescription)).To(BeTrue())
			})

			It("Correctly gets the attribute", func() {
				comments := []string{"// @Description Abcd"}
				holder, _ := annotations.NewAnnotationHolder(comments)
				attrib := holder.GetFirst(annotations.AttributeDescription)
				Expect(attrib).ToNot(BeNil())
			})

			It("Returns correct value from the GetDescription method", func() {
				comments := []string{"// @Description Abcd"}
				holder, _ := annotations.NewAnnotationHolder(comments)
				Expect(holder.GetDescription()).To(Equal("Abcd"))
			})
		})

		When("Attribute is complex", func() {
			It("Constructs without error", func() {
				comments := []string{`// @Security(securitySchemaName, { scopes: ["read:users", "write:users"] }) Abcd`}
				_, err := annotations.NewAnnotationHolder(comments)
				Expect(err).To(BeNil())
			})

			It("Correctly detects attribute exists", func() {
				comments := []string{`// @Security(securitySchemaName, { scopes: ["read:users", "write:users"] }) Abcd`}
				holder, _ := annotations.NewAnnotationHolder(comments)
				Expect(holder.Has(annotations.AttributeSecurity)).To(BeTrue())
			})

			It("Correctly gets the attribute", func() {
				comments := []string{`// @Security(securitySchemaName, { scopes: ["read:users", "write:users"] }) Abcd`}
				holder, _ := annotations.NewAnnotationHolder(comments)
				attrib := holder.GetFirst(annotations.AttributeSecurity)
				Expect(attrib).ToNot(BeNil())
			})

			It("Attribute has correct basic values", func() {
				comments := []string{`// @Security(securitySchemaName, { scopes: ["read:users", "write:users"] }) Abcd`}
				holder, _ := annotations.NewAnnotationHolder(comments)
				attrib := holder.GetFirst(annotations.AttributeSecurity)

				Expect(attrib.Name).To(Equal(annotations.AttributeSecurity))
				Expect(attrib.Value).To(Equal("securitySchemaName"))
				Expect(attrib.Description).To(Equal("Abcd"))
			})

			It("Attribute has correct properties", func() {
				comments := []string{`// @Security(securitySchemaName, { scopes: ["read:users", "write:users"] }) Abcd`}
				holder, _ := annotations.NewAnnotationHolder(comments)
				attrib := holder.GetFirst(annotations.AttributeSecurity)

				Expect(len(attrib.Properties)).To(Equal(1))
				Expect(attrib.HasProperty(annotations.PropertySecurityScopes)).To(BeTrue())
				value, err := annotations.GetCastProperty[[]string](attrib, annotations.PropertySecurityScopes)
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
			_, err := annotations.NewAnnotationHolder(stdComments)
			Expect(err).To(BeNil())
		})

		It("Correctly detects all attributes exist", func() {
			holder, _ := annotations.NewAnnotationHolder(stdComments)

			Expect(holder.Has(annotations.AttributeDescription)).To(BeTrue())
			Expect(holder.Has(annotations.AttributeMethod)).To(BeTrue())
			Expect(holder.Has(annotations.AttributeRoute)).To(BeTrue())
			Expect(holder.Has(annotations.AttributeQuery)).To(BeTrue())
			Expect(holder.Has(annotations.AttributePath)).To(BeTrue())
			Expect(holder.Has(annotations.AttributeBody)).To(BeTrue())
			Expect(holder.Has(annotations.AttributeHeader)).To(BeTrue())
			Expect(holder.Has(annotations.AttributeResponse)).To(BeTrue())
			Expect(holder.Has(annotations.AttributeErrorResponse)).To(BeTrue())
			Expect(holder.Has(annotations.AttributeSecurity)).To(BeTrue())
		})

		It("Correctly gets all attributes of the same type", func() {
			holder, _ := annotations.NewAnnotationHolder(stdComments)
			attributes := holder.GetAll(annotations.AttributePath)
			Expect(attributes).To(HaveLen(3))
		})

		It("Attributes of the same type are ordered and have correct values", func() {
			holder, _ := annotations.NewAnnotationHolder(stdComments)
			allAttributes := holder.GetAll(annotations.AttributePath)

			Expect(allAttributes[0].Value).To(Equal("id"))
			Expect(allAttributes[1].Value).To(Equal("id2"))
			Expect(allAttributes[2].Value).To(Equal("name"))

			value1, err1 := annotations.GetCastProperty[string](allAttributes[0], annotations.PropertyName)
			value2, err2 := annotations.GetCastProperty[string](allAttributes[1], annotations.PropertyName)
			value3, err3 := annotations.GetCastProperty[string](allAttributes[2], annotations.PropertyName)

			Expect(err1).To(BeNil())
			Expect(err2).To(BeNil())
			Expect(err3).To(BeNil())

			Expect(*value1).To(Equal("user_id"))
			Expect(*value2).To(Equal("user_id_2"))
			Expect(*value3).To(Equal("user_name"))

			value1, err1 = annotations.GetCastProperty[string](allAttributes[0], annotations.PropertyValidatorString)
			value2, err2 = annotations.GetCastProperty[string](allAttributes[1], annotations.PropertyValidatorString)
			value3, err3 = annotations.GetCastProperty[string](allAttributes[2], annotations.PropertyValidatorString)

			Expect(err1).To(BeNil())
			Expect(err2).To(BeNil())
			Expect(err3).To(BeNil())

			Expect(*value1).To(Equal("gt=1"))
			Expect(*value2).To(Equal("gt=10"))
			Expect(value3).To(BeNil())
		})
	})
})

func TestAnnotationHolder(t *testing.T) {
	logger.SetLogLevel(logger.LogLevelNone)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Attributes Holder")
}
