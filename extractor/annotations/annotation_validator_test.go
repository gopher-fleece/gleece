package annotations_test

import (
	"github.com/gopher-fleece/gleece/extractor/annotations"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Annotation Validator", func() {

	Context("When validating basic annotations", func() {
		When("Using Tag annotation", func() {
			It("Should validate correctly in controller context", func() {
				attr := annotations.Attribute{
					Name:  "Tag",
					Value: "users",
				}

				err := annotations.IsValidAnnotation(attr, "controller")
				Expect(err).To(BeNil())
			})

			It("Should reject in incorrect context", func() {
				attr := annotations.Attribute{
					Name:  "Tag",
					Value: "users",
				}

				err := annotations.IsValidAnnotation(attr, "route")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("not valid in route context"))
			})

			It("Should reject with properties", func() {
				attr := annotations.Attribute{
					Name:  "Tag",
					Value: "users",
					Properties: map[string]any{
						"invalid": "property", // Tag doesn't accept properties
					},
				}

				err := annotations.IsValidAnnotation(attr, "controller")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("annotation @Tag does not support properties"))
			})
		})

		When("Using Method annotation", func() {
			It("Should validate correctly with valid HTTP method", func() {
				attr := annotations.Attribute{
					Name:  "Method",
					Value: "GET",
				}

				err := annotations.IsValidAnnotation(attr, "route")
				Expect(err).To(BeNil())
			})

			It("Should validate case-insensitive HTTP methods", func() {
				attr := annotations.Attribute{
					Name:  "Method",
					Value: "post",
				}

				err := annotations.IsValidAnnotation(attr, "route")
				Expect(err).To(BeNil())
			})

			It("Should reject invalid HTTP methods", func() {
				attr := annotations.Attribute{
					Name:  "Method",
					Value: "INVALID",
				}

				err := annotations.IsValidAnnotation(attr, "route")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("invalid HTTP method"))
			})

			It("Should reject missing value", func() {
				attr := annotations.Attribute{
					Name:  "Method",
					Value: "",
				}

				err := annotations.IsValidAnnotation(attr, "route")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("requires a value"))
			})
		})
	})

	Context("When validating annotations with properties", func() {
		When("Using Query annotation", func() {
			It("Should validate correctly with valid properties", func() {
				attr := annotations.Attribute{
					Name:  "Query",
					Value: "userId",
					Properties: map[string]any{
						"name":     "user_id",
						"validate": "required",
					},
				}

				err := annotations.IsValidAnnotation(attr, "route")
				Expect(err).To(BeNil())
			})

			It("Should reject with invalid property type", func() {
				attr := annotations.Attribute{
					Name:  "Query",
					Value: "userId",
					Properties: map[string]any{
						"name":     123, // Should be a string
						"validate": "required",
					},
				}

				err := annotations.IsValidAnnotation(attr, "route")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("property name should be a string"))
			})

			It("Should reject with unknown property", func() {
				attr := annotations.Attribute{
					Name:  "Query",
					Value: "userId",
					Properties: map[string]any{
						"unknown": "property",
					},
				}

				err := annotations.IsValidAnnotation(attr, "route")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("property unknown is not allowed"))
			})
		})

		When("Using Security annotation", func() {
			It("Should validate correctly with valid scopes", func() {
				attr := annotations.Attribute{
					Name:  "Security",
					Value: "oauth2",
					Properties: map[string]any{
						"scopes": []any{"read:users", "write:users"},
					},
				}

				err := annotations.IsValidAnnotation(attr, "route")
				Expect(err).To(BeNil())
			})

			It("Should reject with incorrect scopes type", func() {
				attr := annotations.Attribute{
					Name:  "Security",
					Value: "oauth2",
					Properties: map[string]any{
						"scopes": "read:users", // Should be an array
					},
				}

				err := annotations.IsValidAnnotation(attr, "route")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("property scopes should be an array"))
			})
		})
	})

	Context("When validating response annotations", func() {
		When("Using Response annotation", func() {
			It("Should validate correctly with valid status code", func() {
				attr := annotations.Attribute{
					Name:  "Response",
					Value: "200",
				}

				err := annotations.IsValidAnnotation(attr, "route")
				Expect(err).To(BeNil())
			})

			It("Should reject with empty status code", func() {
				attr := annotations.Attribute{
					Name:  "Response",
					Value: "",
				}

				err := annotations.IsValidAnnotation(attr, "route")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("annotation @Response requires a value"))
			})

			It("Should reject with invalid status code", func() {
				attr := annotations.Attribute{
					Name:  "Response",
					Value: "invalid",
				}

				err := annotations.IsValidAnnotation(attr, "route")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("invalid status code: invalid"))
			})
		})

		When("Using ErrorResponse annotation", func() {
			It("Should validate correctly with valid status code", func() {
				attr := annotations.Attribute{
					Name:  "ErrorResponse",
					Value: "404",
				}

				err := annotations.IsValidAnnotation(attr, "route")
				Expect(err).To(BeNil())
			})

			It("Should reject with empty status code", func() {
				attr := annotations.Attribute{
					Name:  "ErrorResponse",
					Value: "",
				}

				err := annotations.IsValidAnnotation(attr, "route")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("annotation @ErrorResponse requires a value"))
			})
		})
	})

	Context("When validating with descriptions", func() {
		It("Should accept annotations with descriptions", func() {
			attr := annotations.Attribute{
				Name:        "Route",
				Value:       "/users",
				Description: "Gets all users",
			}

			err := annotations.IsValidAnnotation(attr, "controller")
			Expect(err).To(BeNil())
		})

		It("Should validate Description annotation", func() {
			attr := annotations.Attribute{
				Name:        "Description",
				Description: "This is a detailed description",
			}

			err := annotations.IsValidAnnotation(attr, "controller")
			Expect(err).To(BeNil())
		})
	})

	Context("When validating annotations with multiple instances", func() {
		It("Should accept multiple Security annotations", func() {
			// Security allows multiple annotations
			attr1 := annotations.Attribute{
				Name:  "Security",
				Value: "oauth2",
				Properties: map[string]any{
					"scopes": []any{"read:users"},
				},
			}

			attr2 := annotations.Attribute{
				Name:  "Security",
				Value: "apiKey",
			}

			err1 := annotations.IsValidAnnotation(attr1, "route")
			err2 := annotations.IsValidAnnotation(attr2, "route")

			Expect(err1).To(BeNil())
			Expect(err2).To(BeNil())
		})
	})

	Context("When validating annotation with dynamic properties", func() {
		It("Should accept any property", func() {
			// Security allows multiple annotations
			attr1 := annotations.Attribute{
				Name:  "TemplateContext",
				Value: "dynamic1",
				Properties: map[string]any{
					"dynamic2": "dynamic3",
				},
			}

			err := annotations.IsValidAnnotation(attr1, "route")
			Expect(err).To(BeNil())
		})
	})

	Context("When validating unknown annotations", func() {
		It("Should reject unknown annotations", func() {
			attr := annotations.Attribute{
				Name:  "UnknownAnnotation",
				Value: "test",
			}

			err := annotations.IsValidAnnotation(attr, "controller")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unknown annotation"))
		})
	})

	Context("When validating annotation collections", func() {
		When("Checking for multiple instances of non-multiple annotations", func() {
			It("Should reject multiple instances of Method annotation", func() {
				// Method doesn't allow multiple annotations
				attrs := []annotations.Attribute{
					{
						Name:  "Method",
						Value: "GET",
					},
					{
						Name:  "Method",
						Value: "POST", // This should cause validation to fail
					},
				}

				err := annotations.IsValidAnnotationCollection(attrs, "route")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("multiple instances of annotation @Method are not allowed"))
			})

			It("Should allow multiple instances of Query annotation", func() {
				// Query allows multiple annotations
				attrs := []annotations.Attribute{
					{
						Name:  "Query",
						Value: "userId",
					},
					{
						Name:  "Query",
						Value: "page",
					},
				}

				err := annotations.IsValidAnnotationCollection(attrs, "route")
				Expect(err).To(BeNil())
			})
		})

		When("Checking for mutually exclusive annotations", func() {
			It("Should reject Body and FormField annotations together", func() {
				attrs := []annotations.Attribute{
					{
						Name:  "Body",
						Value: "User",
					},
					{
						Name:  "FormField",
						Value: "username",
					},
				}

				err := annotations.IsValidAnnotationCollection(attrs, "route")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("annotations @FormField and @Body cannot be used together"))
			})

			It("Should allow Body annotation with non-exclusive annotations", func() {
				attrs := []annotations.Attribute{
					{
						Name:  "Body",
						Value: "User",
					},
					{
						Name:  "Method",
						Value: "POST",
					},
				}

				err := annotations.IsValidAnnotationCollection(attrs, "route")
				Expect(err).To(BeNil())
			})

			It("Should reject FormField and Body even if other annotations exist in between", func() {
				attrs := []annotations.Attribute{
					{
						Name:  "FormField",
						Value: "username",
					},
					{
						Name:  "Method",
						Value: "POST",
					},
					{
						Name:  "Body",
						Value: "User",
					},
				}

				err := annotations.IsValidAnnotationCollection(attrs, "route")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("annotations @Body and @FormField cannot be used together"))
			})
		})

		When("Validating a complex collection of annotations", func() {
			It("Should validate all annotations in a valid collection", func() {
				attrs := []annotations.Attribute{
					{
						Name:  "Method",
						Value: "POST",
					},
					{
						Name:  "Route",
						Value: "/users",
					},
					{
						Name:  "Body",
						Value: "User",
						Properties: map[string]any{
							"validate": "required",
						},
					},
					{
						Name:  "Response",
						Value: "201",
					},
					{
						Name:  "ErrorResponse",
						Value: "400",
					},
					{
						Name:  "ErrorResponse",
						Value: "500",
					},
					{
						Name:  "Security",
						Value: "oauth2",
						Properties: map[string]any{
							"scopes": []any{"write:users"},
						},
					},
				}

				err := annotations.IsValidAnnotationCollection(attrs, "route")
				Expect(err).To(BeNil())
			})

			It("Should fail on invalid annotation in a collection", func() {
				attrs := []annotations.Attribute{
					{
						Name:  "Method",
						Value: "INVALID", // Invalid HTTP method
					},
					{
						Name:  "Route",
						Value: "/users",
					},
				}

				err := annotations.IsValidAnnotationCollection(attrs, "route")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("invalid HTTP method"))
			})

			It("Should fail with multiple non-multiple annotations in a collection", func() {
				attrs := []annotations.Attribute{
					{
						Name:  "Method",
						Value: "GET",
					},
					{
						Name:  "Route",
						Value: "/users",
					},
					{
						Name:  "Body",
						Value: "name1",
					},
					{
						Name:  "Body", // Body doesn't allow multiple annotations
						Value: "name2",
					},
				}

				err := annotations.IsValidAnnotationCollection(attrs, "route")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("multiple instances of annotation @Body are not allowed"))
			})
		})
	})

	Context("When validating unique values across annotation types", func() {
		It("Should reject duplicate values between Path and Query annotations", func() {
			attrs := []annotations.Attribute{
				{
					Name:  "Route",
					Value: "/users/{userId}",
				},
				{
					Name:  "Path",
					Value: "userId",
				},
				{
					Name:  "Query",
					Value: "userId", // Same value as Path - should fail
				},
			}

			err := annotations.IsValidAnnotationCollection(attrs, "route")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("duplicate value 'userId' used in @Path and @Query annotations"))
		})

		It("Should reject duplicate values between Header and FormField annotations", func() {
			attrs := []annotations.Attribute{
				{
					Name:  "Header",
					Value: "authToken",
				},
				{
					Name:  "FormField",
					Value: "authToken", // Same value as Header - should fail
				},
			}

			err := annotations.IsValidAnnotationCollection(attrs, "route")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("duplicate value 'authToken' used in @Header and @FormField annotations"))
		})

		It("Should reject duplicate values between Body and Path annotations", func() {
			attrs := []annotations.Attribute{
				{
					Name:  "Body",
					Value: "userData",
				},
				{
					Name:  "Path",
					Value: "userData", // Same value as Body - should fail
				},
			}

			err := annotations.IsValidAnnotationCollection(attrs, "route")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("duplicate value 'userData' used in @Body and @Path annotations"))
		})

		It("Should reject duplicate values between Query and FormField annotations", func() {
			attrs := []annotations.Attribute{
				{
					Name:  "Query",
					Value: "searchTerm",
				},
				{
					Name:  "FormField",
					Value: "searchTerm", // Same value as Query - should fail
				},
			}

			err := annotations.IsValidAnnotationCollection(attrs, "route")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("duplicate value 'searchTerm' used in @Query and @FormField annotations"))
		})

		It("Should reject duplicate values across multiple annotation types", func() {
			attrs := []annotations.Attribute{
				{
					Name:  "Route",
					Value: "/api/v1/users/{id}",
				},
				{
					Name:  "Method",
					Value: "POST",
				},
				{
					Name:  "Path",
					Value: "id",
				},
				{
					Name:  "Query",
					Value: "filter",
				},
				{
					Name:  "Header",
					Value: "id", // Duplicate of Path - should fail
				},
			}

			err := annotations.IsValidAnnotationCollection(attrs, "route")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("duplicate value 'id' used in @Path and @Header annotations"))
		})

		It("Should allow duplicate values for annotations that don't require uniqueness", func() {
			attrs := []annotations.Attribute{
				{
					Name:  "Route",
					Value: "/api/v1/users/{userId}",
				},
				{
					Name:  "Method",
					Value: "GET",
				},
				{
					Name:  "Response",
					Value: "200",
				},
				{
					Name:  "Security",
					Value: "200", // Same as Response, but doesn't require uniqueness
				},
				{
					Name:  "Path",
					Value: "userId", // This should be unique
				},
			}

			err := annotations.IsValidAnnotationCollection(attrs, "route")
			Expect(err).To(BeNil())
		})

		It("Should allow unique values for all annotations requiring uniqueness", func() {
			attrs := []annotations.Attribute{
				{
					Name:  "Route",
					Value: "/api/v1/users/{userId}",
				},
				{
					Name:  "Path",
					Value: "userId",
				},
				{
					Name:  "Query",
					Value: "pageSize",
				},
				{
					Name:  "Header",
					Value: "authorization",
				},
				{
					Name:  "Body",
					Value: "userPayload",
				},
			}

			err := annotations.IsValidAnnotationCollection(attrs, "route")
			Expect(err).To(BeNil())
		})

		It("Should handle empty values properly", func() {
			// Empty values should be ignored for uniqueness check
			attrs := []annotations.Attribute{
				{
					Name:  "Path",
					Value: "",
				},
				{
					Name:  "Query",
					Value: "",
				},
			}

			err := annotations.IsValidAnnotationCollection(attrs, "route")
			// Should fail for other reasons (empty values) before uniqueness check
			Expect(err).To(HaveOccurred())
			// But not due to duplicates
			Expect(err.Error()).NotTo(ContainSubstring("duplicate value"))
		})
	})

	Context("When validating Path annotations with Route URLs", func() {
		It("Should validate Path annotation that exists in Route URL", func() {
			attrs := []annotations.Attribute{
				{
					Name:  "Route",
					Value: "/users/{userId}",
				},
				{
					Name:  "Path",
					Value: "userId",
				},
			}

			err := annotations.IsValidAnnotationCollection(attrs, "route")
			Expect(err).To(BeNil())
		})

		It("Should reject Path annotation not found in Route URL", func() {
			attrs := []annotations.Attribute{
				{
					Name:  "Route",
					Value: "/users/list",
				},
				{
					Name:  "Path",
					Value: "userId",
				},
			}

			err := annotations.IsValidAnnotationCollection(attrs, "route")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("annotation @Path with name 'userId' is not found in the route URL"))
		})

		It("Should validate Path annotation using name property instead of value", func() {
			attrs := []annotations.Attribute{
				{
					Name:  "Route",
					Value: "/users/{id}",
				},
				{
					Name:  "Path",
					Value: "userId",
					Properties: map[string]any{
						"name": "id",
					},
				},
			}

			err := annotations.IsValidAnnotationCollection(attrs, "route")
			Expect(err).To(BeNil())
		})

		It("Should reject Path annotation with name property not found in Route URL", func() {
			attrs := []annotations.Attribute{
				{
					Name:  "Route",
					Value: "/users/{userId}",
				},
				{
					Name:  "Path",
					Value: "userId",
					Properties: map[string]any{
						"name": "id",
					},
				},
			}

			err := annotations.IsValidAnnotationCollection(attrs, "route")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("annotation @Path with name 'id' is not found in the route URL"))
		})

		It("Should validate Path annotation with complex Route URL", func() {
			attrs := []annotations.Attribute{
				{
					Name:  "Route",
					Value: "/users/{userId}/posts/{postId}",
				},
				{
					Name:  "Path",
					Value: "userId",
				},
				{
					Name:  "Path",
					Value: "postId",
				},
			}

			err := annotations.IsValidAnnotationCollection(attrs, "route")
			Expect(err).To(BeNil())
		})

		It("Should validate Path annotation in a collection with multiple annotations", func() {
			attrs := []annotations.Attribute{
				{
					Name:  "Method",
					Value: "GET",
				},
				{
					Name:  "Route",
					Value: "/api/v1/users/{userId}",
				},
				{
					Name:  "Path",
					Value: "userId",
				},
				{
					Name:  "Response",
					Value: "200",
				},
			}

			err := annotations.IsValidAnnotationCollection(attrs, "route")
			Expect(err).To(BeNil())
		})

		It("Should reject when no Route annotation is provided with Path annotation", func() {
			attrs := []annotations.Attribute{
				{
					Name:  "Method",
					Value: "GET",
				},
				{
					Name:  "Path",
					Value: "userId",
				},
			}

			err := annotations.IsValidAnnotationCollection(attrs, "route")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("annotation @Path with name 'userId' is not found in the route URL"))
		})
	})
})
