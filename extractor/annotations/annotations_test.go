package annotations_test

import (
	"testing"

	"github.com/gopher-fleece/gleece/extractor/annotations"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/titanous/json5"
)

var _ = Describe("Annotation Holder", func() {

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

			It("GetFirstValueOrEmpty returns correct value when a single instance of the attribute exists", func() {
				comments := []string{"// @Method(POST)"}
				holder, _ := annotations.NewAnnotationHolder(comments)
				Expect(holder.GetFirstValueOrEmpty(annotations.AttributeMethod)).To(Equal("POST"))
			})

			It("GetFirstValueOrEmpty returns empty string when attribute does not exist", func() {
				comments := []string{"// @Route(/route)"}
				holder, _ := annotations.NewAnnotationHolder(comments)
				Expect(holder.GetFirstValueOrEmpty(annotations.AttributeMethod)).To(BeEmpty())
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

			It("Returns nil if property does not exist", func() {
				comments := []string{`// @Security(securitySchemaName, { scopes: ["read:users", "write:users"] }) Abcd`}
				holder, _ := annotations.NewAnnotationHolder(comments)
				attrib := holder.GetFirst(annotations.AttributeSecurity)

				value, err := annotations.GetCastProperty[[]string](attrib, "DoesNotExist")
				Expect(err).To(BeNil())
				Expect(value).To(BeNil())
			})

			It("Returns an error if a slice property exists but cannot be cast to the requested non-slice type", func() {
				comments := []string{`// @Security(securitySchemaName, { scopes: ["read:users", "write:users"] }) Abcd`}
				holder, _ := annotations.NewAnnotationHolder(comments)
				attrib := holder.GetFirst(annotations.AttributeSecurity)

				value, err := annotations.GetCastProperty[int](attrib, annotations.PropertySecurityScopes)
				Expect(err).To(MatchError(ContainSubstring("exists but cannot be cast")))
				Expect(value).To(BeNil())
			})

			It("Returns an error if a slice property exists but cannot be cast to the requested slice type", func() {
				comments := []string{`// @Security(securitySchemaName, { scopes: ["read:users", "write:users"] }) Abcd`}
				holder, _ := annotations.NewAnnotationHolder(comments)
				attrib := holder.GetFirst(annotations.AttributeSecurity)

				value, err := annotations.GetCastProperty[[]int](attrib, annotations.PropertySecurityScopes)
				Expect(err).To(MatchError(ContainSubstring("cannot be converted to type")))
				Expect(value).To(BeNil())
			})

			It("Returns an error if attempting to convert a non-slice property to a slice", func() {
				comments := []string{`// @Security(securitySchemaName, { scopes: "V" }) Abcd`}
				holder, _ := annotations.NewAnnotationHolder(comments)
				attrib := holder.GetFirst(annotations.AttributeSecurity)

				value, err := annotations.GetCastProperty[[]string](attrib, annotations.PropertySecurityScopes)
				Expect(err).To(MatchError(ContainSubstring("cannot be converted to type")))
				Expect(value).To(BeNil())
			})

			It("Returns an error an annotation's JSON5 part is malformed", func() {
				comments := []string{`// @Security(securitySchemaName, { scopes: ThisIsMalformed }) Abcd`}
				_, err := annotations.NewAnnotationHolder(comments)
				Expect(err).To(HaveOccurred())
				Expect(err).To(BeAssignableToTypeOf(&json5.SyntaxError{}))
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
			`// @TemplateContext(MODE, {mode: "100"})`,
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
			Expect(holder.Has(annotations.AttributeTemplateContext)).To(BeTrue())
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

		It("Considers non-attribute comments at the start as description, in the absence of a @Description attribute", func() {
			comments := []string{
				"// First comment",
				"// Second comment line",
				"// Third comment line",
				"//",
				`// @Method(POST) This text is not part of the OpenAPI spec`,
				`// @Route(/user/{user_name}/{user_id}/{user_id_2}) Same here`,
				"//",
				"//",
				`// @Query(email, { validate: "required,email" }) The user's email`,
				`// @Path(id, { name: "user_id", validate:"gt=1" }) The user's ID`,
				`// @Security(schema1, { scopes: ["read:users", "write:users"] })`,
			}

			holder, err := annotations.NewAnnotationHolder(comments)
			Expect(err).To(BeNil())

			description := holder.GetDescription()
			expectedStr := "First comment\nSecond comment line\nThird comment line"
			Expect(description).To(Equal(expectedStr))
		})

		It("Considers @Description a priority over comments at the start", func() {
			comments := []string{
				"// First comment",
				"// Second comment line",
				"// Third comment line",
				"//",
				`// @Method(POST) This text is not part of the OpenAPI spec`,
				`// @Route(/user/{user_name}/{user_id}/{user_id_2}) Same here`,
				`// @Query(email, { validate: "required,email" }) The user's email`,
				`// @Path(id, { name: "user_id", validate:"gt=1" }) The user's ID`,
				`// @Description Some description`,
				`// @Security(schema1, { scopes: ["read:users", "write:users"] })`,
			}

			holder, err := annotations.NewAnnotationHolder(comments)
			Expect(err).To(BeNil())

			description := holder.GetDescription()
			Expect(description).To(Equal("Some description"))
		})

		It("GetFirstDescriptionOrEmpty returns an attribute list's first item's Description if not empty", func() {
			comments := []string{
				"// First comment",
				"// Second comment line",
				"// Third comment line",
				"//",
				`// @Method(POST) This text is not part of the OpenAPI spec`,
				`// @Route(/user1) Some description1`,
				`// @Query(email, { validate: "required,email" }) The user's email`,
				`// @Path(id, { name: "user_id", validate:"gt=1" }) The user's ID`,
				`// @Route(/user2) Some description2`,
				`// @Route(/user3) Some description3`,
				`// @Route(/user4) Some description4`,
				`// @Security(schema1, { scopes: ["read:users", "write:users"] })`,
			}

			holder, _ := annotations.NewAnnotationHolder(comments)
			Expect(holder.GetFirstDescriptionOrEmpty(annotations.AttributeRoute)).To(Equal("Some description1"))
		})

		It("GetFirstDescriptionOrEmpty returns empty string if attribute is not present", func() {
			comments := []string{
				"// First comment",
				"// Second comment line",
				"// Third comment line",
				"//",
				`// @Method(POST) This text is not part of the OpenAPI spec`,
				`// @Path(id, { name: "user_id", validate:"gt=1" }) The user's ID`,
				`// @Security(schema1, { scopes: ["read:users", "write:users"] })`,
			}

			holder, _ := annotations.NewAnnotationHolder(comments)
			Expect(holder.GetFirstDescriptionOrEmpty(annotations.AttributeRoute)).To(BeEmpty())
		})

		It("FindFirstByValue returns the first attribute who's value matches the search parameter", func() {
			comments := []string{
				`// @Security(schema1, { scopes: ["read:devices", "write:devices"] }) Match1`,
				`// @Security(schema2, { scopes: ["read:users", "write:users"] }) Match2`,
			}

			holder, _ := annotations.NewAnnotationHolder(comments)
			match := holder.FindFirstByValue("schema2")
			Expect(match).ToNot(BeNil())
			Expect(match.Description).To(Equal("Match2"))
		})

		It("FindFirstByValue returns nil if no value matches the search parameter", func() {
			comments := []string{
				`// @Security(schema1, { scopes: ["read:devices", "write:devices"] }) Match1`,
				`// @Security(schema2, { scopes: ["read:users", "write:users"] }) Match2`,
			}

			holder, _ := annotations.NewAnnotationHolder(comments)
			match := holder.FindFirstByValue("does not exists")
			Expect(match).To(BeNil())
		})

		It("FindFirstByProperty returns the first attribute that has a property who's name and value match the search parameters", func() {
			comments := []string{
				`// @Security(schema1, { scopes: ["read:devices", "write:devices"] }) Match1`,
				`// @Security(schema2, { scopes: ["read:users", "write:users"], extraProp: "" }) Match2`,
			}

			holder, _ := annotations.NewAnnotationHolder(comments)
			match := holder.FindFirstByProperty("extraProp", "")
			Expect(match).ToNot(BeNil())
			Expect(match.Description).To(Equal("Match2"))
		})

		It("FindFirstByProperty returns nil if no attribute has a property who's name and value match the search parameters", func() {
			comments := []string{
				`// @Security(schema1, { scopes: ["read:devices", "write:devices"] }) Match1`,
				`// @Security(schema2, { scopes: ["read:users", "write:users"], extraProp: "" }) Match2`,
			}

			holder, _ := annotations.NewAnnotationHolder(comments)
			match := holder.FindFirstByProperty("extraProp", "this value doesn't exist")
			Expect(match).To(BeNil())
		})
	})
})

func TestAnnotationHolder(t *testing.T) {
	logger.SetLogLevel(logger.LogLevelNone)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Annotation Holder")
}
