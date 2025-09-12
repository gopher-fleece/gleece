package annotations_test

import (
	"testing"

	"github.com/gopher-fleece/gleece/core/annotations"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
	"github.com/gopher-fleece/gleece/test/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/titanous/json5"
)

var _ = Describe("Unit Tests - Annotation Holder", func() {
	Context("Given a single comment", func() {

		When("Attribute is simple", func() {

			It("Constructs without error", func() {
				comments := []string{"// @Description Abcd"}
				nodes := utils.CommentsToCommentBlock(comments, 1)
				_, err := annotations.NewAnnotationHolder(nodes, annotations.CommentSourceController)
				Expect(err).To(BeNil())
			})

			It("Correctly detects attribute exists", func() {
				comments := []string{"// @Description Abcd"}
				nodes := utils.CommentsToCommentBlock(comments, 1)
				holder, _ := annotations.NewAnnotationHolder(nodes, annotations.CommentSourceController)
				Expect(holder.Has(annotations.GleeceAnnotationDescription)).To(BeTrue())
			})

			It("Correctly gets the attribute", func() {
				comments := []string{"// @Description Abcd"}
				nodes := utils.CommentsToCommentBlock(comments, 1)
				holder, _ := annotations.NewAnnotationHolder(nodes, annotations.CommentSourceController)
				attrib := holder.GetFirst(annotations.GleeceAnnotationDescription)
				Expect(attrib).ToNot(BeNil())
			})

			It("Returns correct value from the GetDescription method", func() {
				comments := []string{"// @Description Abcd"}
				nodes := utils.CommentsToCommentBlock(comments, 1)
				holder, _ := annotations.NewAnnotationHolder(nodes, annotations.CommentSourceController)
				Expect(holder.GetDescription()).To(Equal("Abcd"))
			})

			It("GetFirstValueOrEmpty returns correct value when a single instance of the attribute exists", func() {
				comments := []string{"// @Method(POST)"}
				nodes := utils.CommentsToCommentBlock(comments, 1)
				holder, _ := annotations.NewAnnotationHolder(nodes, annotations.CommentSourceRoute)
				Expect(holder.GetFirstValueOrEmpty(annotations.GleeceAnnotationMethod)).To(Equal("POST"))
			})

			It("GetFirstValueOrEmpty returns empty string when attribute does not exist", func() {
				comments := []string{"// @Route(/route)"}
				nodes := utils.CommentsToCommentBlock(comments, 1)
				holder, _ := annotations.NewAnnotationHolder(nodes, annotations.CommentSourceRoute)
				Expect(holder.GetFirstValueOrEmpty(annotations.GleeceAnnotationMethod)).To(BeEmpty())
			})
		})

		When("Attribute is complex", func() {
			It("Constructs without error", func() {
				comments := []string{`// @Security(securitySchemaName, { scopes: ["read:users", "write:users"] }) Abcd`}
				nodes := utils.CommentsToCommentBlock(comments, 1)
				_, err := annotations.NewAnnotationHolder(nodes, annotations.CommentSourceRoute)
				Expect(err).To(BeNil())
			})

			It("Correctly detects attribute exists", func() {
				comments := []string{`// @Security(securitySchemaName, { scopes: ["read:users", "write:users"] }) Abcd`}
				nodes := utils.CommentsToCommentBlock(comments, 1)
				holder, _ := annotations.NewAnnotationHolder(nodes, annotations.CommentSourceRoute)
				Expect(holder.Has(annotations.GleeceAnnotationSecurity)).To(BeTrue())
			})

			It("Correctly gets the attribute", func() {
				comments := []string{`// @Security(securitySchemaName, { scopes: ["read:users", "write:users"] }) Abcd`}
				nodes := utils.CommentsToCommentBlock(comments, 1)
				holder, _ := annotations.NewAnnotationHolder(nodes, annotations.CommentSourceRoute)
				attrib := holder.GetFirst(annotations.GleeceAnnotationSecurity)
				Expect(attrib).ToNot(BeNil())
			})

			It("Attribute has correct basic values", func() {
				comments := []string{`// @Security(securitySchemaName, { scopes: ["read:users", "write:users"] }) Abcd`}
				nodes := utils.CommentsToCommentBlock(comments, 1)
				holder, _ := annotations.NewAnnotationHolder(nodes, annotations.CommentSourceRoute)
				attrib := holder.GetFirst(annotations.GleeceAnnotationSecurity)

				Expect(attrib.Name).To(Equal(annotations.GleeceAnnotationSecurity))
				Expect(attrib.Value).To(Equal("securitySchemaName"))
				Expect(attrib.Description).To(Equal("Abcd"))
			})

			It("Attribute has correct properties", func() {
				comments := []string{`// @Security(securitySchemaName, { scopes: ["read:users", "write:users"] }) Abcd`}
				nodes := utils.CommentsToCommentBlock(comments, 1)
				holder, _ := annotations.NewAnnotationHolder(nodes, annotations.CommentSourceRoute)
				attrib := holder.GetFirst(annotations.GleeceAnnotationSecurity)

				Expect(len(attrib.Properties)).To(Equal(1))
				Expect(attrib.HasProperty(annotations.PropertySecurityScopes)).To(BeTrue())
				value, err := annotations.GetCastProperty[[]string](attrib, annotations.PropertySecurityScopes)
				Expect(err).To(BeNil())
				Expect(*value).To(HaveExactElements("read:users", "write:users"))
			})

			It("Returns nil if property does not exist", func() {
				comments := []string{`// @Security(securitySchemaName, { scopes: ["read:users", "write:users"] }) Abcd`}
				nodes := utils.CommentsToCommentBlock(comments, 1)
				holder, _ := annotations.NewAnnotationHolder(nodes, annotations.CommentSourceRoute)
				attrib := holder.GetFirst(annotations.GleeceAnnotationSecurity)

				value, err := annotations.GetCastProperty[[]string](attrib, "DoesNotExist")
				Expect(err).To(BeNil())
				Expect(value).To(BeNil())
			})

			It("Returns an error if a slice property exists but cannot be cast to the requested non-slice type", func() {
				comments := []string{`// @Security(securitySchemaName, { scopes: ["read:users", "write:users"] }) Abcd`}
				nodes := utils.CommentsToCommentBlock(comments, 1)
				holder, _ := annotations.NewAnnotationHolder(nodes, annotations.CommentSourceRoute)
				attrib := holder.GetFirst(annotations.GleeceAnnotationSecurity)

				value, err := annotations.GetCastProperty[int](attrib, annotations.PropertySecurityScopes)
				Expect(err).To(MatchError(ContainSubstring("exists but cannot be cast")))
				Expect(value).To(BeNil())
			})

			It("Returns an error if a slice property exists but cannot be cast to the requested slice type", func() {
				comments := []string{`// @Security(securitySchemaName, { scopes: ["read:users", "write:users"] }) Abcd`}
				nodes := utils.CommentsToCommentBlock(comments, 1)
				holder, _ := annotations.NewAnnotationHolder(nodes, annotations.CommentSourceRoute)
				attrib := holder.GetFirst(annotations.GleeceAnnotationSecurity)

				value, err := annotations.GetCastProperty[[]int](attrib, annotations.PropertySecurityScopes)
				Expect(err).To(MatchError(ContainSubstring("cannot be converted to type")))
				Expect(value).To(BeNil())
			})

			It("Returns an error if attempting to convert a non-slice property to a slice", func() {
				comments := []string{`// @TemplateContext(securitySchemaName, { scopes: "V" }) Abcd`}
				nodes := utils.CommentsToCommentBlock(comments, 1)
				holder, _ := annotations.NewAnnotationHolder(nodes, annotations.CommentSourceRoute)
				attrib := holder.GetFirst(annotations.GleeceAnnotationTemplateContext)

				value, err := annotations.GetCastProperty[[]string](attrib, annotations.PropertySecurityScopes)
				Expect(err).To(MatchError(ContainSubstring("cannot be converted to type")))
				Expect(value).To(BeNil())
			})

			It("Returns an error an annotation's JSON5 part is malformed", func() {
				comments := []string{`// @Security(securitySchemaName, { scopes: ThisIsMalformed }) Abcd`}
				nodes := utils.CommentsToCommentBlock(comments, 1)
				_, err := annotations.NewAnnotationHolder(nodes, annotations.CommentSourceRoute)
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
			nodes := utils.CommentsToCommentBlock(stdComments, 1)
			_, err := annotations.NewAnnotationHolder(nodes, annotations.CommentSourceRoute)
			Expect(err).To(BeNil())
		})

		It("Correctly detects all attributes exist", func() {
			nodes := utils.CommentsToCommentBlock(stdComments, 1)
			holder, _ := annotations.NewAnnotationHolder(nodes, annotations.CommentSourceRoute)

			Expect(holder.Has(annotations.GleeceAnnotationDescription)).To(BeTrue())
			Expect(holder.Has(annotations.GleeceAnnotationMethod)).To(BeTrue())
			Expect(holder.Has(annotations.GleeceAnnotationRoute)).To(BeTrue())
			Expect(holder.Has(annotations.GleeceAnnotationQuery)).To(BeTrue())
			Expect(holder.Has(annotations.GleeceAnnotationPath)).To(BeTrue())
			Expect(holder.Has(annotations.GleeceAnnotationBody)).To(BeTrue())
			Expect(holder.Has(annotations.GleeceAnnotationHeader)).To(BeTrue())
			Expect(holder.Has(annotations.GleeceAnnotationResponse)).To(BeTrue())
			Expect(holder.Has(annotations.GleeceAnnotationErrorResponse)).To(BeTrue())
			Expect(holder.Has(annotations.GleeceAnnotationSecurity)).To(BeTrue())
			Expect(holder.Has(annotations.GleeceAnnotationTemplateContext)).To(BeTrue())
		})

		It("Correctly gets all attributes of the same type", func() {
			nodes := utils.CommentsToCommentBlock(stdComments, 1)
			holder, _ := annotations.NewAnnotationHolder(nodes, annotations.CommentSourceRoute)
			attributes := holder.GetAll(annotations.GleeceAnnotationPath)
			Expect(attributes).To(HaveLen(3))
		})

		It("Attributes of the same type are ordered and have correct values", func() {
			nodes := utils.CommentsToCommentBlock(stdComments, 1)
			holder, _ := annotations.NewAnnotationHolder(nodes, annotations.CommentSourceRoute)
			allAttributes := holder.GetAll(annotations.GleeceAnnotationPath)

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

			nodes := utils.CommentsToCommentBlock(comments, 1)
			holder, err := annotations.NewAnnotationHolder(nodes, annotations.CommentSourceRoute)
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

			nodes := utils.CommentsToCommentBlock(comments, 1)
			holder, err := annotations.NewAnnotationHolder(nodes, annotations.CommentSourceRoute)
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

			nodes := utils.CommentsToCommentBlock(comments, 1)
			holder, _ := annotations.NewAnnotationHolder(nodes, annotations.CommentSourceRoute)
			Expect(holder.GetFirstDescriptionOrEmpty(annotations.GleeceAnnotationRoute)).To(Equal("Some description1"))
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

			nodes := utils.CommentsToCommentBlock(comments, 1)
			holder, _ := annotations.NewAnnotationHolder(nodes, annotations.CommentSourceRoute)
			Expect(holder.GetFirstDescriptionOrEmpty(annotations.GleeceAnnotationRoute)).To(BeEmpty())
		})

		It("FindFirstByValue returns the first attribute who's value matches the search parameter", func() {
			comments := []string{
				`// @Security(schema1, { scopes: ["read:devices", "write:devices"] }) Match1`,
				`// @Security(schema2, { scopes: ["read:users", "write:users"] }) Match2`,
			}

			nodes := utils.CommentsToCommentBlock(comments, 1)
			holder, _ := annotations.NewAnnotationHolder(nodes, annotations.CommentSourceRoute)
			match := holder.FindFirstByValue("schema2")
			Expect(match).ToNot(BeNil())
			Expect(match.Description).To(Equal("Match2"))
		})

		It("FindFirstByValue returns nil if no value matches the search parameter", func() {
			comments := []string{
				`// @Security(schema1, { scopes: ["read:devices", "write:devices"] }) Match1`,
				`// @Security(schema2, { scopes: ["read:users", "write:users"] }) Match2`,
			}

			nodes := utils.CommentsToCommentBlock(comments, 1)
			holder, _ := annotations.NewAnnotationHolder(nodes, annotations.CommentSourceRoute)
			match := holder.FindFirstByValue("does not exists")
			Expect(match).To(BeNil())
		})

		It("FindFirstByProperty returns the first attribute that has a property who's name and value match the search parameters", func() {
			comments := []string{
				`// @TemplateContext(schema1, { scopes: ["read:devices", "write:devices"] }) Match1`,
				`// @TemplateContext(schema2, { scopes: ["read:users", "write:users"], extraProp: "" }) Match2`,
			}

			nodes := utils.CommentsToCommentBlock(comments, 1)
			holder, _ := annotations.NewAnnotationHolder(nodes, annotations.CommentSourceRoute)
			match := holder.FindFirstByProperty("extraProp", "")
			Expect(match).ToNot(BeNil())
			Expect(match.Description).To(Equal("Match2"))
		})

		It("FindFirstByProperty returns nil if no attribute has a property who's name and value match the search parameters", func() {
			comments := []string{
				`// @Security(schema1, { scopes: ["read:devices", "write:devices"] }) Match1`,
				`// @Security(schema2, { scopes: ["read:users", "write:users"], extraProp: "" }) Match2`,
			}

			nodes := utils.CommentsToCommentBlock(comments, 1)
			holder, _ := annotations.NewAnnotationHolder(nodes, annotations.CommentSourceRoute)
			match := holder.FindFirstByProperty("extraProp", "this value doesn't exist")
			Expect(match).To(BeNil())
		})

		It("Should trim valid attribute with extra space at the end", func() {
			comments := []string{
				`// @Method(GET)	`,
				`// @Route(/test-response-validation-ptr-2)      `,
			}

			nodes := utils.CommentsToCommentBlock(comments, 1)
			holder, _ := annotations.NewAnnotationHolder(nodes, annotations.CommentSourceRoute)
			methodAttr := holder.GetFirst(annotations.GleeceAnnotationMethod)
			routeAttr := holder.GetFirst(annotations.GleeceAnnotationRoute)
			Expect(methodAttr.Name).To(Equal(annotations.GleeceAnnotationMethod))
			Expect(routeAttr.Name).To(Equal(annotations.GleeceAnnotationRoute))
		})
	})

	Context("Given comments", func() {

		When("Annotation is wrong", func() {

			It("Unknown Annotation", func() {
				comments := []string{"// @UnknownAnnotation"}
				nodes := utils.CommentsToCommentBlock(comments, 1)
				_, err := annotations.NewAnnotationHolder(nodes, annotations.CommentSourceController)
				Expect(err).To(MatchError(ContainSubstring("unknown annotation @UnknownAnnotation")))
			})

			It("Wrong Source Annotation", func() {
				comments := []string{"// @Tag(the tag)"}
				nodes := utils.CommentsToCommentBlock(comments, 1)
				_, notErr := annotations.NewAnnotationHolder(nodes, annotations.CommentSourceController)
				Expect(notErr).To(BeNil())
				_, err := annotations.NewAnnotationHolder(nodes, annotations.CommentSourceRoute)
				Expect(err).To(MatchError(ContainSubstring("annotation @Tag is not valid in route context")))
			})

			It("Missing Annotation value", func() {
				comments := []string{"// @Tag"}
				nodes := utils.CommentsToCommentBlock(comments, 1)
				_, err := annotations.NewAnnotationHolder(nodes, annotations.CommentSourceController)
				Expect(err).To(MatchError(ContainSubstring("annotation @Tag requires a value")))
			})

			It("Wrong Annotation value", func() {
				comments := []string{"// @Method(INVALID)"}
				nodes := utils.CommentsToCommentBlock(comments, 1)
				_, err := annotations.NewAnnotationHolder(nodes, annotations.CommentSourceRoute)
				Expect(err).To(MatchError(ContainSubstring("invalid HTTP method: INVALID")))
			})

			It("Wrong Annotation value type", func() {
				comments := []string{"// @Response(INVALID)"}
				nodes := utils.CommentsToCommentBlock(comments, 1)
				_, err := annotations.NewAnnotationHolder(nodes, annotations.CommentSourceRoute)
				Expect(err).To(MatchError(ContainSubstring("invalid status code: INVALID")))
			})

			It("Wrong Annotation properties - no properties allowed", func() {
				comments := []string{"// @Method(POST, { invalid: \"properties\" })"}
				nodes := utils.CommentsToCommentBlock(comments, 1)
				_, err := annotations.NewAnnotationHolder(nodes, annotations.CommentSourceRoute)
				Expect(err).To(MatchError(ContainSubstring("annotation @Method does not support properties")))
			})

			It("Wrong Annotation properties - not allowed property", func() {
				comments := []string{"// @Query(value, { invalid: \"properties\" })"}
				nodes := utils.CommentsToCommentBlock(comments, 1)
				_, err := annotations.NewAnnotationHolder(nodes, annotations.CommentSourceRoute)
				Expect(err).To(MatchError(ContainSubstring("property invalid is not allowed for annotation @Query")))
			})

			It("Wrong Annotation properties - wrong property type ", func() {
				comments := []string{"// @Query(value, { name: 123 })"}
				nodes := utils.CommentsToCommentBlock(comments, 1)
				_, err := annotations.NewAnnotationHolder(nodes, annotations.CommentSourceRoute)
				Expect(err).To(MatchError(ContainSubstring("invalid property name for annotation @Query: property name should be a string")))
			})
		})

		When("Annotation combination", func() {

			It("Wrong Annotation combination - duplicate not allowed annotation type", func() {
				comments := []string{"// @Body(value1)", "// @Body(value2)"}
				nodes := utils.CommentsToCommentBlock(comments, 1)
				_, err := annotations.NewAnnotationHolder(nodes, annotations.CommentSourceRoute)
				Expect(err).To(MatchError(ContainSubstring("multiple instances of annotation @Body are not allowed")))
			})

			It("Wrong Annotation combination - two from different not allowed type", func() {
				comments := []string{"// @Body(value1)", "// @FormField(value2)"}
				nodes := utils.CommentsToCommentBlock(comments, 1)
				_, err := annotations.NewAnnotationHolder(nodes, annotations.CommentSourceRoute)
				Expect(err).To(MatchError(ContainSubstring("annotations @FormField and @Body cannot be used together")))
			})
		})

		When("Annotation values combination", func() {

			It("Wrong Annotation values combination - in the same annotation type", func() {
				comments := []string{"// @Query(the_value)", "// @Query(the_value)"}
				nodes := utils.CommentsToCommentBlock(comments, 1)
				_, err := annotations.NewAnnotationHolder(nodes, annotations.CommentSourceRoute)
				Expect(err).To(MatchError(ContainSubstring("duplicate value 'the_value' used in @Query and @Query annotations")))
			})

			It("Wrong Annotation combination - two from different not allowed type", func() {
				comments := []string{"// @Query(the_value)", "// @Header(the_value)"}
				nodes := utils.CommentsToCommentBlock(comments, 1)
				_, err := annotations.NewAnnotationHolder(nodes, annotations.CommentSourceRoute)
				Expect(err).To(MatchError(ContainSubstring("duplicate value 'the_value' used in @Query and @Header annotations")))
			})

			It("Valid Annotation combination - two and one not allowed", func() {
				comments := []string{"// @Query(the_value)", "// @Security(the_value)"}
				nodes := utils.CommentsToCommentBlock(comments, 1)
				_, err := annotations.NewAnnotationHolder(nodes, annotations.CommentSourceRoute)
				Expect(err).To(BeNil())
			})

			It("Valid Annotation combination - two and both allowed", func() {
				comments := []string{"// @Security(the_value)", "// @Security(the_value)"}
				nodes := utils.CommentsToCommentBlock(comments, 1)
				_, err := annotations.NewAnnotationHolder(nodes, annotations.CommentSourceRoute)
				Expect(err).To(BeNil())
			})
		})

	})
})

func TestUnitAnnotationHolder(t *testing.T) {
	logger.SetLogLevel(logger.LogLevelNone)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Unit Tests - Annotation Holder")
}
